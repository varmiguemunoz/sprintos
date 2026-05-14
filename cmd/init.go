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

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize SprintOS in the current repository (.sprintos config file)",
	RunE: func(cmd *cobra.Command, args []string) error {
		session, err := auth.LoadSession()
		if err != nil {
			return fmt.Errorf("not logged in — run `sprintos start` first")
		}

		userSvc := app.NewUserService(DB)
		orgSvc := app.NewOrganizationService(DB)
		projectSvc := app.NewProjectService(DB)

		user, err := userSvc.GetByEmail(session.Email)
		if err != nil {
			return fmt.Errorf("could not load user: %w", err)
		}

		org, err := orgSvc.GetByOwnerID(user.ID)
		if err != nil {
			return fmt.Errorf("could not load organization: %w", err)
		}

		projects, err := projectSvc.ListByOrganization(org.ID)
		if err != nil || len(projects) == 0 {
			return fmt.Errorf("no projects found — create one with `sprintos start`")
		}

		fmt.Println("SprintOS — Init")
		fmt.Printf("Organization: %s\n\n", org.Name)
		fmt.Println("Available projects:")
		for _, p := range projects {
			fmt.Printf("  [%d] %s\n", p.ID, p.Name)
		}

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("\nSelect project ID: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		projectID, err := strconv.ParseUint(input, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid project ID")
		}

		var selectedProject app.ProjectInfo
		for _, p := range projects {
			if p.ID == uint(projectID) {
				selectedProject = app.ProjectInfo{ID: p.ID, Name: p.Name}
				break
			}
		}
		if selectedProject.ID == 0 {
			return fmt.Errorf("project %d not found", projectID)
		}

		cfg := app.LocalConfig{
			ProjectID:   selectedProject.ID,
			ProjectName: selectedProject.Name,
			OrgID:       org.ID,
		}

		if err := app.SaveLocalConfig(cfg); err != nil {
			return fmt.Errorf("could not save .sprintos: %w", err)
		}

		fmt.Printf("\n✓ .sprintos created\n")
		fmt.Printf("  Project: %s (ID: %d)\n", selectedProject.Name, selectedProject.ID)
		fmt.Printf("\nYou can now run:\n")
		fmt.Printf("  sprintos task ls\n")
		fmt.Printf("  sprintos task create \"Fix login bug\" --state backlog\n")
		fmt.Printf("  sprintos task move 5 \"In Review\"\n")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
