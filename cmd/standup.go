package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/infrastructure/auth"
)

var standupCmd = &cobra.Command{
	Use:   "standup",
	Short: "Generate a standup update based on recent task activity",
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

		projects, err := projectSvc.ListByOrganization(org.ID)
		if err != nil {
			return fmt.Errorf("could not load projects: %w", err)
		}

		yesterday := time.Now().Add(-24 * time.Hour)
		now := time.Now()

		var done []string
		var inProgress []string
		var overdue []string

		for _, project := range projects {
			tasks, err := taskSvc.ListByProject(project.ID)
			if err != nil {
				continue
			}
			states, _ := stateSvc.ListByProject(project.ID)
			stateNames := make(map[uint]string)
			isDoneState := make(map[uint]bool)
			for _, s := range states {
				stateNames[s.ID] = s.Name
				isDoneState[s.ID] = s.IsDone
			}

			for _, t := range tasks {
				if t.AssignedTo == nil || *t.AssignedTo != user.ID {
					continue
				}
				if isDoneState[t.StateID] && t.CompletedAt != nil && t.CompletedAt.After(yesterday) {
					done = append(done, fmt.Sprintf("  ✓ [%s] #%d %s", project.Name, t.TaskNumber, t.Title))
				} else if !isDoneState[t.StateID] && stateNames[t.StateID] == "In Progress" {
					inProgress = append(inProgress, fmt.Sprintf("  → [%s] #%d %s", project.Name, t.TaskNumber, t.Title))
				}
				if t.DueDate != nil && t.DueDate.Before(now) && t.CompletedAt == nil {
					overdue = append(overdue, fmt.Sprintf("  ⚠ [%s] #%d %s (due %s)", project.Name, t.TaskNumber, t.Title, t.DueDate.Format("Jan 2")))
				}
			}
		}

		fmt.Printf("\n── Standup — %s ──────────────────────\n\n", time.Now().Format("Monday Jan 2"))

		fmt.Println("Yesterday:")
		if len(done) == 0 {
			fmt.Println("  nothing completed")
		} else {
			for _, l := range done {
				fmt.Println(l)
			}
		}

		fmt.Println("\nToday:")
		if len(inProgress) == 0 {
			fmt.Println("  nothing in progress")
		} else {
			for _, l := range inProgress {
				fmt.Println(l)
			}
		}

		fmt.Println("\nBlockers:")
		if len(overdue) == 0 {
			fmt.Println("  none")
		} else {
			for _, l := range overdue {
				fmt.Println(l)
			}
		}

		fmt.Println()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(standupCmd)
}
