package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/infrastructure/auth"
)

var sprintCmd = &cobra.Command{
	Use:   "sprint",
	Short: "Manage sprints",
}

var sprintCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new sprint",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		goal, _ := cmd.Flags().GetString("goal")
		projectID, _ := cmd.Flags().GetUint("project")
		startStr, _ := cmd.Flags().GetString("start")
		endStr, _ := cmd.Flags().GetString("end")

		if name == "" || projectID == 0 || startStr == "" || endStr == "" {
			return fmt.Errorf("--name, --project, --start and --end are required")
		}

		start, err := time.Parse("2006-01-02", startStr)
		if err != nil {
			return fmt.Errorf("invalid --start date, use YYYY-MM-DD")
		}
		end, err := time.Parse("2006-01-02", endStr)
		if err != nil {
			return fmt.Errorf("invalid --end date, use YYYY-MM-DD")
		}

		sprintSvc := app.NewSprintService(DB)
		sprint, err := sprintSvc.Create(name, goal, projectID, start, end)
		if err != nil {
			return err
		}

		fmt.Printf("✓ Sprint created: [%d] %s (%s → %s)\n",
			sprint.ID, sprint.Name,
			sprint.StartDate.Format("Jan 2"),
			sprint.EndDate.Format("Jan 2"),
		)
		return nil
	},
}

var sprintListCmd = &cobra.Command{
	Use:   "list",
	Short: "List sprints for a project",
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID, _ := cmd.Flags().GetUint("project")
		if projectID == 0 {
			return fmt.Errorf("--project is required")
		}

		sprintSvc := app.NewSprintService(DB)
		sprints, err := sprintSvc.ListByProject(projectID)
		if err != nil {
			return err
		}

		if len(sprints) == 0 {
			fmt.Println("No sprints yet.")
			return nil
		}

		now := time.Now()
		fmt.Printf("%-5s  %-25s  %-12s  %-12s  %s\n", "ID", "Name", "Start", "End", "Status")
		for _, sp := range sprints {
			status := "upcoming"
			if sp.CompletedAt != nil {
				status = "completed"
			} else if sp.StartDate.Before(now) && sp.EndDate.After(now) {
				status = "active"
			}
			fmt.Printf("%-5d  %-25s  %-12s  %-12s  %s\n",
				sp.ID, sp.Name,
				sp.StartDate.Format("2006-01-02"),
				sp.EndDate.Format("2006-01-02"),
				status,
			)
		}
		return nil
	},
}

var sprintCompleteCmd = &cobra.Command{
	Use:   "complete",
	Short: "Complete a sprint and move unfinished tasks to backlog",
	RunE: func(cmd *cobra.Command, args []string) error {
		sprintID, _ := cmd.Flags().GetUint("id")
		backlogStateID, _ := cmd.Flags().GetUint("backlog-state")
		if sprintID == 0 || backlogStateID == 0 {
			return fmt.Errorf("--id and --backlog-state are required")
		}

		sprintSvc := app.NewSprintService(DB)
		completed, total, _ := sprintSvc.Velocity(sprintID)

		if err := sprintSvc.Complete(sprintID, backlogStateID); err != nil {
			return err
		}

		moved := total - completed
		fmt.Printf("✓ Sprint completed\n")
		fmt.Printf("  Completed tasks: %d / %d\n", completed, total)
		if moved > 0 {
			fmt.Printf("  Tasks moved to backlog: %d\n", moved)
		}
		return nil
	},
}

var sprintVelocityCmd = &cobra.Command{
	Use:   "velocity",
	Short: "Show velocity and progress for a sprint",
	RunE: func(cmd *cobra.Command, args []string) error {
		sprintID, _ := cmd.Flags().GetUint("id")
		if sprintID == 0 {
			return fmt.Errorf("--id is required")
		}

		sprintSvc := app.NewSprintService(DB)
		sprint, err := sprintSvc.GetByID(sprintID)
		if err != nil {
			return err
		}

		completed, total, err := sprintSvc.Velocity(sprintID)
		if err != nil {
			return err
		}

		pct := 0
		if total > 0 {
			pct = (completed * 100) / total
		}

		bar := ""
		for i := 0; i < 20; i++ {
			if i < (pct / 5) {
				bar += "█"
			} else {
				bar += "░"
			}
		}

		fmt.Printf("\nSprint: %s\n", sprint.Name)
		fmt.Printf("Period: %s → %s\n", sprint.StartDate.Format("Jan 2"), sprint.EndDate.Format("Jan 2"))
		fmt.Printf("Progress: [%s] %d%% (%d/%d tasks)\n\n", bar, pct, completed, total)

		tasks, _ := sprintSvc.ListTasks(sprintID)
		if len(tasks) > 0 {
			fmt.Println("Tasks:")
			for _, t := range tasks {
				done := " "
				if t.CompletedAt != nil {
					done = "✓"
				}
				fmt.Printf("  [%s] #%d %s  (%s)\n", done, t.TaskNumber, t.Title, t.State.Name)
			}
		}
		return nil
	},
}

