package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type DiscordChannel struct {
	webhookURL string
}

func NewDiscordChannel(webhookURL string) *DiscordChannel {
	return &DiscordChannel{webhookURL: webhookURL}
}

func (d *DiscordChannel) Name() string { return "discord" }

func (d *DiscordChannel) Send(event Event) error {
	payload := map[string]interface{}{
		"content": fmt.Sprintf("**[SprintOS]** %s\n%s", event.Title, event.Details),
	}
	body, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(d.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("discord send failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("discord returned %d", resp.StatusCode)
	}
	return nil
}
