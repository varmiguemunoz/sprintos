package app

import (
	"errors"
	"fmt"

	"github.com/varmiguemunoz/sprintos/internal/domain"
	"gorm.io/gorm"
)

type SubtaskCommentService struct {
	db *gorm.DB
}

func NewSubtaskCommentService(db *gorm.DB) *SubtaskCommentService {
	return &SubtaskCommentService{db: db}
}

func (s *SubtaskCommentService) Create(content string, subtaskID uint, authorID uint) (*domain.SubtaskComment, error) {
	comment := domain.SubtaskComment{
		Content:   content,
		SubtaskID: subtaskID,
		AuthorID:  authorID,
	}

	if err := s.db.Create(&comment).Error; err != nil {
		return nil, fmt.Errorf("could not create subtask comment: %w", err)
	}

	return &comment, nil
}

func (s *SubtaskCommentService) Update(id uint, content string) (*domain.SubtaskComment, error) {
	var comment domain.SubtaskComment

	if err := s.db.First(&comment, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("subtask comment with id %d not found", id)
		}
		return nil, fmt.Errorf("error fetching subtask comment: %w", err)
	}

	if err := s.db.Model(&comment).Update("content", content).Error; err != nil {
		return nil, fmt.Errorf("could not update subtask comment: %w", err)
	}

	return &comment, nil
}

func (s *SubtaskCommentService) ListBySubtask(subtaskID uint) ([]domain.SubtaskComment, error) {
	var comments []domain.SubtaskComment

	if err := s.db.Preload("Author").
		Where("subtask_id = ?", subtaskID).
		Order("created_at asc").
		Find(&comments).Error; err != nil {
		return nil, fmt.Errorf("could not list subtask comments: %w", err)
	}

	return comments, nil
}
