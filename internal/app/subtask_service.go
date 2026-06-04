package app

import (
	"errors"
	"fmt"

	"github.com/varmiguemunoz/sprintos/internal/domain"
	"gorm.io/gorm"
)

type SubtaskService struct {
	db *gorm.DB
}

func NewSubtaskService(db *gorm.DB) *SubtaskService {
	return &SubtaskService{db: db}
}

func (s *SubtaskService) Create(title string, description string, taskID uint, createdByID uint) (*domain.Subtask, error) {
	desc := description
	subtask := domain.Subtask{
		Title:       title,
		Description: &desc,
		TaskID:      taskID,
		CreatedByID: createdByID,
	}

	if err := s.db.Create(&subtask).Error; err != nil {
		return nil, fmt.Errorf("could not create subtask: %w", err)
	}

	return &subtask, nil
}

func (s *SubtaskService) Update(id uint, title string, description string) (*domain.Subtask, error) {
	var subtask domain.Subtask

	if err := s.db.First(&subtask, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("subtask with id %d not found", id)
		}
		return nil, fmt.Errorf("error fetching subtask: %w", err)
	}

	updates := map[string]interface{}{
		"title":       title,
		"description": description,
	}

	if err := s.db.Model(&subtask).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("could not update subtask: %w", err)
	}

	return &subtask, nil
}

func (s *SubtaskService) ToggleDone(id uint) (*domain.Subtask, error) {
	var subtask domain.Subtask

	if err := s.db.First(&subtask, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("subtask with id %d not found", id)
		}
		return nil, fmt.Errorf("error fetching subtask: %w", err)
	}

	if err := s.db.Model(&subtask).Update("done", !subtask.Done).Error; err != nil {
		return nil, fmt.Errorf("could not toggle subtask: %w", err)
	}

	subtask.Done = !subtask.Done
	return &subtask, nil
}

func (s *SubtaskService) Delete(id uint) error {
	var subtask domain.Subtask

	if err := s.db.First(&subtask, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("subtask with id %d not found", id)
		}
		return fmt.Errorf("error fetching subtask: %w", err)
	}

	if err := s.db.Delete(&subtask).Error; err != nil {
		return fmt.Errorf("could not delete subtask: %w", err)
	}

	return nil
}

func (s *SubtaskService) GetByID(id uint) (*domain.Subtask, error) {
	var subtask domain.Subtask

	err := s.db.
		Preload("CreatedBy").
		First(&subtask, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("subtask with id %d not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("error fetching subtask: %w", err)
	}

	return &subtask, nil
}

func (s *SubtaskService) ListByTask(taskID uint) ([]domain.Subtask, error) {
	var subtasks []domain.Subtask

	if err := s.db.
		Where("task_id = ?", taskID).
		Order("created_at asc").
		Find(&subtasks).Error; err != nil {
		return nil, fmt.Errorf("could not list subtasks: %w", err)
	}

	return subtasks, nil
}
