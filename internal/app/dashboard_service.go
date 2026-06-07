package app

import (
	"sync"
	"time"

	"github.com/varmiguemunoz/sprintos/internal/domain"
	"gorm.io/gorm"
)

type SprintVelocity struct {
	SprintName string
	Completed  int
	Total      int
}

type StateCount struct {
	StateName string
	Color     string
	Count     int
}

type MemberWorkload struct {
	UserName       string
	AssignedCount  int
	CompletedCount int
}

type DashboardMetrics struct {
	SprintCompletionRate float64
	AvgCycleTimeDays     float64
	WeeklyThroughput     int
	OnTimeDeliveryRate   float64
	OverdueCount         int
	OverdueTasks         []domain.Task
	VelocityTrend        []SprintVelocity
	StateDistribution    []StateCount
	TeamWorkload         []MemberWorkload
	RecentlyCompleted    []domain.Task
	HasSprints           bool
	HasAnyData           bool
	Projects             []domain.Project
	ActiveSprintName     string
}

type DashboardService struct {
	db *gorm.DB
}

func NewDashboardService(db *gorm.DB) *DashboardService {
	return &DashboardService{db: db}
}

func (s *DashboardService) GetMetrics(orgID uint, projectID *uint) (*DashboardMetrics, error) {
	metrics := &DashboardMetrics{}

	var projects []domain.Project
	s.db.Where("organization_id = ?", orgID).Find(&projects)
	metrics.Projects = projects

	var projectIDs []uint
	if projectID != nil {
		projectIDs = []uint{*projectID}
	} else {
		for _, p := range projects {
			projectIDs = append(projectIDs, p.ID)
		}
	}

	if len(projectIDs) == 0 {
		return metrics, nil
	}

	var taskCount int64
	s.db.Model(&domain.Task{}).
		Where("project_id IN ? AND deleted_at IS NULL", projectIDs).
		Count(&taskCount)
	metrics.HasAnyData = taskCount > 0

	var (
		weeklyThroughput   int
		overdueCount       int
		overdueTasks       []domain.Task
		avgCycleTimeDays   float64
		onTimeDeliveryRate float64
		stateDistribution  []StateCount
		teamWorkload       []MemberWorkload
		recentlyCompleted  []domain.Task
		velocities         []SprintVelocity
		hasSprints         bool
		sprintRate         float64
		activeSprintName   string
	)

	var wg sync.WaitGroup
	wg.Add(8)

	go func() {
		defer wg.Done()
		weeklyThroughput = s.weeklyThroughput(projectIDs)
	}()
	go func() {
		defer wg.Done()
		overdueCount, overdueTasks = s.overdueTasks(projectIDs)
	}()
	go func() {
		defer wg.Done()
		avgCycleTimeDays = s.avgCycleTimeDays(projectIDs)
	}()
	go func() {
		defer wg.Done()
		onTimeDeliveryRate = s.onTimeDeliveryRate(projectIDs)
	}()
	go func() {
		defer wg.Done()
		stateDistribution = s.stateDistribution(projectIDs)
	}()
	go func() {
		defer wg.Done()
		teamWorkload = s.teamWorkload(projectIDs)
	}()
	go func() {
		defer wg.Done()
		recentlyCompleted = s.recentlyCompleted(projectIDs)
	}()
	go func() {
		defer wg.Done()
		velocities, hasSprints, sprintRate, activeSprintName = s.sprintMetrics(projectIDs)
	}()

	wg.Wait()

	metrics.WeeklyThroughput = weeklyThroughput
	metrics.OverdueCount = overdueCount
	metrics.OverdueTasks = overdueTasks
	metrics.AvgCycleTimeDays = avgCycleTimeDays
	metrics.OnTimeDeliveryRate = onTimeDeliveryRate
	metrics.StateDistribution = stateDistribution
	metrics.TeamWorkload = teamWorkload
	metrics.RecentlyCompleted = recentlyCompleted
	metrics.VelocityTrend = velocities
	metrics.HasSprints = hasSprints
	metrics.SprintCompletionRate = sprintRate
	metrics.ActiveSprintName = activeSprintName

	return metrics, nil
}

func (s *DashboardService) weeklyThroughput(projectIDs []uint) int {
	var count int64
	weekAgo := time.Now().Add(-7 * 24 * time.Hour)
	s.db.Model(&domain.Task{}).
		Where("project_id IN ? AND completed_at >= ? AND deleted_at IS NULL", projectIDs, weekAgo).
		Count(&count)
	return int(count)
}

func (s *DashboardService) overdueTasks(projectIDs []uint) (int, []domain.Task) {
	var tasks []domain.Task
	now := time.Now()
	s.db.Preload("State").Preload("Assignee").
		Where("project_id IN ? AND due_date < ? AND completed_at IS NULL AND deleted_at IS NULL", projectIDs, now).
		Order("due_date asc").
		Limit(10).
		Find(&tasks)
	return len(tasks), tasks
}

