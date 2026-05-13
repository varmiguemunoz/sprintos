package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/infrastructure/auth"
	"gorm.io/gorm"
)

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

func StartServer(db *gorm.DB) error {
	s := server.NewMCPServer("SprintOS", "1.0.0")

	projectSvc := app.NewProjectService(db)
	taskSvc := app.NewTaskService(db)
	stateSvc := app.NewStateService(db)
	userSvc := app.NewUserService(db)
	orgSvc := app.NewOrganizationService(db)

	session, err := auth.LoadSession()
	if err != nil {
		return fmt.Errorf("no active session — run `commandpm start` and log in first")
	}

	currentUser, err := userSvc.GetByEmail(session.Email)
	if err != nil {
		return fmt.Errorf("could not load user from session: %w", err)
	}

	org, err := orgSvc.GetByOwnerID(currentUser.ID)
	if err != nil {
		return fmt.Errorf("could not load organization: %w", err)
	}

	s.AddTool(
		mcp.NewTool("list_projects",
			mcp.WithDescription("List all projects in the organization"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			projects, err := projectSvc.ListByOrganization(org.ID)
			if err != nil {
				return mcp.NewToolResultText(fmt.Sprintf("Error: %s", err)), nil
			}
			out, _ := json.MarshalIndent(projects, "", "  ")
			return mcp.NewToolResultText(string(out)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("list_states",
			mcp.WithDescription("List states for a project"),
			mcp.WithNumber("project_id",
				mcp.Required(),
				mcp.Description("The project ID"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := getArgs(req)
			projectID := getUint(args, "project_id")
			states, err := stateSvc.ListByProject(projectID)
			if err != nil {
				return mcp.NewToolResultText(fmt.Sprintf("Error: %s", err)), nil
			}
			out, _ := json.MarshalIndent(states, "", "  ")
			return mcp.NewToolResultText(string(out)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("list_tasks",
			mcp.WithDescription("List all tasks in a project"),
			mcp.WithNumber("project_id",
				mcp.Required(),
				mcp.Description("The project ID"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := getArgs(req)
			projectID := getUint(args, "project_id")
			tasks, err := taskSvc.ListByProject(projectID)
			if err != nil {
				return mcp.NewToolResultText(fmt.Sprintf("Error: %s", err)), nil
			}
			out, _ := json.MarshalIndent(tasks, "", "  ")
			return mcp.NewToolResultText(string(out)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("list_members",
			mcp.WithDescription("List all members of the organization"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			teamSvc := app.NewTeamService(db)
			members, err := teamSvc.ListMembers(org.ID)
			if err != nil {
				return mcp.NewToolResultText(fmt.Sprintf("Error: %s", err)), nil
			}
			out, _ := json.MarshalIndent(members, "", "  ")
			return mcp.NewToolResultText(string(out)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("create_task",
			mcp.WithDescription("Create a new task in a project"),
			mcp.WithString("title",
				mcp.Required(),
				mcp.Description("Task title"),
			),
			mcp.WithString("description",
				mcp.Description("Task description"),
			),
			mcp.WithNumber("project_id",
				mcp.Required(),
				mcp.Description("Project ID"),
			),
			mcp.WithNumber("state_id",
				mcp.Required(),
				mcp.Description("State ID (column) for the task"),
			),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := getArgs(req)
			title := getString(args, "title")
			description := getString(args, "description")
			projectID := getUint(args, "project_id")
			stateID := getUint(args, "state_id")

			task, err := taskSvc.Create(
				title, description, stateID, projectID,
				currentUser.ID, nil, nil, nil,
			)
			if err != nil {
				return mcp.NewToolResultText(fmt.Sprintf("Error: %s", err)), nil
			}
			out, _ := json.MarshalIndent(task, "", "  ")
			return mcp.NewToolResultText(fmt.Sprintf("Task created:\n%s", string(out))), nil
		},
	)

	return server.ServeStdio(s)
}
