package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/infrastructure/auth"
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks from the command line",
}

var taskCreateCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]
		projectFlag, _ := cmd.Flags().GetUint("project")
		stateFlag, _ := cmd.Flags().GetString("state")
		priority, _ := cmd.Flags().GetString("priority")
		format, _ := cmd.Flags().GetString("format")

		session, err := auth.LoadSession()
		if err != nil {
			return fmt.Errorf("not logged in — run `sprintos start` first")
		}

		userSvc := app.NewUserService(DB)
		stateSvc := app.NewStateService(DB)
		taskSvc := app.NewTaskService(DB)

		user, err := userSvc.GetByEmail(session.Email)
		if err != nil {
			return fmt.Errorf("could not load user: %w", err)
		}

		projectID, err := app.ResolveProjectID(projectFlag, stateSvc)
		if err != nil {
			return err
		}

		if stateFlag == "" {
			stateFlag = "backlog"
		}
		stateID, err := app.ResolveStateByName(stateSvc, projectID, stateFlag)
		if err != nil {
			return err
		}

		_ = priority

		task, err := taskSvc.Create(title, "", stateID, projectID, user.ID, nil, nil, nil)
		if err != nil {
			return fmt.Errorf("could not create task: %w", err)
		}

		if format == "json" {
			return json.NewEncoder(os.Stdout).Encode(task)
		}

		fmt.Printf("✓ Task created: #%d %s\n", task.TaskNumber, task.Title)
		return nil
	},
}

var taskMoveCmd = &cobra.Command{
	Use:   "move [task-number] [state]",
	Short: "Move a task to a different state",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskNumber, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid task number: %s", args[0])
		}
		stateName := args[1]

		projectFlag, _ := cmd.Flags().GetUint("project")
		stateSvc := app.NewStateService(DB)
		taskSvc := app.NewTaskService(DB)

		projectID, err := app.ResolveProjectID(projectFlag, stateSvc)
		if err != nil {
			return err
		}

		stateID, err := app.ResolveStateByName(stateSvc, projectID, stateName)
		if err != nil {
			return err
		}

		task, err := taskSvc.GetByTaskNumber(taskNumber, projectID)
		if err != nil {
			return fmt.Errorf("task #%d not found: %w", taskNumber, err)
		}

		if _, err := taskSvc.MoveState(task.ID, stateID); err != nil {
			return fmt.Errorf("could not move task: %w", err)
		}

		fmt.Printf("✓ Task #%d moved to %q\n", taskNumber, stateName)
		return nil
	},
}

var taskLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		projectFlag, _ := cmd.Flags().GetUint("project")
		stateFlag, _ := cmd.Flags().GetString("state")
		format, _ := cmd.Flags().GetString("format")

		stateSvc := app.NewStateService(DB)
		taskSvc := app.NewTaskService(DB)

		projectID, err := app.ResolveProjectID(projectFlag, stateSvc)
		if err != nil {
			return err
		}

		tasks, err := taskSvc.ListByProject(projectID)
		if err != nil {
			return fmt.Errorf("could not list tasks: %w", err)
		}

		if stateFlag != "" {
			stateID, err := app.ResolveStateByName(stateSvc, projectID, stateFlag)
			if err != nil {
				return err
			}
			var filtered []interface{}
			for _, t := range tasks {
				if t.StateID == stateID {
					filtered = append(filtered, t)
				}
			}
			if format == "json" {
				return json.NewEncoder(os.Stdout).Encode(filtered)
			}
		}

		if format == "json" {
			return json.NewEncoder(os.Stdout).Encode(tasks)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tTITLE\tSTATE\tPRIORITY\tASSIGNED")
		for _, t := range tasks {
			if stateFlag != "" {
				stateID, _ := app.ResolveStateByName(stateSvc, projectID, stateFlag)
				if t.StateID != stateID {
					continue
				}
			}
			assigned := "-"
			if t.Assignee != nil {
				assigned = t.Assignee.Name
			}
			fmt.Fprintf(w, "#%d\t%s\t%s\t%s\t%s\n",
				t.TaskNumber, t.Title, t.State.Name, t.Priority, assigned)
		}
		return w.Flush()
	},
}

var taskShowCmd = &cobra.Command{
	Use:   "show [task-number]",
	Short: "Show full task detail",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskNumber, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid task number: %s", args[0])
		}

		projectFlag, _ := cmd.Flags().GetUint("project")
		format, _ := cmd.Flags().GetString("format")
		stateSvc := app.NewStateService(DB)
		taskSvc := app.NewTaskService(DB)
		commentSvc := app.NewCommentService(DB)

		projectID, err := app.ResolveProjectID(projectFlag, stateSvc)
		if err != nil {
			return err
		}

		task, err := taskSvc.GetByTaskNumber(taskNumber, projectID)
		if err != nil {
			return fmt.Errorf("task #%d not found", taskNumber)
		}

		comments, _ := commentSvc.ListByTask(task.ID)

		if format == "json" {
			return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
				"task":     task,
				"comments": comments,
			})
		}

		fmt.Printf("\n#%d %s\n", task.TaskNumber, task.Title)
		fmt.Printf("State:    %s\n", task.State.Name)
		fmt.Printf("Priority: %s\n", task.Priority)
		if task.Assignee != nil {
			fmt.Printf("Assigned: %s\n", task.Assignee.Name)
		}
		if task.DueDate != nil {
			fmt.Printf("Due:      %s\n", task.DueDate.Format("2006-01-02"))
		}
		if task.Description != nil && *task.Description != "" {
			fmt.Printf("\n%s\n", *task.Description)
		}
		if len(comments) > 0 {
			fmt.Printf("\nComments (%d):\n", len(comments))
			for _, c := range comments {
				fmt.Printf("  [%s] %s: %s\n", c.CreatedAt.Format("Jan 2 15:04"), c.Author.Name, c.Content)
			}
		}
		fmt.Println()
		return nil
	},
}

func init() {
	taskCreateCmd.Flags().Uint("project", 0, "Project ID (or read from .sprintos)")
	taskCreateCmd.Flags().String("state", "backlog", "State name or ID")
	taskCreateCmd.Flags().String("priority", "medium", "Priority: low, medium, high, critical")
	taskCreateCmd.Flags().String("format", "table", "Output format: table, json")

	taskMoveCmd.Flags().Uint("project", 0, "Project ID (or read from .sprintos)")

	taskLsCmd.Flags().Uint("project", 0, "Project ID (or read from .sprintos)")
	taskLsCmd.Flags().String("state", "", "Filter by state name or ID")
	taskLsCmd.Flags().String("format", "table", "Output format: table, json")

	taskShowCmd.Flags().Uint("project", 0, "Project ID (or read from .sprintos)")
	taskShowCmd.Flags().String("format", "table", "Output format: table, json")

	taskCmd.AddCommand(taskCreateCmd)
	taskCmd.AddCommand(taskMoveCmd)
	taskCmd.AddCommand(taskLsCmd)
	taskCmd.AddCommand(taskShowCmd)
	rootCmd.AddCommand(taskCmd)
}


