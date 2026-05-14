package app

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/varmiguemunoz/sprintos/internal/domain"
	"github.com/varmiguemunoz/sprintos/internal/infrastructure/notifications"
	"gorm.io/gorm"
)

var mentionPattern = regexp.MustCompile(`@(\w+)`)

type NotificationService struct {
	db *gorm.DB
}

func NewNotificationService(db *gorm.DB) *NotificationService {
	return &NotificationService{db: db}
}

func (s *NotificationService) getChannels(orgID uint) []notifications.Channel {
	var configs []domain.NotificationConfig
	s.db.Where("org_id = ? AND enabled = true", orgID).Find(&configs)

	var channels []notifications.Channel
	for _, c := range configs {
		switch c.Channel {
		case "slack":
			channels = append(channels, notifications.NewSlackChannel(c.WebhookURL))
		case "discord":
			channels = append(channels, notifications.NewDiscordChannel(c.WebhookURL))
		}
	}
	return channels
}

func (s *NotificationService) Fire(orgID uint, event notifications.Event) {
	channels := s.getChannels(orgID)
	for _, ch := range channels {
		go func(c notifications.Channel) {
			_ = c.Send(event)
		}(ch)
	}
}

func (s *NotificationService) TaskCreated(orgID uint, task *domain.Task, projectName string) {
	s.Fire(orgID, notifications.Event{
		Type:    "task.created",
		OrgID:   orgID,
		Title:   fmt.Sprintf("New task in %s", projectName),
		Details: fmt.Sprintf("#%d %s", task.TaskNumber, task.Title),
	})
}

func (s *NotificationService) TaskMoved(orgID uint, task *domain.Task, fromState, toState, projectName string) {
	s.Fire(orgID, notifications.Event{
		Type:    "task.moved",
		OrgID:   orgID,
		Title:   fmt.Sprintf("Task moved in %s", projectName),
		Details: fmt.Sprintf("#%d %s\n%s → %s", task.TaskNumber, task.Title, fromState, toState),
	})
}

func (s *NotificationService) TaskCompleted(orgID uint, task *domain.Task, projectName, comments string) {
	s.Fire(orgID, notifications.Event{
		Type:    "task.completed",
		OrgID:   orgID,
		Title:   fmt.Sprintf("Task completed in %s", projectName),
		Details: fmt.Sprintf("#%d %s\nComments: %s", task.TaskNumber, task.Title, comments),
	})
}

func (s *NotificationService) TaskAssigned(orgID uint, task *domain.Task, assigneeName, projectName string) {
	s.Fire(orgID, notifications.Event{
		Type:    "task.assigned",
		OrgID:   orgID,
		Title:   fmt.Sprintf("Task assigned in %s", projectName),
		Details: fmt.Sprintf("#%d %s → assigned to %s", task.TaskNumber, task.Title, assigneeName),
	})
}

func (s *NotificationService) CommentCreated(orgID uint, task *domain.Task, authorName, content, projectName string) {
	s.Fire(orgID, notifications.Event{
		Type:    "comment.created",
		OrgID:   orgID,
		Title:   fmt.Sprintf("New comment in %s", projectName),
		Details: fmt.Sprintf("#%d %s\n%s: %s", task.TaskNumber, task.Title, authorName, content),
	})

	mentions := mentionPattern.FindAllStringSubmatch(content, -1)
	for _, m := range mentions {
		username := strings.ToLower(m[1])
		s.notifyMention(orgID, username, authorName, task, content)
	}
}

func (s *NotificationService) notifyMention(orgID uint, username, authorName string, task *domain.Task, content string) {
	var user domain.User
	if err := s.db.Where("LOWER(name) = ?", username).First(&user).Error; err != nil {
		return
	}
	if user.Email == "" {
		return
	}
	ch := notifications.NewEmailChannel(user.Email)
	_ = ch.Send(notifications.Event{
		Type:    "mention",
		OrgID:   orgID,
		Title:   fmt.Sprintf("%s mentioned you on task #%d", authorName, task.TaskNumber),
		Details: content,
	})
}

func (s *NotificationService) OverdueAlert(orgID uint, tasks []domain.Task) {
	if len(tasks) == 0 {
		return
	}
	lines := fmt.Sprintf("%d overdue task(s):\n", len(tasks))
	for _, t := range tasks {
		lines += fmt.Sprintf("  #%d %s\n", t.TaskNumber, t.Title)
	}
	s.Fire(orgID, notifications.Event{
		Type:    "task.overdue",
		OrgID:   orgID,
		Title:   "Overdue tasks need attention",
		Details: lines,
	})
}

func (s *NotificationService) SaveConfig(orgID uint, channel, webhookURL string) error {
	var existing domain.NotificationConfig
	err := s.db.Where("org_id = ? AND channel = ?", orgID, channel).First(&existing).Error
	if err == nil {
		return s.db.Model(&existing).Updates(map[string]interface{}{
			"webhook_url": webhookURL,
			"enabled":     true,
		}).Error
	}
	config := domain.NotificationConfig{
		OrgID:      orgID,
		Channel:    channel,
		WebhookURL: webhookURL,
		Enabled:    true,
	}
	return s.db.Create(&config).Error
}

func (s *NotificationService) ListConfigs(orgID uint) ([]domain.NotificationConfig, error) {
	var configs []domain.NotificationConfig
	err := s.db.Where("org_id = ?", orgID).Find(&configs).Error
	return configs, err
}

func (s *NotificationService) TestAll(orgID uint) []string {
	channels := s.getChannels(orgID)
	var results []string
	event := notifications.Event{
		Type:    "test",
		OrgID:   orgID,
		Title:   "SprintOS test notification",
		Details: "If you see this, notifications are working correctly.",
	}
	if len(channels) == 0 {
		return []string{"No channels configured. Run: sprintos notify config"}
	}
	for _, ch := range channels {
		if err := ch.Send(event); err != nil {
			results = append(results, fmt.Sprintf("✗ %s: %s", ch.Name(), err))
		} else {
			results = append(results, fmt.Sprintf("✓ %s: sent", ch.Name()))
		}
	}
	return results
}
