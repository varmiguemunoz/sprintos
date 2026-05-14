package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/infrastructure/auth"
)

var apiKeyCmd = &cobra.Command{
	Use:   "api-key",
	Short: "Manage API keys",
}

var apiKeyCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Generate a new API key",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("--name is required")
		}

		session, err := auth.LoadSession()
		if err != nil {
			return fmt.Errorf("not logged in — run `sprintos start` first")
		}

		userSvc := app.NewUserService(DB)
		orgSvc := app.NewOrganizationService(DB)
		apiKeySvc := app.NewAPIKeyService(DB)

		user, err := userSvc.GetByEmail(session.Email)
		if err != nil {
			return fmt.Errorf("could not load user: %w", err)
		}

		org, err := orgSvc.GetByOwnerID(user.ID)
		if err != nil {
			return fmt.Errorf("could not load organization: %w", err)
		}

		raw, _, err := apiKeySvc.Create(name, user.ID, org.ID)
		if err != nil {
			return fmt.Errorf("could not create key: %w", err)
		}

		fmt.Println("API key created. Copy it now — it will not be shown again.")
		fmt.Println("")
		fmt.Printf("  %s\n", raw)
		fmt.Println("")
		fmt.Println("Use it in requests:")
		fmt.Printf("  Authorization: Bearer %s\n", raw)
		return nil
	},
}

var apiKeyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all API keys for your organization",
	RunE: func(cmd *cobra.Command, args []string) error {
		session, err := auth.LoadSession()
		if err != nil {
			return fmt.Errorf("not logged in — run `sprintos start` first")
		}

		userSvc := app.NewUserService(DB)
		orgSvc := app.NewOrganizationService(DB)
		apiKeySvc := app.NewAPIKeyService(DB)

		user, err := userSvc.GetByEmail(session.Email)
		if err != nil {
			return fmt.Errorf("could not load user: %w", err)
		}

		org, err := orgSvc.GetByOwnerID(user.ID)
		if err != nil {
			return fmt.Errorf("could not load organization: %w", err)
		}

		keys, err := apiKeySvc.ListByOrg(org.ID)
		if err != nil {
			return fmt.Errorf("could not list keys: %w", err)
		}

		if len(keys) == 0 {
			fmt.Println("No API keys. Create one with: sprintos api-key create --name \"my-key\"")
			return nil
		}

		fmt.Printf("%-5s  %-20s  %-15s  %s\n", "ID", "Name", "Prefix", "Last Used")
		for _, k := range keys {
			lastUsed := "never"
			if k.LastUsedAt != nil {
				lastUsed = k.LastUsedAt.Format("2006-01-02 15:04")
			}
			fmt.Printf("%-5d  %-20s  %-15s  %s\n", k.ID, k.Name, k.KeyPrefix, lastUsed)
		}
		return nil
	},
}

var apiKeyRevokeCmd = &cobra.Command{
	Use:   "revoke",
	Short: "Revoke an API key",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetUint("id")
		if id == 0 {
			return fmt.Errorf("--id is required")
		}

		session, err := auth.LoadSession()
		if err != nil {
			return fmt.Errorf("not logged in — run `sprintos start` first")
		}

		userSvc := app.NewUserService(DB)
		orgSvc := app.NewOrganizationService(DB)
		apiKeySvc := app.NewAPIKeyService(DB)

		user, err := userSvc.GetByEmail(session.Email)
		if err != nil {
			return fmt.Errorf("could not load user: %w", err)
		}

		org, err := orgSvc.GetByOwnerID(user.ID)
		if err != nil {
			return fmt.Errorf("could not load organization: %w", err)
		}

		if err := apiKeySvc.Revoke(id, org.ID); err != nil {
			return fmt.Errorf("could not revoke key: %w", err)
		}

		fmt.Printf("API key %d revoked.\n", id)
		return nil
	},
}

func init() {
	apiKeyCreateCmd.Flags().String("name", "", "Name for the API key")
	apiKeyRevokeCmd.Flags().Uint("id", 0, "ID of the key to revoke")
	apiKeyCmd.AddCommand(apiKeyCreateCmd)
	apiKeyCmd.AddCommand(apiKeyListCmd)
	apiKeyCmd.AddCommand(apiKeyRevokeCmd)
	rootCmd.AddCommand(apiKeyCmd)
}
