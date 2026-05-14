package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/infrastructure/auth"
)

var myTasksCmd = &cobra.Command{
	Use:   "my-tasks",
	Short: "Show all tasks assigned to you across all projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		session, err := auth.LoadSession()
		if err != nil {
			return fmt.Errorf("not logged in — run `sprintos start` first")
		}

		userSvc := app.NewUserService(DB)
		orgSvc := app.NewOrganizationService(DB)
		projectSvc := app.NewProjectService(DB)
		taskSvc := app.NewTaskService(DB)
		stateSvc := app.NewStateService(DB)

		user, err := userSvc.GetByEmail(session.Email)
		if err != nil {
			return fmt.Errorf("could not load user: %w", err)
		}

		org, err := orgSvc.GetByOwnerID(user.ID)
		if err != nil {
			return fmt.Errorf("could not load organization: %w", err)
		}

		projects, _ := projectSvc.ListByOrganization(org.ID)

		fmt.Printf("\nTasks assigned to %s\n", user.Name)
		fmt.Println("─────────────────────────────────────────")

		total := 0
		for _, project := range projects {
			tasks, _ := taskSvc.ListByProject(project.ID)
			states, _ := stateSvc.ListByProject(project.ID)
			stateNames := make(map[uint]string)
			for _, s := range states {
				stateNames[s.ID] = s.Name
			}

			var mine []string
			for _, t := range tasks {
				if t.AssignedTo != nil && *t.AssignedTo == user.ID && t.CompletedAt == nil {
					due := ""
					if t.DueDate != nil {
						due = fmt.Sprintf(" (due %s)", t.DueDate.Format("Jan 2"))
					}
					mine = append(mine, fmt.Sprintf("  #%d  %-30s  %-15s%s",
						t.TaskNumber, t.Title, stateNames[t.StateID], due))
				}
			}

			if len(mine) > 0 {
				fmt.Printf("\n%s\n", project.Name)
				for _, l := range mine {
					fmt.Println(l)
				}
				total += len(mine)
			}
		}

		if total == 0 {
			fmt.Println("\nNo open tasks assigned to you.")
		} else {
			fmt.Printf("\n%d open task(s) total\n", total)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(myTasksCmd)
}
