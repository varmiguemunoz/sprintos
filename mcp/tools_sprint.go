package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/varmiguemunoz/sprintos/internal/app"
)

func registerSprintTools(
	s *server.MCPServer,
	sprintSvc *app.SprintService,
) {
	s.AddTool(
		mcp.NewTool("list_sprints",
			mcp.WithDescription("List all sprints in a project"),
			mcp.WithNumber("project_id", mcp.Required(), mcp.Description("Project ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := getArgs(req)
			projectID := getUint(args, "project_id")
			sprints, err := sprintSvc.ListByProject(projectID)
			if err != nil {
				return mcp.NewToolResultText("Error: " + err.Error()), nil
			}
			if len(sprints) == 0 {
				return mcp.NewToolResultText("No sprints found for this project."), nil
			}
			return mcp.NewToolResultText(marshal(sprints)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("create_sprint",
			mcp.WithDescription("Create a new sprint in a project"),
			mcp.WithNumber("project_id", mcp.Required(), mcp.Description("Project ID")),
			mcp.WithString("name", mcp.Required(), mcp.Description("Sprint name (e.g. 'Sprint 1')")),
			mcp.WithString("start_date", mcp.Required(), mcp.Description("Start date in YYYY-MM-DD format")),
			mcp.WithString("end_date", mcp.Required(), mcp.Description("End date in YYYY-MM-DD format")),
			mcp.WithString("goal", mcp.Description("Sprint goal or description")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := getArgs(req)
			projectID := getUint(args, "project_id")
			name := getString(args, "name")
			if name == "" {
				return mcp.NewToolResultText("Error: name is required"), nil
			}
			startStr := getString(args, "start_date")
			endStr := getString(args, "end_date")

			startDate, err := time.Parse("2006-01-02", startStr)
			if err != nil {
				return mcp.NewToolResultText("Error: invalid start_date — use YYYY-MM-DD"), nil
			}
			endDate, err := time.Parse("2006-01-02", endStr)
			if err != nil {
				return mcp.NewToolResultText("Error: invalid end_date — use YYYY-MM-DD"), nil
			}
			if !endDate.After(startDate) {
				return mcp.NewToolResultText("Error: end_date must be after start_date"), nil
			}

			goal := getString(args, "goal")
			sprint, err := sprintSvc.Create(name, goal, projectID, startDate, endDate)
			if err != nil {
				return mcp.NewToolResultText("Error: " + err.Error()), nil
			}
			return mcp.NewToolResultText("Sprint created:\n" + marshal(sprint)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("edit_sprint",
			mcp.WithDescription("Edit a sprint's name, goal, or dates"),
			mcp.WithNumber("sprint_id", mcp.Required(), mcp.Description("Sprint ID")),
			mcp.WithString("name", mcp.Description("New sprint name")),
			mcp.WithString("goal", mcp.Description("New sprint goal")),
			mcp.WithString("start_date", mcp.Description("New start date in YYYY-MM-DD format")),
			mcp.WithString("end_date", mcp.Description("New end date in YYYY-MM-DD format")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := getArgs(req)
			sprintID := getUint(args, "sprint_id")

			current, err := sprintSvc.GetByID(sprintID)
			if err != nil {
				return mcp.NewToolResultText("Error: " + err.Error()), nil
			}

			name := getString(args, "name")
			if name == "" {
				name = current.Name
			}

			goal := getString(args, "goal")
			if goal == "" && current.Goal != nil {
				goal = *current.Goal
			}

			startDate := current.StartDate
			if s := getString(args, "start_date"); s != "" {
				parsed, err := time.Parse("2006-01-02", s)
				if err != nil {
					return mcp.NewToolResultText("Error: invalid start_date — use YYYY-MM-DD"), nil
				}
				startDate = parsed
			}

			endDate := current.EndDate
			if e := getString(args, "end_date"); e != "" {
				parsed, err := time.Parse("2006-01-02", e)
				if err != nil {
					return mcp.NewToolResultText("Error: invalid end_date — use YYYY-MM-DD"), nil
				}
				endDate = parsed
			}

			updated, err := sprintSvc.Update(sprintID, name, goal, startDate, endDate)
			if err != nil {
				return mcp.NewToolResultText("Error: " + err.Error()), nil
			}
			return mcp.NewToolResultText("Sprint updated:\n" + marshal(updated)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("delete_sprint",
			mcp.WithDescription("Delete a sprint. Tasks in the sprint are not deleted, only unlinked."),
			mcp.WithNumber("sprint_id", mcp.Required(), mcp.Description("Sprint ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := getArgs(req)
			sprintID := getUint(args, "sprint_id")
			if err := sprintSvc.Delete(sprintID); err != nil {
				return mcp.NewToolResultText("Error: " + err.Error()), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Sprint %d deleted.", sprintID)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("plan_sprint",
			mcp.WithDescription("Add a task to a sprint"),
			mcp.WithNumber("sprint_id", mcp.Required(), mcp.Description("Sprint ID")),
			mcp.WithNumber("task_id", mcp.Required(), mcp.Description("Task ID to add to the sprint")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := getArgs(req)
			sprintID := getUint(args, "sprint_id")
			taskID := getUint(args, "task_id")

			if _, err := sprintSvc.GetByID(sprintID); err != nil {
				return mcp.NewToolResultText("Error: sprint not found — " + err.Error()), nil
			}

			if err := sprintSvc.AddTask(sprintID, taskID); err != nil {
				return mcp.NewToolResultText("Error: " + err.Error()), nil
			}

			return mcp.NewToolResultText(fmt.Sprintf("✓ Task %d added to sprint %d.", taskID, sprintID)), nil
		},
	)
}
