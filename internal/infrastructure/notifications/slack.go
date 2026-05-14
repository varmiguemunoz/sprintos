package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type SlackChannel struct {
	webhookURL string
}

func NewSlackChannel(webhookURL string) *SlackChannel {
	return &SlackChannel{webhookURL: webhookURL}
}

func (s *SlackChannel) Name() string { return "slack" }

func (s *SlackChannel) Send(event Event) error {
	payload := map[string]interface{}{
		"text": fmt.Sprintf("*[SprintOS]* %s\n%s", event.Title, event.Details),
	}
	body, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(s.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("slack send failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("slack returned %d", resp.StatusCode)
	}
	return nil
}
