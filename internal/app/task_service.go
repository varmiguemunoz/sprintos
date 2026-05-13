package app

import (
	"errors"
	"fmt"
	"time"

	"github.com/varmiguemunoz/sprintos/internal/domain"
	"gorm.io/gorm"
)

type TaskService struct {
	db *gorm.DB
}

func NewTaskService(db *gorm.DB) *TaskService {
	return &TaskService{db: db}
}

func (s *TaskService) Create(
	title string,
	description string,
	stateID uint,
	projectID uint,
	createdByID uint,
	assignedTo *uint,
	startDate *time.Time,
	dueDate *time.Time,
) (*domain.Task, error) {
	task := domain.Task{
		Title:       title,
		Description: &description,
		StateID:     stateID,
		ProjectID:   projectID,
		CreatedByID: createdByID,
		AssignedTo:  assignedTo,
		StartDate:   startDate,
		DueDate:     dueDate,
	}

	if err := s.db.Create(&task).Error; err != nil {
		return nil, fmt.Errorf("could not create task: %w", err)
	}

	return &task, nil
}

func (s *TaskService) Update(
	id uint,
	title string,
	description string,
	assignedTo *uint,
	startDate *time.Time,
	dueDate *time.Time,
) (*domain.Task, error) {
	var task domain.Task

	if err := s.db.First(&task, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("task with id %d not found", id)
		}
		return nil, fmt.Errorf("error fetching task: %w", err)
	}

	updates := map[string]interface{}{
		"title":       title,
		"description": description,
		"assigned_to": assignedTo,
		"start_date":  startDate,
		"due_date":    dueDate,
	}

	if err := s.db.Model(&task).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("could not update task: %w", err)
	}

	return &task, nil
}

func (s *TaskService) Delete(id uint) error {
	var task domain.Task

	if err := s.db.First(&task, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("task with id %d not found", id)
		}
		return fmt.Errorf("error fetching task: %w", err)
	}

	if err := s.db.Delete(&task).Error; err != nil {
		return fmt.Errorf("could not delete task: %w", err)
	}

	return nil
}

func (s *TaskService) GetByID(id uint) (*domain.Task, error) {
	var task domain.Task

	err := s.db.
		Preload("State").
		Preload("Assignee").
		Preload("CreatedBy").
		First(&task, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("task with id %d not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("error fetching task: %w", err)
	}

	return &task, nil
}

func (s *TaskService) ListByProject(projectID uint) ([]domain.Task, error) {
	var tasks []domain.Task

	if err := s.db.
		Preload("State").
		Preload("Assignee").
		Where("project_id = ?", projectID).
		Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("could not list tasks: %w", err)
	}

	return tasks, nil
}

func (s *TaskService) ListByState(stateID uint) ([]domain.Task, error) {
	var tasks []domain.Task

	if err := s.db.Where("state_id = ?", stateID).Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("could not list tasks by state: %w", err)
	}

	return tasks, nil
}

func (s *TaskService) MoveState(taskID uint, newStateID uint) (*domain.Task, error) {
	var task domain.Task
	var newState domain.State

	if err := s.db.First(&task, taskID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("task with id %d not found", taskID)
		}
		return nil, fmt.Errorf("error fetching task: %w", err)
	}

	if err := s.db.First(&newState, newStateID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("state with id %d not found", newStateID)
		}
		return nil, fmt.Errorf("error fetching state: %w", err)
	}

	updates := map[string]interface{}{
		"state_id":     newStateID,
		"completed_at": nil,
	}

	if newState.IsDone {
		now := time.Now()
		updates["completed_at"] = now
	}

	if err := s.db.Model(&task).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("could not move task: %w", err)
	}

	task.State = newState

	return &task, nil
}

func (s *TaskService) AssignUser(taskID uint, userID *uint) (*domain.Task, error) {
	var task domain.Task

	if err := s.db.First(&task, taskID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("task with id %d not found", taskID)
		}
		return nil, fmt.Errorf("error fetching task: %w", err)
	}

	if err := s.db.Model(&task).Update("assigned_to", userID).Error; err != nil {
		return nil, fmt.Errorf("could not assign user: %w", err)
	}

	return &task, nil
}
