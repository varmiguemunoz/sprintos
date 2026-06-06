package app

import (
	"fmt"
	"math"
	"time"

	"github.com/varmiguemunoz/sprintos/internal/domain"
	"gorm.io/gorm"
)

type ReportService struct {
	db *gorm.DB
}

func NewReportService(db *gorm.DB) *ReportService {
	return &ReportService{db: db}
}

type ReportParams struct {
	OrgID      uint
	ProjectIDs []uint
	From       time.Time
	To         time.Time
}

type ProjectReportData struct {
	ProjectName      string
	TotalTasks       int
	CompletedTasks   int
	InProgressTasks  int
	BacklogTasks     int
	OverdueTasks     int
	TotalHours       float64
	AvgCycleTimeDays float64
	OnTimeRate       float64
}

type MemberReportData struct {
	Name           string
	TasksCompleted int
	HoursLogged    float64
}

type WeeklyVelocityPoint struct {
	WeekLabel string
	Completed int
}

type RiskTaskData struct {
	Title       string
	ProjectName string
	Priority    string
	AgeDays     int
	DueDate     *time.Time
}

type ReportData struct {
	OrgName          string
	From             time.Time
	To               time.Time
	TotalCreated     int
	TotalCompleted   int
	TotalHours       float64
	OnTimeRate       float64
	AvgCycleTimeDays float64
	Projects         []ProjectReportData
	Members          []MemberReportData
	WeeklyVelocity   []WeeklyVelocityPoint
	OverdueTasks     []RiskTaskData
	CriticalHighOpen []RiskTaskData
}

func (s *ReportService) Generate(params ReportParams) (*ReportData, error) {
	projectIDs, err := s.resolveProjectIDs(params)
	if err != nil {
		return nil, err
	}
	if len(projectIDs) == 0 {
		return &ReportData{From: params.From, To: params.To}, nil
	}

	var org domain.Organization
	orgName := ""
	if err := s.db.First(&org, params.OrgID).Error; err == nil {
		orgName = org.Name
	}

	var projects []domain.Project
	s.db.Where("id IN ?", projectIDs).Find(&projects)
	projectMap := make(map[uint]string, len(projects))
	for _, p := range projects {
		projectMap[p.ID] = p.Name
	}

	var allTasks []domain.Task
	s.db.Preload("State").Where("project_id IN ? AND deleted_at IS NULL", projectIDs).Find(&allTasks)

	toEndOfDay := params.To.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	var createdInPeriod []domain.Task
	s.db.Preload("State").
		Where("project_id IN ? AND created_at BETWEEN ? AND ? AND deleted_at IS NULL", projectIDs, params.From, toEndOfDay).
		Find(&createdInPeriod)

	var completedInPeriod []domain.Task
	s.db.Preload("State").
		Where("project_id IN ? AND completed_at BETWEEN ? AND ? AND deleted_at IS NULL", projectIDs, params.From, toEndOfDay).
		Find(&completedInPeriod)

	taskIDs := make([]uint, 0, len(allTasks))
	taskProjectMap := make(map[uint]uint, len(allTasks))
	for _, t := range allTasks {
		taskIDs = append(taskIDs, t.ID)
		taskProjectMap[t.ID] = t.ProjectID
	}

	var timeEntries []domain.TimeEntry
	if len(taskIDs) > 0 {
		s.db.Where("task_id IN ? AND logged_at BETWEEN ? AND ? AND deleted_at IS NULL", taskIDs, params.From, toEndOfDay).
			Find(&timeEntries)
	}

	onTimeCount, withDueDate := 0, 0
	cycleSum := 0.0
	for _, t := range completedInPeriod {
		if t.CompletedAt != nil {
			cycleSum += t.CompletedAt.Sub(t.CreatedAt).Hours() / 24.0
		}
		if t.DueDate != nil {
			withDueDate++
			if t.CompletedAt != nil && !t.CompletedAt.After(*t.DueDate) {
				onTimeCount++
			}
		}
	}

	onTimeRate := 0.0
	if withDueDate > 0 {
		onTimeRate = round1(float64(onTimeCount) / float64(withDueDate) * 100.0)
	}
	avgCycle := 0.0
	if len(completedInPeriod) > 0 {
		avgCycle = round1(cycleSum / float64(len(completedInPeriod)))
	}

	totalHours := 0.0
	hoursByProject := make(map[uint]float64)
	for _, te := range timeEntries {
		if te.TaskID == nil {
			continue
		}
		h := float64(te.Minutes) / 60.0
		totalHours += h
		pid := taskProjectMap[*te.TaskID]
		hoursByProject[pid] += h
	}

	projectsData := s.buildProjectData(projectIDs, projectMap, allTasks, completedInPeriod, hoursByProject)

	var members []domain.TeamMember
	s.db.Preload("User").Where("organization_id = ? AND deleted_at IS NULL", params.OrgID).Find(&members)
	membersData := s.buildMemberData(members, completedInPeriod, params.From, toEndOfDay)

	weeklyVelocity := buildWeeklyVelocity(completedInPeriod, params.From, params.To)

	var overdueTasks []RiskTaskData
	for _, t := range allTasks {
		if !t.State.IsDone && t.DueDate != nil && t.DueDate.Before(time.Now()) {
			due := *t.DueDate
			overdueTasks = append(overdueTasks, RiskTaskData{
				Title:       t.Title,
				ProjectName: projectMap[t.ProjectID],
				Priority:    t.Priority,
				AgeDays:     int(time.Since(due).Hours() / 24),
				DueDate:     &due,
			})
		}
	}

	var criticalOpen []RiskTaskData
	for _, t := range allTasks {
		if !t.State.IsDone && (t.Priority == "critical" || t.Priority == "high") {
			criticalOpen = append(criticalOpen, RiskTaskData{
				Title:       t.Title,
				ProjectName: projectMap[t.ProjectID],
				Priority:    t.Priority,
				AgeDays:     int(time.Since(t.CreatedAt).Hours() / 24),
			})
		}
	}

	return &ReportData{
		OrgName:          orgName,
		From:             params.From,
		To:               params.To,
		TotalCreated:     len(createdInPeriod),
		TotalCompleted:   len(completedInPeriod),
		TotalHours:       round1(totalHours),
		OnTimeRate:       onTimeRate,
		AvgCycleTimeDays: avgCycle,
		Projects:         projectsData,
		Members:          membersData,
		WeeklyVelocity:   weeklyVelocity,
		OverdueTasks:     overdueTasks,
		CriticalHighOpen: criticalOpen,
	}, nil
}

