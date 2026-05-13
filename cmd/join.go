package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/domain"
	"github.com/varmiguemunoz/sprintos/internal/infrastructure/auth"
)

var joinCmd = &cobra.Command{
	Use:   "join",
	Short: "Accept an invitation to join an organization",
	RunE: func(cmd *cobra.Command, args []string) error {
		token, _ := cmd.Flags().GetString("token")
		if token == "" {
			return fmt.Errorf("--token is required")
		}

		invitationSvc := app.NewInvitationService(DB)
		userSvc := app.NewUserService(DB)
		teamSvc := app.NewTeamService(DB)

		inv, err := invitationSvc.GetByToken(token)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Invalid or expired invitation token.")
			return err
		}

		if inv.AcceptedAt != nil {
			return fmt.Errorf("this invitation has already been accepted")
		}

		fmt.Println("Opening GitHub login to verify your identity...")

		auth.SetupProviders()
		gothUser, err := auth.StartLogin("github")
		if err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}

		user, _, err := userSvc.FindOrCreateByOAuth(
			gothUser.Provider,
			gothUser.UserID,
			gothUser.Email,
			gothUser.Name,
			&gothUser.AvatarURL,
		)
		if err != nil {
			return fmt.Errorf("could not resolve user: %w", err)
		}

		_, err = teamSvc.AddMember(user.ID, inv.OrganizationID, "user")
		if err != nil {
			return fmt.Errorf("could not add you to the organization: %w", err)
		}

		if _, err := invitationSvc.Accept(token); err != nil {
			return fmt.Errorf("could not mark invitation as accepted: %w", err)
		}

		if err := auth.SaveSession(&domain.User{
			Name:       user.Name,
			Email:      user.Email,
			Provider:   user.Provider,
			ProviderID: user.ProviderID,
		}); err != nil {
			return fmt.Errorf("could not save session: %w", err)
		}

		fmt.Printf("\n✓ Welcome! You have joined \"%s\".\n", inv.Organization.Name)
		fmt.Println("Run `commandpm start` to launch the app.")

		return nil
	},
}

func init() {
	joinCmd.Flags().String("token", "", "Invitation token from your email")
	rootCmd.AddCommand(joinCmd)
}
