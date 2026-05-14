package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/varmiguemunoz/sprintos/internal/config"
)

type WhatsAppChannel struct {
	phoneNumber string
}

func NewWhatsAppChannel(phoneNumber string) *WhatsAppChannel {
	return &WhatsAppChannel{phoneNumber: phoneNumber}
}

func (w *WhatsAppChannel) Name() string { return "whatsapp" }

func (w *WhatsAppChannel) Send(event Event) error {
	apiURL := config.GetEvolutionAPIURL()
	apiToken := config.GetEvolutionAPIToken()

	if apiURL == "" {
		return fmt.Errorf("Evolution API not configured")
	}

	payload := map[string]interface{}{
		"number":  w.phoneNumber,
		"options": map[string]interface{}{"delay": 0},
		"textMessage": map[string]interface{}{
			"text": fmt.Sprintf("*[SprintOS]* %s\n%s", event.Title, event.Details),
		},
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", apiURL+"/message/sendText", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("whatsapp request error: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", apiToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("whatsapp send failed: %w", err)
	}
	defer resp.Body.Close()
	return nil
}