func (s *ReportService) resolveProjectIDs(params ReportParams) ([]uint, error) {
	if len(params.ProjectIDs) > 0 {
		return params.ProjectIDs, nil
	}
	var projects []domain.Project
	if err := s.db.Where("organization_id = ? AND deleted_at IS NULL", params.OrgID).Find(&projects).Error; err != nil {
		return nil, fmt.Errorf("could not list projects: %w", err)
	}
	ids := make([]uint, len(projects))
	for i, p := range projects {
		ids[i] = p.ID
	}
	return ids, nil
}

func (s *ReportService) buildProjectData(
	projectIDs []uint,
	projectMap map[uint]string,
	allTasks []domain.Task,
	completedInPeriod []domain.Task,
	hoursByProject map[uint]float64,
) []ProjectReportData {
	result := make([]ProjectReportData, 0, len(projectIDs))
	for _, pid := range projectIDs {
		pd := ProjectReportData{
			ProjectName: projectMap[pid],
			TotalHours:  round1(hoursByProject[pid]),
		}

		for _, t := range allTasks {
			if t.ProjectID != pid {
				continue
			}
			pd.TotalTasks++
			if t.State.IsDone {
				pd.CompletedTasks++
			} else if t.DueDate != nil && t.DueDate.Before(time.Now()) {
				pd.OverdueTasks++
			} else if t.State.Name == "In Progress" {
				pd.InProgressTasks++
			} else {
				pd.BacklogTasks++
			}
		}

		pWithDue, pOnTime := 0, 0
		cycleSum := 0.0
		count := 0
		for _, t := range completedInPeriod {
			if t.ProjectID != pid {
				continue
			}
			count++
			if t.CompletedAt != nil {
				cycleSum += t.CompletedAt.Sub(t.CreatedAt).Hours() / 24.0
			}
			if t.DueDate != nil {
				pWithDue++
				if t.CompletedAt != nil && !t.CompletedAt.After(*t.DueDate) {
					pOnTime++
				}
			}
		}
		if count > 0 {
			pd.AvgCycleTimeDays = round1(cycleSum / float64(count))
		}
		if pWithDue > 0 {
			pd.OnTimeRate = round1(float64(pOnTime) / float64(pWithDue) * 100.0)
		}

		result = append(result, pd)
	}
	return result
}

func (s *ReportService) buildMemberData(
	members []domain.TeamMember,
	completedInPeriod []domain.Task,
	from, to time.Time,
) []MemberReportData {
	result := make([]MemberReportData, 0, len(members))
	for _, m := range members {
		md := MemberReportData{Name: m.User.Name}
		for _, t := range completedInPeriod {
			if t.AssignedTo != nil && *t.AssignedTo == m.UserID {
				md.TasksCompleted++
			}
		}
		var mins float64
		s.db.Raw(
			"SELECT COALESCE(SUM(minutes), 0) FROM time_entries WHERE user_id = ? AND logged_at BETWEEN ? AND ? AND deleted_at IS NULL",
			m.UserID, from, to,
		).Scan(&mins)
		md.HoursLogged = round1(mins / 60.0)
		result = append(result, md)
	}
	return result
}

func buildWeeklyVelocity(completedTasks []domain.Task, from, to time.Time) []WeeklyVelocityPoint {
	type weekKey struct {
		year int
		week int
	}

	counts := make(map[weekKey]int)
	var order []weekKey
	seen := make(map[weekKey]bool)

	cur := from
	for !cur.After(to) {
		y, w := cur.ISOWeek()
		k := weekKey{y, w}
		if !seen[k] {
			seen[k] = true
			order = append(order, k)
			counts[k] = 0
		}
		cur = cur.Add(24 * time.Hour)
	}

	for _, t := range completedTasks {
		if t.CompletedAt == nil {
			continue
		}
		y, w := t.CompletedAt.ISOWeek()
		k := weekKey{y, w}
		counts[k]++
	}

	result := make([]WeeklyVelocityPoint, 0, len(order))
	for _, k := range order {
		result = append(result, WeeklyVelocityPoint{
			WeekLabel: fmt.Sprintf("W%d", k.week),
			Completed: counts[k],
		})
	}
	return result
}

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}
