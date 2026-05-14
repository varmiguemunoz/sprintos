package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/infrastructure/auth"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export tasks to CSV or JSON",
	RunE: func(cmd *cobra.Command, args []string) error {
		projectFlag, _ := cmd.Flags().GetUint("project")
		format, _ := cmd.Flags().GetString("format")
		output, _ := cmd.Flags().GetString("output")

		stateSvc := app.NewStateService(DB)
		taskSvc := app.NewTaskService(DB)

		session, err := auth.LoadSession()
		if err != nil {
			return fmt.Errorf("not logged in — run `sprintos start` first")
		}

		userSvc := app.NewUserService(DB)
		orgSvc := app.NewOrganizationService(DB)
		user, _ := userSvc.GetByEmail(session.Email)
		org, _ := orgSvc.GetByOwnerID(user.ID)
		_ = org

		projectID, err := app.ResolveProjectID(projectFlag, stateSvc)
		if err != nil {
			return err
		}

		tasks, err := taskSvc.ListByProject(projectID)
		if err != nil {
			return fmt.Errorf("could not load tasks: %w", err)
		}

		var out *os.File
		if output != "" {
			out, err = os.Create(output)
			if err != nil {
				return fmt.Errorf("could not create file: %w", err)
			}
			defer out.Close()
		} else {
			out = os.Stdout
		}

		switch format {
		case "json":
			enc := json.NewEncoder(out)
			enc.SetIndent("", "  ")
			if err := enc.Encode(tasks); err != nil {
				return fmt.Errorf("could not encode JSON: %w", err)
			}

		case "csv":
			w := csv.NewWriter(out)
			w.Write([]string{
				"id", "task_number", "title", "description",
				"state", "priority", "assigned_to",
				"start_date", "due_date", "completed_at", "created_at",
			})
			for _, t := range tasks {
				desc := ""
				if t.Description != nil {
					desc = *t.Description
				}
				assignee := ""
				if t.Assignee != nil {
					assignee = t.Assignee.Name
				}
				startDate, dueDate, completedAt := "", "", ""
				if t.StartDate != nil {
					startDate = t.StartDate.Format("2006-01-02")
				}
				if t.DueDate != nil {
					dueDate = t.DueDate.Format("2006-01-02")
				}
				if t.CompletedAt != nil {
					completedAt = t.CompletedAt.Format("2006-01-02")
				}
				w.Write([]string{
					fmt.Sprintf("%d", t.ID),
					fmt.Sprintf("%d", t.TaskNumber),
					t.Title,
					desc,
					t.State.Name,
					t.Priority,
					assignee,
					startDate,
					dueDate,
					completedAt,
					t.CreatedAt.Format("2006-01-02"),
				})
			}
			w.Flush()
			if err := w.Error(); err != nil {
				return fmt.Errorf("CSV write error: %w", err)
			}

		default:
			return fmt.Errorf("unknown format %q — use csv or json", format)
		}

		if output != "" {
			fmt.Fprintf(os.Stderr, "✓ Exported %d tasks to %s\n", len(tasks), output)
		}
		return nil
	},
}

func init() {
	exportCmd.Flags().Uint("project", 0, "Project ID (or read from .sprintos)")
	exportCmd.Flags().String("format", "json", "Export format: csv, json")
	exportCmd.Flags().String("output", "", "Output file (default: stdout)")
	rootCmd.AddCommand(exportCmd)
}
