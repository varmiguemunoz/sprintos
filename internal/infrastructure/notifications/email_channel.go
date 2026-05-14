package notifications

import (
	"fmt"
	"net/smtp"

	"github.com/varmiguemunoz/sprintos/internal/config"
)

type EmailChannel struct {
	toEmail string
}

func NewEmailChannel(toEmail string) *EmailChannel {
	return &EmailChannel{toEmail: toEmail}
}

func (e *EmailChannel) Name() string { return "email" }

func (e *EmailChannel) Send(event Event) error {
	host := config.GetSMTPHost()
	port := config.GetSMTPPort()
	from := config.GetSMTPFrom()
	password := config.GetSMTPPassword()

	if host == "" || from == "" {
		return fmt.Errorf("SMTP not configured")
	}

	message := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: [SprintOS] %s\r\n\r\n%s\r\n%s",
		from, e.toEmail, event.Title, event.Details, event.URL,
	)

	auth := smtp.PlainAuth("", from, password, host)
	return smtp.SendMail(fmt.Sprintf("%s:%s", host, port), auth, from, []string{e.toEmail}, []byte(message))
}
