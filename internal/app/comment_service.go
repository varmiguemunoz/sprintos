package app

import (
	"errors"
	"fmt"

	"github.com/varmiguemunoz/command_pm_app/internal/domain"
	"gorm.io/gorm"
)

type CommentService struct {
	db *gorm.DB
}

// Constructor.
func NewCommentService(db *gorm.DB) *CommentService {
	return &CommentService{db: db}
}

// Create adds a new comment to a task.
func (s *CommentService) Create(content string, taskID uint, authorID uint) (*domain.Comment, error) {
	comment := domain.Comment{
		Content:  content,
		TaskID:   taskID,
		AuthorID: authorID,
	}

	if err := s.db.Create(&comment).Error; err != nil {
		return nil, fmt.Errorf("could not create comment: %w", err)
	}

	return &comment, nil
}

// ListByTask returns all comments for a task, ordered by creation time.
// Preloads the Author so callers can display the commenter's name directly.
func (s *CommentService) ListByTask(taskID uint) ([]domain.Comment, error) {
	var comments []domain.Comment

	if err := s.db.Preload("Author").
		Where("task_id = ?", taskID).
		Order("created_at asc").
		Find(&comments).Error; err != nil {
		return nil, fmt.Errorf("could not list comments: %w", err)
	}

	return comments, nil
}

// Delete soft-deletes a comment by its primary key.
func (s *CommentService) Delete(id uint) error {
	var comment domain.Comment

	if err := s.db.First(&comment, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("comment with id %d not found", id)
		}
		return fmt.Errorf("error fetching comment: %w", err)
	}

	if err := s.db.Delete(&comment).Error; err != nil {
		return fmt.Errorf("could not delete comment: %w", err)
	}

	return nil
}
