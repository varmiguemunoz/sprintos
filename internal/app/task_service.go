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

func (s *TaskService) nextTaskNumber(projectID uint) (int, error) {
	var orgID uint
	s.db.Raw("SELECT organization_id FROM projects WHERE id = ?", projectID).Scan(&orgID)

	var max int
	err := s.db.Raw("SELECT COALESCE(MAX(t.task_number), 0) FROM tasks t JOIN projects p ON p.id = t.project_id WHERE p.organization_id = ? AND t.deleted_at IS NULL", orgID).Scan(&max).Error

	if err != nil {
		return 0, fmt.Errorf("could not get next task number: %w", err)
	}
	return max + 1, nil
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
	taskNumber, err := s.nextTaskNumber(projectID)
	if err != nil {
		taskNumber = 1
	}

	task := domain.Task{
		TaskNumber:  taskNumber,
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

	oldStateID := task.StateID
	secondsInState := int64(time.Since(task.UpdatedAt).Seconds())

	if err := s.db.Model(&task).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("could not move task: %w", err)
	}

	transition := domain.StateTransition{
		TaskID:             taskID,
		FromStateID:        &oldStateID,
		ToStateID:          newStateID,
		ChangedByID:        task.CreatedByID,
		SecondsInFromState: secondsInState,
	}
	s.db.Create(&transition)

	task.State = newState

	return &task, nil
}

func (s *TaskService) GetByTaskNumber(taskNumber int, projectID uint) (*domain.Task, error) {
	var task domain.Task
	err := s.db.
		Preload("State").
		Preload("Assignee").
		Where("task_number = ? AND project_id = ?", taskNumber, projectID).
		First(&task).Error
	if err != nil {
		return nil, fmt.Errorf("task %d not found in project %d: %w", taskNumber, projectID, err)
	}
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
