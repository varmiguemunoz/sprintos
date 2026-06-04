package app

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/varmiguemunoz/sprintos/internal/domain"
	"gorm.io/gorm"
)

type TimeEntryService struct {
	db *gorm.DB
}

func NewTimeEntryService(db *gorm.DB) *TimeEntryService {
	return &TimeEntryService{db: db}
}

func (s *TimeEntryService) LogManual(taskID *uint, subtaskID *uint, userID uint, minutes int, note string) (*domain.TimeEntry, error) {
	if minutes <= 0 {
		return nil, fmt.Errorf("minutes must be greater than zero")
	}

	var n *string
	if note != "" {
		n = &note
	}

	entry := domain.TimeEntry{
		TaskID:    taskID,
		SubtaskID: subtaskID,
		UserID:    userID,
		Minutes:   minutes,
		Note:      n,
		LoggedAt:  time.Now(),
	}

	if err := s.db.Create(&entry).Error; err != nil {
		return nil, fmt.Errorf("could not log time entry: %w", err)
	}

	return &entry, nil
}

func (s *TimeEntryService) StartTimer(taskID *uint, subtaskID *uint, userID uint) (*domain.ActiveTimer, error) {
	if _, err := s.stopAndSaveExistingTimer(userID); err != nil {
		return nil, err
	}

	timer := domain.ActiveTimer{
		UserID:    userID,
		TaskID:    taskID,
		SubtaskID: subtaskID,
		StartedAt: time.Now(),
	}

	if err := s.db.Create(&timer).Error; err != nil {
		return nil, fmt.Errorf("could not start timer: %w", err)
	}

	return &timer, nil
}

func (s *TimeEntryService) StopTimer(userID uint) (*domain.TimeEntry, error) {
	return s.stopAndSaveExistingTimer(userID)
}

func (s *TimeEntryService) stopAndSaveExistingTimer(userID uint) (*domain.TimeEntry, error) {
	var timer domain.ActiveTimer

	err := s.db.Where("user_id = ?", userID).First(&timer).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error fetching active timer: %w", err)
	}

	elapsed := time.Since(timer.StartedAt)
	minutes := int(math.Ceil(elapsed.Minutes()))
	if minutes < 1 {
		minutes = 1
	}

	entry := domain.TimeEntry{
		TaskID:    timer.TaskID,
		SubtaskID: timer.SubtaskID,
		UserID:    userID,
		Minutes:   minutes,
		LoggedAt:  time.Now(),
	}

	if err := s.db.Create(&entry).Error; err != nil {
		return nil, fmt.Errorf("could not save time entry: %w", err)
	}

	if err := s.db.Delete(&timer).Error; err != nil {
		return nil, fmt.Errorf("could not delete active timer: %w", err)
	}

	return &entry, nil
}

func (s *TimeEntryService) GetActiveTimer(userID uint) (*domain.ActiveTimer, error) {
	var timer domain.ActiveTimer

	err := s.db.Where("user_id = ?", userID).First(&timer).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error fetching active timer: %w", err)
	}

	return &timer, nil
}

func (s *TimeEntryService) GetTotalMinutesForTask(taskID uint) int {
	var total int
	s.db.Raw("SELECT COALESCE(SUM(minutes), 0) FROM time_entries WHERE task_id = ? AND deleted_at IS NULL", taskID).Scan(&total)
	return total
}

func (s *TimeEntryService) GetTotalMinutesForSubtask(subtaskID uint) int {
	var total int
	s.db.Raw("SELECT COALESCE(SUM(minutes), 0) FROM time_entries WHERE subtask_id = ? AND deleted_at IS NULL", subtaskID).Scan(&total)
	return total
}

func (s *TimeEntryService) GetTotalMinutesForTaskWithSubtasks(taskID uint) int {
	var subtaskIDs []uint
	s.db.Raw("SELECT id FROM subtasks WHERE task_id = ? AND deleted_at IS NULL", taskID).Scan(&subtaskIDs)

	if len(subtaskIDs) == 0 {
		return s.GetTotalMinutesForTask(taskID)
	}

	var total int
	s.db.Raw("SELECT COALESCE(SUM(minutes), 0) FROM time_entries WHERE subtask_id IN ? AND deleted_at IS NULL", subtaskIDs).Scan(&total)
	return total
}

func (s *TimeEntryService) ListByTask(taskID uint) ([]domain.TimeEntry, error) {
	var entries []domain.TimeEntry

	if err := s.db.Preload("User").
		Where("task_id = ?", taskID).
		Order("logged_at desc").
		Find(&entries).Error; err != nil {
		return nil, fmt.Errorf("could not list time entries: %w", err)
	}

	return entries, nil
}

func (s *TimeEntryService) ListBySubtask(subtaskID uint) ([]domain.TimeEntry, error) {
	var entries []domain.TimeEntry

	if err := s.db.Preload("User").
		Where("subtask_id = ?", subtaskID).
		Order("logged_at desc").
		Find(&entries).Error; err != nil {
		return nil, fmt.Errorf("could not list subtask time entries: %w", err)
	}

	return entries, nil
}

func FormatMinutes(minutes int) string {
	if minutes == 0 {
		return "0m"
	}
	h := minutes / 60
	m := minutes % 60
	if h == 0 {
		return fmt.Sprintf("%dm", m)
	}
	if m == 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dh %dm", h, m)
}

func FormatElapsed(since time.Time) string {
	elapsed := time.Since(since)
	h := int(elapsed.Hours())
	m := int(elapsed.Minutes()) % 60
	sec := int(elapsed.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, sec)
}
