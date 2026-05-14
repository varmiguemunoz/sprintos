package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/infrastructure/auth"
)

var notifyCmd = &cobra.Command{
	Use:   "notify",
	Short: "Manage notification channels",
}

var notifyTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Send a test notification to all configured channels",
	RunE: func(cmd *cobra.Command, args []string) error {
		session, err := auth.LoadSession()
		if err != nil {
			return fmt.Errorf("not logged in — run `sprintos start` first")
		}
		userSvc := app.NewUserService(DB)
		orgSvc := app.NewOrganizationService(DB)
		notifSvc := app.NewNotificationService(DB)

		user, err := userSvc.GetByEmail(session.Email)
		if err != nil {
			return fmt.Errorf("could not load user: %w", err)
		}
		org, err := orgSvc.GetByOwnerID(user.ID)
		if err != nil {
			return fmt.Errorf("could not load organization: %w", err)
		}

		fmt.Println("Sending test notifications...")
		results := notifSvc.TestAll(org.ID)
		for _, r := range results {
			fmt.Println(" ", r)
		}
		return nil
	},
}

var notifyConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure a notification channel (slack or discord)",
	RunE: func(cmd *cobra.Command, args []string) error {
		session, err := auth.LoadSession()
		if err != nil {
			return fmt.Errorf("not logged in — run `sprintos start` first")
		}
		userSvc := app.NewUserService(DB)
		orgSvc := app.NewOrganizationService(DB)
		notifSvc := app.NewNotificationService(DB)

		user, err := userSvc.GetByEmail(session.Email)
		if err != nil {
			return fmt.Errorf("could not load user: %w", err)
		}
		org, err := orgSvc.GetByOwnerID(user.ID)
		if err != nil {
			return fmt.Errorf("could not load organization: %w", err)
		}

		reader := bufio.NewReader(os.Stdin)

		fmt.Println("SprintOS — Configure Notifications")
		fmt.Println("Available channels: slack, discord")
		fmt.Print("\nChannel name: ")
		channel, _ := reader.ReadString('\n')
		channel = strings.TrimSpace(channel)

		if channel != "slack" && channel != "discord" {
			return fmt.Errorf("unsupported channel — choose slack or discord")
		}

		fmt.Printf("Webhook URL for %s: ", channel)
		webhookURL, _ := reader.ReadString('\n')
		webhookURL = strings.TrimSpace(webhookURL)

		if err := notifSvc.SaveConfig(org.ID, channel, webhookURL); err != nil {
			return fmt.Errorf("could not save config: %w", err)
		}

		fmt.Printf("\n✓ %s configured. Run `sprintos notify test` to verify.\n", channel)
		return nil
	},
}

var notifyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured notification channels",
	RunE: func(cmd *cobra.Command, args []string) error {
		session, err := auth.LoadSession()
		if err != nil {
			return fmt.Errorf("not logged in — run `sprintos start` first")
		}
		userSvc := app.NewUserService(DB)
		orgSvc := app.NewOrganizationService(DB)
		notifSvc := app.NewNotificationService(DB)

		user, err := userSvc.GetByEmail(session.Email)
		if err != nil {
			return fmt.Errorf("could not load user: %w", err)
		}
		org, err := orgSvc.GetByOwnerID(user.ID)
		if err != nil {
			return fmt.Errorf("could not load organization: %w", err)
		}

		configs, err := notifSvc.ListConfigs(org.ID)
		if err != nil {
			return fmt.Errorf("could not list configs: %w", err)
		}

		if len(configs) == 0 {
			fmt.Println("No channels configured.")
			fmt.Println("Run: sprintos notify config")
			return nil
		}

		fmt.Println("Configured channels:")
		for _, c := range configs {
			status := "enabled"
			if !c.Enabled {
				status = "disabled"
			}
			fmt.Printf("  %-10s  %-10s  %s\n", c.Channel, status, c.WebhookURL[:min(40, len(c.WebhookURL))]+"...")
		}
		return nil
	},
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func init() {
	notifyCmd.AddCommand(notifyTestCmd)
	notifyCmd.AddCommand(notifyConfigCmd)
	notifyCmd.AddCommand(notifyListCmd)
	rootCmd.AddCommand(notifyCmd)
}
