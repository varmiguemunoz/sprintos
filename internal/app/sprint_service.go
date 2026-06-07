package app

import (
	"errors"
	"fmt"
	"time"

	"github.com/varmiguemunoz/sprintos/internal/domain"
	"gorm.io/gorm"
)

type SprintService struct {
	db *gorm.DB
}

func NewSprintService(db *gorm.DB) *SprintService {
	return &SprintService{db: db}
}

func (s *SprintService) Create(name, goal string, projectID uint, start, end time.Time) (*domain.Sprint, error) {
	g := &goal
	if goal == "" {
		g = nil
	}
	sprint := domain.Sprint{
		Name:      name,
		Goal:      g,
		ProjectID: projectID,
		StartDate: start,
		EndDate:   end,
	}
	if err := s.db.Create(&sprint).Error; err != nil {
		return nil, fmt.Errorf("could not create sprint: %w", err)
	}
	return &sprint, nil
}

func (s *SprintService) ListByProject(projectID uint) ([]domain.Sprint, error) {
	var sprints []domain.Sprint
	if err := s.db.Where("project_id = ?", projectID).Order("start_date desc").Find(&sprints).Error; err != nil {
		return nil, fmt.Errorf("could not list sprints: %w", err)
	}
	return sprints, nil
}

func (s *SprintService) GetByID(id uint) (*domain.Sprint, error) {
	var sprint domain.Sprint
	err := s.db.First(&sprint, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("sprint %d not found", id)
	}
	return &sprint, err
}

func (s *SprintService) GetActive(projectID uint) (*domain.Sprint, error) {
	var sprint domain.Sprint
	now := time.Now()
	err := s.db.Where("project_id = ? AND start_date <= ? AND end_date >= ? AND completed_at IS NULL", projectID, now, now).
		First(&sprint).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("no active sprint for project %d", projectID)
	}
	return &sprint, err
}

func (s *SprintService) Update(id uint, name, goal string, start, end time.Time) (*domain.Sprint, error) {
	sprint, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	var g *string
	if goal != "" {
		g = &goal
	}

	sprint.Name = name
	sprint.Goal = g
	sprint.StartDate = start
	sprint.EndDate = end

	if err := s.db.Select("name", "goal", "start_date", "end_date").Save(sprint).Error; err != nil {
		return nil, fmt.Errorf("could not update sprint: %w", err)
	}

	return sprint, nil
}

func (s *SprintService) Delete(id uint) error {
	var sprint domain.Sprint
	if err := s.db.First(&sprint, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("sprint %d not found", id)
		}
		return fmt.Errorf("error fetching sprint: %w", err)
	}
	if err := s.db.Delete(&sprint).Error; err != nil {
		return fmt.Errorf("could not delete sprint: %w", err)
	}
	return nil
}

func (s *SprintService) AddTask(sprintID, taskID uint) error {
	return s.db.Model(&domain.Task{}).Where("id = ?", taskID).Update("sprint_id", sprintID).Error
}

func (s *SprintService) RemoveTask(taskID uint) error {
	return s.db.Model(&domain.Task{}).Where("id = ?", taskID).Update("sprint_id", nil).Error
}

func (s *SprintService) ListTasks(sprintID uint) ([]domain.Task, error) {
	var tasks []domain.Task
	err := s.db.Preload("State").Where("sprint_id = ?", sprintID).Find(&tasks).Error
	return tasks, err
}

func (s *SprintService) Complete(sprintID uint, backlogStateID uint) error {
	sprint, err := s.GetByID(sprintID)
	if err != nil {
		return err
	}
	if sprint.CompletedAt != nil {
		return fmt.Errorf("sprint is already completed")
	}

	var unfinished []domain.Task
	s.db.Joins("JOIN states ON states.id = tasks.state_id").
		Where("tasks.sprint_id = ? AND states.is_done = false AND tasks.deleted_at IS NULL", sprintID).
		Find(&unfinished)

	for _, t := range unfinished {
		s.db.Model(&t).Updates(map[string]interface{}{
			"sprint_id": nil,
			"state_id":  backlogStateID,
		})
	}

	now := time.Now()
	return s.db.Model(sprint).Update("completed_at", now).Error
}

func (s *SprintService) Velocity(sprintID uint) (completed, total int, err error) {
	var tasks []domain.Task
	if err = s.db.Joins("JOIN states ON states.id = tasks.state_id").
		Where("tasks.sprint_id = ? AND tasks.deleted_at IS NULL", sprintID).
		Find(&tasks).Error; err != nil {
		return
	}
	total = len(tasks)
	for _, t := range tasks {
		if t.CompletedAt != nil {
			completed++
		}
	}
	return
}

func (s *SprintService) TakeBurndownSnapshot(sprintID uint) error {
	var remaining, completed int64
	s.db.Model(&domain.Task{}).
		Joins("JOIN states ON states.id = tasks.state_id").
		Where("tasks.sprint_id = ? AND states.is_done = false AND tasks.deleted_at IS NULL", sprintID).
		Count(&remaining)
	s.db.Model(&domain.Task{}).
		Where("sprint_id = ? AND completed_at IS NOT NULL AND deleted_at IS NULL", sprintID).
		Count(&completed)

	snapshot := domain.BurndownSnapshot{
		SprintID:       sprintID,
		Date:           time.Now().Truncate(24 * time.Hour),
		RemainingTasks: int(remaining),
		CompletedTasks: int(completed),
	}
	return s.db.Create(&snapshot).Error
}

func (s *SprintService) GetBurndown(sprintID uint) ([]domain.BurndownSnapshot, error) {
	var snapshots []domain.BurndownSnapshot
	err := s.db.Where("sprint_id = ?", sprintID).Order("date asc").Find(&snapshots).Error
	return snapshots, err
}
