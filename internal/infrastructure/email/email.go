package email

import (
	"fmt"
	"net/smtp"

	"github.com/spf13/viper"
)

func SendInvitation(toEmail, orgName, token string) error {
	host := viper.GetString("smtp.host")
	port := viper.GetString("smtp.port")
	from := viper.GetString("smtp.from")
	password := viper.GetString("SMTP_PASSWORD")

	subject := fmt.Sprintf("You've been invited to join %s on SprintOS", orgName)
	body := fmt.Sprintf(`Hi!

You've been invited to join "%s" on SprintOS, a terminal-based project manager.

To accept this invitation, run the following command in your terminal:

  commandpm join --token %s

This invitation expires in 7 days.
`, orgName, token)

	message := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		from, toEmail, subject, body,
	)

	auth := smtp.PlainAuth("", from, password, host)
	addr := fmt.Sprintf("%s:%s", host, port)

	if err := smtp.SendMail(addr, auth, from, []string{toEmail}, []byte(message)); err != nil {
		return fmt.Errorf("could not send invitation email: %w", err)
	}

	return nil
}
