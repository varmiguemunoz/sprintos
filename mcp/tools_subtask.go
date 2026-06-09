package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/domain"
)

func registerSubtaskTools(
	s *server.MCPServer,
	subtaskSvc *app.SubtaskService,
	currentUser *domain.User,
) {
	s.AddTool(
		mcp.NewTool("list_subtasks",
			mcp.WithDescription("List all subtasks inside a task"),
			mcp.WithNumber("task_id", mcp.Required(), mcp.Description("Task ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := getArgs(req)
			taskID := getUint(args, "task_id")
			subtasks, err := subtaskSvc.ListByTask(taskID)
			if err != nil {
				return mcp.NewToolResultText("Error: " + err.Error()), nil
			}
			if len(subtasks) == 0 {
				return mcp.NewToolResultText("No subtasks found for this task."), nil
			}
			return mcp.NewToolResultText(marshal(subtasks)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("create_subtask",
			mcp.WithDescription("Create a subtask inside a task"),
			mcp.WithNumber("task_id", mcp.Required(), mcp.Description("Parent task ID")),
			mcp.WithString("title", mcp.Required(), mcp.Description("Subtask title")),
			mcp.WithString("description", mcp.Description("Subtask description")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := getArgs(req)
			taskID := getUint(args, "task_id")
			title := getString(args, "title")
			if title == "" {
				return mcp.NewToolResultText("Error: title is required"), nil
			}
			description := getString(args, "description")
			subtask, err := subtaskSvc.Create(title, description, taskID, currentUser.ID)
			if err != nil {
				return mcp.NewToolResultText("Error: " + err.Error()), nil
			}
			return mcp.NewToolResultText("Subtask created:\n" + marshal(subtask)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("edit_subtask",
			mcp.WithDescription("Edit a subtask's title, description, or done status"),
			mcp.WithNumber("subtask_id", mcp.Required(), mcp.Description("Subtask ID")),
			mcp.WithString("title", mcp.Description("New title")),
			mcp.WithString("description", mcp.Description("New description")),
			mcp.WithString("done", mcp.Description("Set done status: 'true' or 'false'")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := getArgs(req)
			subtaskID := getUint(args, "subtask_id")

			current, err := subtaskSvc.GetByID(subtaskID)
			if err != nil {
				return mcp.NewToolResultText("Error: " + err.Error()), nil
			}

			title := getString(args, "title")
			if title == "" {
				title = current.Title
			}
			description := getString(args, "description")
			if description == "" && current.Description != nil {
				description = *current.Description
			}

			updated, err := subtaskSvc.Update(subtaskID, title, description)
			if err != nil {
				return mcp.NewToolResultText("Error: " + err.Error()), nil
			}

			doneStr := getString(args, "done")
			if doneStr == "true" && !current.Done {
				updated, err = subtaskSvc.ToggleDone(subtaskID)
				if err != nil {
					return mcp.NewToolResultText("Updated but could not toggle done: " + err.Error()), nil
				}
			} else if doneStr == "false" && current.Done {
				updated, err = subtaskSvc.ToggleDone(subtaskID)
				if err != nil {
					return mcp.NewToolResultText("Updated but could not toggle done: " + err.Error()), nil
				}
			}

			return mcp.NewToolResultText("Subtask updated:\n" + marshal(updated)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("delete_subtask",
			mcp.WithDescription("Delete a subtask permanently"),
			mcp.WithNumber("subtask_id", mcp.Required(), mcp.Description("Subtask ID")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := getArgs(req)
			subtaskID := getUint(args, "subtask_id")
			if err := subtaskSvc.Delete(subtaskID); err != nil {
				return mcp.NewToolResultText("Error: " + err.Error()), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Subtask %d deleted.", subtaskID)), nil
		},
	)
}
