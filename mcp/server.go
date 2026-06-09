package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/infrastructure/auth"
	"gorm.io/gorm"
)

func StartServer(db *gorm.DB) error {
	s := server.NewMCPServer("SprintOS", "1.0.0")

	projectSvc := app.NewProjectService(db)
	taskSvc := app.NewTaskService(db)
	stateSvc := app.NewStateService(db)
	userSvc := app.NewUserService(db)
	orgSvc := app.NewOrganizationService(db)
	commentSvc := app.NewCommentService(db)
	teamSvc := app.NewTeamService(db)
	invitationSvc := app.NewInvitationService(db)
	subtaskSvc := app.NewSubtaskService(db)
	sprintSvc := app.NewSprintService(db)
	timeSvc := app.NewTimeEntryService(db)

	session, err := auth.LoadSession()
	if err != nil {
		return fmt.Errorf("no active session — run `sprintos start` and log in first")
	}

	currentUser, err := userSvc.GetByEmail(session.Email)
	if err != nil {
		return fmt.Errorf("could not load user from session: %w", err)
	}

	org, err := orgSvc.GetByOwnerID(currentUser.ID)
	if err != nil {
		return fmt.Errorf("could not load organization: %w", err)
	}

	marshal := func(v interface{}) string {
		b, _ := json.MarshalIndent(v, "", "  ")
		return string(b)
	}

	s.AddTool(
		mcp.NewTool("list_projects",
			mcp.WithDescription("List all projects in the organization"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			projects, err := projectSvc.ListByOrganization(org.ID)
			if err != nil {
				return mcp.NewToolResultText("Error: " + err.Error()), nil
			}
			return mcp.NewToolResultText(marshal(projects)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("create_project",
			mcp.WithDescription("Create a new project with states applied from a template"),
			mcp.WithString("name", mcp.Required(), mcp.Description("Project name")),
			mcp.WithString("description", mcp.Description("Project description")),
			mcp.WithString("template", mcp.Description("Board template: 'standard' (Backlog→In Progress→In Review→Done) or 'simple' (Todo→Done). Defaults to 'standard'")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := getArgs(req)
			name := getString(args, "name")
			description := getString(args, "description")
			template := getString(args, "template")
			if template == "" {
				template = "standard"
			}
			project, err := projectSvc.Create(name, description, org.ID, currentUser.ID)
			if err != nil {
				return mcp.NewToolResultText("Error: " + err.Error()), nil
			}
			if err := stateSvc.ApplyTemplate(project.ID, template); err != nil {
				return mcp.NewToolResultText("Project created but template failed: " + err.Error()), nil
			}
			states, _ := stateSvc.ListByProject(project.ID)
			return mcp.NewToolResultText(fmt.Sprintf("Project created:\n%s\n\nStates:\n%s", marshal(project), marshal(states))), nil
		},
	)

	s.AddTool(
		mcp.NewTool("list_states",
			mcp.WithDescription("List states (board columns) for a project"),
			mcp.WithNumber("project_id", mcp.Required(), mcp.Description("Project ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := getArgs(req)
			projectID := getUint(args, "project_id")
			states, err := stateSvc.ListByProject(projectID)
			if err != nil {
				return mcp.NewToolResultText("Error: " + err.Error()), nil
			}
			return mcp.NewToolResultText(marshal(states)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("list_tasks",
			mcp.WithDescription("List all tasks in a project including their task IDs (e.g. TSK-1)"),
			mcp.WithNumber("project_id", mcp.Required(), mcp.Description("Project ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := getArgs(req)
			projectID := getUint(args, "project_id")
			tasks, err := taskSvc.ListByProject(projectID)
			if err != nil {
				return mcp.NewToolResultText("Error: " + err.Error()), nil
			}
			return mcp.NewToolResultText(marshal(tasks)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("get_task_detail",
			mcp.WithDescription("Get full details of a task including all comments"),
			mcp.WithNumber("task_id", mcp.Required(), mcp.Description("Task ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := getArgs(req)
			taskID := getUint(args, "task_id")
			task, err := taskSvc.GetByID(taskID)
			if err != nil {
				return mcp.NewToolResultText("Error: " + err.Error()), nil
			}
			comments, _ := commentSvc.ListByTask(taskID)
			result := map[string]interface{}{
				"task":     task,
				"comments": comments,
			}
			return mcp.NewToolResultText(marshal(result)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("create_task",
			mcp.WithDescription("Create a new task in a project"),
			mcp.WithString("title", mcp.Required(), mcp.Description("Task title")),
			mcp.WithString("description", mcp.Description("Task description")),
			mcp.WithNumber("project_id", mcp.Required(), mcp.Description("Project ID")),
			mcp.WithNumber("state_id", mcp.Required(), mcp.Description("State (column) ID")),
			mcp.WithString("due_date", mcp.Description("Due date in YYYY-MM-DD format")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := getArgs(req)
			title := getString(args, "title")
			description := getString(args, "description")
			projectID := getUint(args, "project_id")
			stateID := getUint(args, "state_id")
			var dueDate *time.Time
			if d := getString(args, "due_date"); d != "" {
				t, err := time.Parse("2006-01-02", d)
				if err == nil {
					dueDate = &t
				}
			}
			task, err := taskSvc.Create(title, description, stateID, projectID, currentUser.ID, nil, nil, dueDate)
			if err != nil {
				return mcp.NewToolResultText("Error: " + err.Error()), nil
			}
			return mcp.NewToolResultText("Task created:\n" + marshal(task)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("update_task",
			mcp.WithDescription("Update a task's title, description or due date"),
			mcp.WithNumber("task_id", mcp.Required(), mcp.Description("Task ID")),
			mcp.WithString("title", mcp.Description("New title")),
			mcp.WithString("description", mcp.Description("New description")),
			mcp.WithString("due_date", mcp.Description("New due date in YYYY-MM-DD format")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := getArgs(req)
			taskID := getUint(args, "task_id")
			task, err := taskSvc.GetByID(taskID)
			if err != nil {
				return mcp.NewToolResultText("Error: " + err.Error()), nil
			}
			title := getString(args, "title")
			if title == "" {
				title = task.Title
			}
			description := getString(args, "description")
			if description == "" && task.Description != nil {
				description = *task.Description
			}
			var dueDate *time.Time
			if d := getString(args, "due_date"); d != "" {
				t, _ := time.Parse("2006-01-02", d)
				dueDate = &t
			}
			updated, err := taskSvc.Update(taskID, title, description, task.AssignedTo, task.StartDate, dueDate)
			if err != nil {
				return mcp.NewToolResultText("Error: " + err.Error()), nil
			}
			return mcp.NewToolResultText("Task updated:\n" + marshal(updated)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("delete_task",
			mcp.WithDescription("Delete a task permanently"),
			mcp.WithNumber("task_id", mcp.Required(), mcp.Description("Task ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := getArgs(req)
			taskID := getUint(args, "task_id")
			if err := taskSvc.Delete(taskID); err != nil {
				return mcp.NewToolResultText("Error: " + err.Error()), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Task %d deleted.", taskID)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("move_task",
			mcp.WithDescription("Move a task to a different state (column) on the board"),
			mcp.WithNumber("task_id", mcp.Required(), mcp.Description("Task ID")),
			mcp.WithNumber("state_id", mcp.Required(), mcp.Description("Target state ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := getArgs(req)
			taskID := getUint(args, "task_id")
			stateID := getUint(args, "state_id")
			task, err := taskSvc.MoveState(taskID, stateID)
			if err != nil {
				return mcp.NewToolResultText("Error: " + err.Error()), nil
			}
			return mcp.NewToolResultText("Task moved:\n" + marshal(task)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("assign_task",
			mcp.WithDescription("Assign a task to a team member"),
			mcp.WithNumber("task_id", mcp.Required(), mcp.Description("Task ID")),
			mcp.WithNumber("user_id", mcp.Description("User ID to assign. Pass 0 to unassign.")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := getArgs(req)
			taskID := getUint(args, "task_id")
			userID := getUint(args, "user_id")
			var assignTo *uint
			if userID > 0 {
				assignTo = &userID
			}
			task, err := taskSvc.AssignUser(taskID, assignTo)
			if err != nil {
				return mcp.NewToolResultText("Error: " + err.Error()), nil
			}
			return mcp.NewToolResultText("Task assigned:\n" + marshal(task)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("add_comment",
			mcp.WithDescription("Add a comment to a task"),
			mcp.WithNumber("task_id", mcp.Required(), mcp.Description("Task ID")),
			mcp.WithString("content", mcp.Required(), mcp.Description("Comment text")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := getArgs(req)
			taskID := getUint(args, "task_id")
			content := getString(args, "content")
			comment, err := commentSvc.Create(content, taskID, currentUser.ID)
			if err != nil {
				return mcp.NewToolResultText("Error: " + err.Error()), nil
			}
			return mcp.NewToolResultText("Comment added:\n" + marshal(comment)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("list_members",
			mcp.WithDescription("List all members of the organization"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			members, err := teamSvc.ListMembers(org.ID)
			if err != nil {
				return mcp.NewToolResultText("Error: " + err.Error()), nil
			}
			return mcp.NewToolResultText(marshal(members)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("list_overdue_tasks",
			mcp.WithDescription("List all tasks that are past their due date and not yet completed"),
			mcp.WithNumber("project_id", mcp.Description("Filter by project ID (optional)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := getArgs(req)
			projectID := getUint(args, "project_id")
			now := time.Now()

			var projects []uint
			if projectID > 0 {
				projects = []uint{projectID}
			} else {
				ps, _ := projectSvc.ListByOrganization(org.ID)
				for _, p := range ps {
					projects = append(projects, p.ID)
				}
			}

			type overdueTask struct {
				TaskID      uint      `json:"task_id"`
				TaskNumber  int       `json:"task_number"`
				Title       string    `json:"title"`
				ProjectID   uint      `json:"project_id"`
				DueDate     time.Time `json:"due_date"`
				DaysOverdue int       `json:"days_overdue"`
			}

			var overdue []overdueTask
			for _, pid := range projects {
				tasks, _ := taskSvc.ListByProject(pid)
				for _, t := range tasks {
					if t.DueDate != nil && t.DueDate.Before(now) && t.CompletedAt == nil {
						days := int(now.Sub(*t.DueDate).Hours() / 24)
						overdue = append(overdue, overdueTask{
							TaskID:      t.ID,
							TaskNumber:  t.TaskNumber,
							Title:       t.Title,
							ProjectID:   t.ProjectID,
							DueDate:     *t.DueDate,
							DaysOverdue: days,
						})
					}
				}
			}

			if len(overdue) == 0 {
				return mcp.NewToolResultText("No overdue tasks. Great job!"), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("%d overdue task(s):\n%s", len(overdue), marshal(overdue))), nil
		},
	)

	s.AddTool(
		mcp.NewTool("analyze_stale_tasks",
			mcp.WithDescription("Find tasks that have been sitting in the same state without updates for too long, and suggest actions"),
			mcp.WithNumber("project_id", mcp.Required(), mcp.Description("Project ID")),
			mcp.WithNumber("days", mcp.Description("Consider a task stale if not updated in this many days (default: 5)")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := getArgs(req)
			projectID := getUint(args, "project_id")
			days := getUint(args, "days")
			if days == 0 {
				days = 5
			}

			cutoff := time.Now().Add(-time.Duration(days) * 24 * time.Hour)
			tasks, err := taskSvc.ListByProject(projectID)
			if err != nil {
				return mcp.NewToolResultText("Error: " + err.Error()), nil
			}
			states, _ := stateSvc.ListByProject(projectID)
			stateNames := make(map[uint]string)
			for _, st := range states {
				stateNames[st.ID] = st.Name
			}

			type staleTask struct {
				TaskID            uint   `json:"task_id"`
				TaskNumber        int    `json:"task_number"`
				Title             string `json:"title"`
				CurrentState      string `json:"current_state"`
				DaysWithoutUpdate int    `json:"days_without_update"`
				SuggestedAction   string `json:"suggested_action"`
			}

			var stale []staleTask
			for _, t := range tasks {
				if t.UpdatedAt.Before(cutoff) && t.CompletedAt == nil {
					daysStale := int(time.Since(t.UpdatedAt).Hours() / 24)
					stateName := stateNames[t.StateID]
					action := "Follow up with assignee or move to backlog"
					if t.AssignedTo == nil {
						action = "Assign this task to a team member"
					} else if stateName == "In Review" {
						action = "Review is overdue — merge or request changes"
					} else if stateName == "In Progress" {
						action = "Check if there are blockers"
					}
					stale = append(stale, staleTask{
						TaskID:            t.ID,
						TaskNumber:        t.TaskNumber,
						Title:             t.Title,
						CurrentState:      stateName,
						DaysWithoutUpdate: daysStale,
						SuggestedAction:   action,
					})
				}
			}

			if len(stale) == 0 {
				return mcp.NewToolResultText(fmt.Sprintf("No stale tasks (threshold: %d days). Board is healthy!", days)), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("%d stale task(s) found:\n%s", len(stale), marshal(stale))), nil
		},
	)

	s.AddTool(
		mcp.NewTool("summarize_project",
			mcp.WithDescription("Get a structured health summary of a project: task counts by state, overdue, unassigned, and team workload"),
			mcp.WithNumber("project_id", mcp.Required(), mcp.Description("Project ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := getArgs(req)
			projectID := getUint(args, "project_id")

			tasks, err := taskSvc.ListByProject(projectID)
			if err != nil {
				return mcp.NewToolResultText("Error: " + err.Error()), nil
			}
			states, _ := stateSvc.ListByProject(projectID)
			stateNames := make(map[uint]string)
			for _, st := range states {
				stateNames[st.ID] = st.Name
			}

			byState := make(map[string]int)
			var overdue, unassigned, completed int
			workload := make(map[uint]int)
			now := time.Now()

			for _, t := range tasks {
				byState[stateNames[t.StateID]]++
				if t.CompletedAt != nil {
					completed++
				}
				if t.AssignedTo == nil {
					unassigned++
				} else {
					workload[*t.AssignedTo]++
				}
				if t.DueDate != nil && t.DueDate.Before(now) && t.CompletedAt == nil {
					overdue++
				}
			}

			summary := map[string]interface{}{
				"total_tasks":   len(tasks),
				"completed":     completed,
				"overdue":       overdue,
				"unassigned":    unassigned,
				"by_state":      byState,
				"team_workload": workload,
			}
			return mcp.NewToolResultText(marshal(summary)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("generate_sprint",
			mcp.WithDescription("Create multiple tasks at once from a JSON list. The AI should analyze a PRD or description and call this with the resulting task list."),
			mcp.WithNumber("project_id", mcp.Required(), mcp.Description("Project ID")),
			mcp.WithNumber("default_state_id", mcp.Required(), mcp.Description("State ID where tasks land by default (usually Backlog)")),
			mcp.WithString("tasks_json", mcp.Required(), mcp.Description(`JSON array of tasks. Each object: {"title":"...","description":"...","state_id":1}. state_id is optional per task.`)),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := getArgs(req)
			projectID := getUint(args, "project_id")
			defaultStateID := getUint(args, "default_state_id")
			tasksJSON := getString(args, "tasks_json")

			var defs []struct {
				Title       string `json:"title"`
				Description string `json:"description"`
				StateID     uint   `json:"state_id"`
			}
			if err := json.Unmarshal([]byte(tasksJSON), &defs); err != nil {
				return mcp.NewToolResultText("Invalid tasks_json: " + err.Error()), nil
			}

			var created []string
			for _, d := range defs {
				stateID := d.StateID
				if stateID == 0 {
					stateID = defaultStateID
				}
				task, err := taskSvc.Create(d.Title, d.Description, stateID, projectID, currentUser.ID, nil, nil, nil)
				if err != nil {
					created = append(created, fmt.Sprintf("FAILED %q: %s", d.Title, err))
					continue
				}
				created = append(created, fmt.Sprintf("✓ #%d %s", task.TaskNumber, task.Title))
			}

			return mcp.NewToolResultText(fmt.Sprintf("Sprint generated — %d task(s) created:\n%s",
				len(defs), joinLines(created))), nil
		},
	)

	s.AddTool(
		mcp.NewTool("list_organizations",
			mcp.WithDescription("List all organizations the current user belongs to (owned and member)"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			result := []interface{}{}
			if ownerOrg, err := orgSvc.GetByOwnerID(currentUser.ID); err == nil {
				result = append(result, map[string]interface{}{
					"id":   ownerOrg.ID,
					"name": ownerOrg.Name,
					"role": "owner",
				})
			}
			memberOrgs, _ := teamSvc.GetOrganizationsByMemberUserID(currentUser.ID)
			for _, o := range memberOrgs {
				result = append(result, map[string]interface{}{
					"id":   o.ID,
					"name": o.Name,
					"role": "member",
				})
			}
			return mcp.NewToolResultText(marshal(result)), nil
		},
	)

	registerOrgTools(s, org, currentUser, teamSvc, invitationSvc)
	registerSubtaskTools(s, subtaskSvc, currentUser)
	registerTimerTools(s, timeSvc, currentUser)
	registerSprintTools(s, sprintSvc)

	return server.ServeStdio(s)
}

func getArgs(req mcp.CallToolRequest) map[string]interface{} {
	args, _ := req.Params.Arguments.(map[string]interface{})
	return args
}

func getUint(args map[string]interface{}, key string) uint {
	v, _ := args[key].(float64)
	return uint(v)
}

func getString(args map[string]interface{}, key string) string {
	v, _ := args[key].(string)
	return v
}

func joinLines(lines []string) string {
	result := ""
	for _, l := range lines {
		result += l + "\n"
	}
	return result
}