func (s *DashboardService) avgCycleTimeDays(projectIDs []uint) float64 {
	type row struct {
		AvgSeconds float64
	}
	var r row
	s.db.Raw(`
		SELECT AVG(EXTRACT(EPOCH FROM (completed_at - created_at))) AS avg_seconds
		FROM tasks
		WHERE project_id IN ? AND completed_at IS NOT NULL AND deleted_at IS NULL
	`, projectIDs).Scan(&r)

	if r.AvgSeconds == 0 {
		return 0
	}
	return r.AvgSeconds / 86400.0
}

func (s *DashboardService) onTimeDeliveryRate(projectIDs []uint) float64 {
	var total, onTime int64

	s.db.Model(&domain.Task{}).
		Where("project_id IN ? AND due_date IS NOT NULL AND completed_at IS NOT NULL AND deleted_at IS NULL", projectIDs).
		Count(&total)

	if total == 0 {
		return 0
	}

	s.db.Model(&domain.Task{}).
		Where("project_id IN ? AND due_date IS NOT NULL AND completed_at IS NOT NULL AND completed_at <= due_date AND deleted_at IS NULL", projectIDs).
		Count(&onTime)

	return float64(onTime) / float64(total) * 100
}

func (s *DashboardService) stateDistribution(projectIDs []uint) []StateCount {
	type row struct {
		StateName string
		Color     string
		Count     int
	}
	var rows []row
	s.db.Raw(`
		SELECT st.name AS state_name, st.color, COUNT(t.id) AS count
		FROM tasks t
		JOIN states st ON st.id = t.state_id
		WHERE t.project_id IN ? AND t.deleted_at IS NULL
		GROUP BY st.id, st.name, st.color, st.position
		ORDER BY st.position ASC
	`, projectIDs).Scan(&rows)

	result := make([]StateCount, len(rows))
	for i, r := range rows {
		result[i] = StateCount{StateName: r.StateName, Color: r.Color, Count: r.Count}
	}
	return result
}

func (s *DashboardService) teamWorkload(projectIDs []uint) []MemberWorkload {
	type row struct {
		UserName       string
		AssignedCount  int
		CompletedCount int
	}
	var rows []row
	s.db.Raw(`
		SELECT u.name AS user_name,
		       COUNT(t.id) AS assigned_count,
		       COUNT(CASE WHEN t.completed_at IS NOT NULL THEN 1 END) AS completed_count
		FROM tasks t
		JOIN users u ON u.id = t.assigned_to
		WHERE t.project_id IN ? AND t.deleted_at IS NULL AND t.assigned_to IS NOT NULL
		GROUP BY u.id, u.name
		ORDER BY assigned_count DESC
	`, projectIDs).Scan(&rows)

	result := make([]MemberWorkload, len(rows))
	for i, r := range rows {
		result[i] = MemberWorkload{
			UserName:       r.UserName,
			AssignedCount:  r.AssignedCount,
			CompletedCount: r.CompletedCount,
		}
	}
	return result
}

func (s *DashboardService) recentlyCompleted(projectIDs []uint) []domain.Task {
	var tasks []domain.Task
	s.db.Preload("Assignee").
		Where("project_id IN ? AND completed_at IS NOT NULL AND deleted_at IS NULL", projectIDs).
		Order("completed_at desc").
		Limit(8).
		Find(&tasks)
	return tasks
}

func (s *DashboardService) sprintMetrics(projectIDs []uint) ([]SprintVelocity, bool, float64, string) {
	var sprints []domain.Sprint
	s.db.Where("project_id IN ?", projectIDs).
		Order("start_date desc").
		Limit(4).
		Find(&sprints)

	if len(sprints) == 0 {
		return nil, false, 0, ""
	}

	sprintIDs := make([]uint, len(sprints))
	for i, sp := range sprints {
		sprintIDs[i] = sp.ID
	}

	type sprintCount struct {
		SprintID  uint
		Total     int
		Completed int
	}
	var counts []sprintCount
	s.db.Raw(`
		SELECT sprint_id,
		       COUNT(*) AS total,
		       COUNT(completed_at) AS completed
		FROM tasks
		WHERE sprint_id IN ?
		  AND deleted_at IS NULL
		GROUP BY sprint_id
	`, sprintIDs).Scan(&counts)

	countMap := make(map[uint]sprintCount, len(counts))
	for _, c := range counts {
		countMap[c.SprintID] = c
	}

	now := time.Now()
	var velocities []SprintVelocity
	var sprintRate float64
	var activeSprintName string

	for i := len(sprints) - 1; i >= 0; i-- {
		sp := sprints[i]
		c := countMap[sp.ID]

		velocities = append(velocities, SprintVelocity{
			SprintName: sp.Name,
			Completed:  c.Completed,
			Total:      c.Total,
		})

		if sp.StartDate.Before(now) && sp.EndDate.After(now) && sp.CompletedAt == nil {
			activeSprintName = sp.Name
			if c.Total > 0 {
				sprintRate = float64(c.Completed) / float64(c.Total) * 100
			}
		}
	}

	return velocities, true, sprintRate, activeSprintName
}
