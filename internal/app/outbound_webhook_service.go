package app

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/varmiguemunoz/sprintos/internal/domain"
	"gorm.io/gorm"
)

type OutboundWebhookService struct {
	db *gorm.DB
}

func NewOutboundWebhookService(db *gorm.DB) *OutboundWebhookService {
	return &OutboundWebhookService{db: db}
}

func (s *OutboundWebhookService) Create(orgID uint, url string, events []string) (*domain.OutboundWebhook, error) {
	b := make([]byte, 16)
	rand.Read(b)
	secret := hex.EncodeToString(b)

	hook := domain.OutboundWebhook{
		OrgID:  orgID,
		URL:    url,
		Secret: secret,
		Events: strings.Join(events, ","),
		Active: true,
	}

	if err := s.db.Create(&hook).Error; err != nil {
		return nil, fmt.Errorf("could not create webhook: %w", err)
	}

	return &hook, nil
}

func (s *OutboundWebhookService) ListByOrg(orgID uint) ([]domain.OutboundWebhook, error) {
	var hooks []domain.OutboundWebhook
	if err := s.db.Where("org_id = ? AND active = true", orgID).Find(&hooks).Error; err != nil {
		return nil, fmt.Errorf("could not list webhooks: %w", err)
	}
	return hooks, nil
}

func (s *OutboundWebhookService) Delete(id, orgID uint) error {
	return s.db.Where("id = ? AND org_id = ?", id, orgID).Delete(&domain.OutboundWebhook{}).Error
}

type WebhookPayload struct {
	Event     string      `json:"event"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

func (s *OutboundWebhookService) Fire(orgID uint, event string, data interface{}) {
	hooks, err := s.ListByOrg(orgID)
	if err != nil || len(hooks) == 0 {
		return
	}

	payload := WebhookPayload{
		Event:     event,
		Timestamp: time.Now(),
		Data:      data,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return
	}

	for _, hook := range hooks {
		if !strings.Contains(hook.Events, event) {
			continue
		}
		go func(h domain.OutboundWebhook) {
			sig := signPayload(body, h.Secret)
			req, err := http.NewRequest("POST", h.URL, bytes.NewReader(body))
			if err != nil {
				return
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-SprintOS-Signature", sig)
			req.Header.Set("X-SprintOS-Event", event)
			client := &http.Client{Timeout: 10 * time.Second}
			client.Do(req)
		}(hook)
	}
}

func signPayload(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