var sprintAssignCmd = &cobra.Command{
	Use:   "assign",
	Short: "Assign a task to a sprint",
	RunE: func(cmd *cobra.Command, args []string) error {
		sprintID, _ := cmd.Flags().GetUint("sprint")
		taskID, _ := cmd.Flags().GetUint("task")
		if sprintID == 0 || taskID == 0 {
			return fmt.Errorf("--sprint and --task are required")
		}
		sprintSvc := app.NewSprintService(DB)
		if err := sprintSvc.AddTask(sprintID, taskID); err != nil {
			return err
		}
		fmt.Printf("✓ Task %d assigned to sprint %d\n", taskID, sprintID)
		return nil
	},
}

var sprintSnapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Take a burndown snapshot for a sprint (run daily via cron)",
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID, _ := cmd.Flags().GetUint("project")
		sprintIDFlag, _ := cmd.Flags().GetString("id")

		sprintSvc := app.NewSprintService(DB)

		if sprintIDFlag != "" {
			id, err := strconv.ParseUint(sprintIDFlag, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid --id")
			}
			if err := sprintSvc.TakeBurndownSnapshot(uint(id)); err != nil {
				return err
			}
			fmt.Printf("✓ Burndown snapshot taken for sprint %d\n", id)
			return nil
		}

		if projectID == 0 {
			session, _ := auth.LoadSession()
			userSvc := app.NewUserService(DB)
			orgSvc := app.NewOrganizationService(DB)
			projectSvc := app.NewProjectService(DB)
			user, _ := userSvc.GetByEmail(session.Email)
			org, _ := orgSvc.GetByOwnerID(user.ID)
			projects, _ := projectSvc.ListByOrganization(org.ID)

			count := 0
			for _, p := range projects {
				sprint, err := sprintSvc.GetActive(p.ID)
				if err != nil {
					continue
				}
				sprintSvc.TakeBurndownSnapshot(sprint.ID)
				fmt.Printf("✓ Snapshot for sprint [%d] %s\n", sprint.ID, sprint.Name)
				count++
			}
			if count == 0 {
				fmt.Println("No active sprints found.")
			}
			return nil
		}

		sprint, err := sprintSvc.GetActive(projectID)
		if err != nil {
			return err
		}
		if err := sprintSvc.TakeBurndownSnapshot(sprint.ID); err != nil {
			return err
		}
		fmt.Printf("✓ Burndown snapshot taken for sprint [%d] %s\n", sprint.ID, sprint.Name)
		return nil
	},
}

func init() {
	sprintCreateCmd.Flags().String("name", "", "Sprint name")
	sprintCreateCmd.Flags().String("goal", "", "Sprint goal (optional)")
	sprintCreateCmd.Flags().Uint("project", 0, "Project ID")
	sprintCreateCmd.Flags().String("start", "", "Start date (YYYY-MM-DD)")
	sprintCreateCmd.Flags().String("end", "", "End date (YYYY-MM-DD)")

	sprintListCmd.Flags().Uint("project", 0, "Project ID")

	sprintCompleteCmd.Flags().Uint("id", 0, "Sprint ID")
	sprintCompleteCmd.Flags().Uint("backlog-state", 0, "State ID to move unfinished tasks to")

	sprintVelocityCmd.Flags().Uint("id", 0, "Sprint ID")

	sprintAssignCmd.Flags().Uint("sprint", 0, "Sprint ID")
	sprintAssignCmd.Flags().Uint("task", 0, "Task ID")

	sprintSnapshotCmd.Flags().Uint("project", 0, "Project ID (optional, defaults to all active sprints)")
	sprintSnapshotCmd.Flags().String("id", "", "Sprint ID (optional)")

	sprintCmd.AddCommand(sprintCreateCmd)
	sprintCmd.AddCommand(sprintListCmd)
	sprintCmd.AddCommand(sprintCompleteCmd)
	sprintCmd.AddCommand(sprintVelocityCmd)
	sprintCmd.AddCommand(sprintAssignCmd)
	sprintCmd.AddCommand(sprintSnapshotCmd)
	rootCmd.AddCommand(sprintCmd)
}
