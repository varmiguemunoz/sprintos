package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/infrastructure/auth"
	"github.com/varmiguemunoz/sprintos/internal/infrastructure/email"
)

var reviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Run a board health review and optionally send a digest notification",
	RunE: func(cmd *cobra.Command, args []string) error {
		notify, _ := cmd.Flags().GetBool("notify")
		days, _ := cmd.Flags().GetInt("days")

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

		cutoff := time.Now().Add(-time.Duration(days) * 24 * time.Hour)
		now := time.Now()

		var report strings.Builder
		report.WriteString(fmt.Sprintf("SprintOS Board Review — %s\n", now.Format("Mon Jan 2 2006")))
		report.WriteString(strings.Repeat("─", 50) + "\n\n")

		totalStale, totalOverdue := 0, 0

		for _, project := range projects {
			tasks, _ := taskSvc.ListByProject(project.ID)
			states, _ := stateSvc.ListByProject(project.ID)
			stateNames := make(map[uint]string)
			for _, s := range states {
				stateNames[s.ID] = s.Name
			}

			var stale, overdueTasks []string
			for _, t := range tasks {
				if t.CompletedAt != nil {
					continue
				}
				if t.UpdatedAt.Before(cutoff) {
					d := int(now.Sub(t.UpdatedAt).Hours() / 24)
					stale = append(stale, fmt.Sprintf("  #%d %s (%s, %d days)", t.TaskNumber, t.Title, stateNames[t.StateID], d))
					totalStale++
				}
				if t.DueDate != nil && t.DueDate.Before(now) {
					overdueTasks = append(overdueTasks, fmt.Sprintf("  #%d %s (due %s)", t.TaskNumber, t.Title, t.DueDate.Format("Jan 2")))
					totalOverdue++
				}
			}

			if len(stale) > 0 || len(overdueTasks) > 0 {
				report.WriteString(fmt.Sprintf("Project: %s\n", project.Name))
				if len(stale) > 0 {
					report.WriteString(fmt.Sprintf("  Stale tasks (%d+ days without update):\n", days))
					for _, l := range stale {
						report.WriteString(l + "\n")
					}
				}
				if len(overdueTasks) > 0 {
					report.WriteString("  Overdue tasks:\n")
					for _, l := range overdueTasks {
						report.WriteString(l + "\n")
					}
				}
				report.WriteString("\n")
			}
		}

		report.WriteString(fmt.Sprintf("Summary: %d stale, %d overdue\n", totalStale, totalOverdue))

		fmt.Print(report.String())

		if notify && user.Email != "" {
			err := email.SendReview(user.Email, org.Name, report.String())
			if err != nil {
				fmt.Printf("Warning: could not send email notification: %s\n", err)
			} else {
				fmt.Printf("Digest sent to %s\n", user.Email)
			}
		}

		return nil
	},
}

func init() {
	reviewCmd.Flags().Bool("notify", false, "Send digest via email after review")
	reviewCmd.Flags().Int("days", 5, "Number of days to consider a task stale")
	rootCmd.AddCommand(reviewCmd)
}
