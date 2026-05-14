package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/infrastructure/auth"
)

var githubCmd = &cobra.Command{
	Use:   "github",
	Short: "Manage GitHub integration",
}

var githubSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Connect a GitHub repository to a SprintOS project",
	RunE: func(cmd *cobra.Command, args []string) error {
		session, err := auth.LoadSession()
		if err != nil {
			return fmt.Errorf("not logged in — run `sprintos start` first")
		}

		userSvc := app.NewUserService(DB)
		orgSvc := app.NewOrganizationService(DB)
		projectSvc := app.NewProjectService(DB)
		stateSvc := app.NewStateService(DB)
		githubSvc := app.NewGitHubService(DB)

		user, err := userSvc.GetByEmail(session.Email)
		if err != nil {
			return fmt.Errorf("could not load user: %w", err)
		}

		org, err := orgSvc.GetByOwnerID(user.ID)
		if err != nil {
			return fmt.Errorf("could not load organization: %w", err)
		}

		reader := bufio.NewReader(os.Stdin)

		fmt.Println("SprintOS — GitHub Integration Setup")
		fmt.Println("────────────────────────────────────")

		fmt.Print("\nGitHub repository (format: owner/repo): ")
		repoInput, _ := reader.ReadString('\n')
		repoInput = strings.TrimSpace(repoInput)
		parts := strings.SplitN(repoInput, "/", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid format — use owner/repo")
		}
		repoOwner, repoName := parts[0], parts[1]

		projects, err := projectSvc.ListByOrganization(org.ID)
		if err != nil {
			return fmt.Errorf("could not list projects: %w", err)
		}

		fmt.Println("\nAvailable projects:")
		for _, p := range projects {
			fmt.Printf("  [%d] %s\n", p.ID, p.Name)
		}

		fmt.Print("\nProject ID to link: ")
		projectInput, _ := reader.ReadString('\n')
		projectID, err := strconv.ParseUint(strings.TrimSpace(projectInput), 10, 64)
		if err != nil {
			return fmt.Errorf("invalid project ID")
		}

		states, err := stateSvc.ListByProject(uint(projectID))
		if err != nil {
			return fmt.Errorf("could not list states: %w", err)
		}

		fmt.Println("\nAvailable states:")
		for _, s := range states {
			fmt.Printf("  [%d] %s\n", s.ID, s.Name)
		}

		fmt.Print("\nState ID for 'In Review' (when PR is opened): ")
		inReviewInput, _ := reader.ReadString('\n')
		inReviewStateID, err := strconv.ParseUint(strings.TrimSpace(inReviewInput), 10, 64)
		if err != nil {
			return fmt.Errorf("invalid state ID")
		}

		fmt.Print("State ID for 'Done' (when PR is merged): ")
		doneInput, _ := reader.ReadString('\n')
		doneStateID, err := strconv.ParseUint(strings.TrimSpace(doneInput), 10, 64)
		if err != nil {
			return fmt.Errorf("invalid state ID")
		}

		integration, err := githubSvc.CreateIntegration(
			org.ID,
			repoOwner,
			repoName,
			uint(projectID),
			uint(inReviewStateID),
			uint(doneStateID),
		)
		if err != nil {
			return fmt.Errorf("could not create integration: %w", err)
		}

		fmt.Println("\n✓ Integration created successfully!")
		fmt.Println("\n────────────────────────────────────")
		fmt.Println("Add this webhook to your GitHub repository:")
		fmt.Printf("  URL:         http://YOUR_SERVER_IP:8090/webhooks/github\n")
		fmt.Printf("  Secret:      %s\n", integration.WebhookSecret)
		fmt.Printf("  Content type: application/json\n")
		fmt.Printf("  Events:      Pull requests\n")
		fmt.Println("\nGitHub → Repository → Settings → Webhooks → Add webhook")
		fmt.Println("\nThen start the webhook server:")
		fmt.Println("  sprintos serve --port 8090")
		fmt.Println("\nIn your PRs use the task ID in the title or branch name:")
		fmt.Printf("  Example: '%s-1: Fix login bug' or branch 'feature/%s-1-fix-login'\n",
			strings.ToUpper(org.Prefix), strings.ToUpper(org.Prefix))

		return nil
	},
}

var githubListCmd = &cobra.Command{
	Use:   "list",
	Short: "List GitHub integrations for your organization",
	RunE: func(cmd *cobra.Command, args []string) error {
		session, err := auth.LoadSession()
		if err != nil {
			return fmt.Errorf("not logged in — run `sprintos start` first")
		}

		userSvc := app.NewUserService(DB)
		orgSvc := app.NewOrganizationService(DB)
		githubSvc := app.NewGitHubService(DB)

		user, err := userSvc.GetByEmail(session.Email)
		if err != nil {
			return fmt.Errorf("could not load user: %w", err)
		}

		org, err := orgSvc.GetByOwnerID(user.ID)
		if err != nil {
			return fmt.Errorf("could not load organization: %w", err)
		}

		integrations, err := githubSvc.ListByOrg(org.ID)
		if err != nil {
			return fmt.Errorf("could not list integrations: %w", err)
		}

		if len(integrations) == 0 {
			fmt.Println("No GitHub integrations configured.")
			fmt.Println("Run: sprintos github setup")
			return nil
		}

		fmt.Println("GitHub Integrations:")
		for _, i := range integrations {
			fmt.Printf("  [%d] %s/%s → Project %d\n", i.ID, i.RepoOwner, i.RepoName, i.ProjectID)
		}

		return nil
	},
}

func init() {
	githubCmd.AddCommand(githubSetupCmd)
	githubCmd.AddCommand(githubListCmd)
	rootCmd.AddCommand(githubCmd)
}
