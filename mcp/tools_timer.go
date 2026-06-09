package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/domain"
)

func registerTimerTools(
	s *server.MCPServer,
	timeSvc *app.TimeEntryService,
	currentUser *domain.User,
) {
	s.AddTool(
		mcp.NewTool("start_timer",
			mcp.WithDescription("Start a timer on a task or subtask. Automatically stops any running timer first."),
			mcp.WithNumber("task_id", mcp.Description("Task ID to track time against")),
			mcp.WithNumber("subtask_id", mcp.Description("Subtask ID to track time against")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := getArgs(req)
			taskID := getUint(args, "task_id")
			subtaskID := getUint(args, "subtask_id")

			if taskID == 0 && subtaskID == 0 {
				return mcp.NewToolResultText("Error: provide task_id or subtask_id (or both)"), nil
			}

			var taskPtr *uint
			if taskID > 0 {
				taskPtr = &taskID
			}
			var subtaskPtr *uint
			if subtaskID > 0 {
				subtaskPtr = &subtaskID
			}

			timer, err := timeSvc.StartTimer(taskPtr, subtaskPtr, currentUser.ID)
			if err != nil {
				return mcp.NewToolResultText("Error: " + err.Error()), nil
			}

			msg := fmt.Sprintf("✓ Timer started at %s", timer.StartedAt.Format("15:04:05"))
			if taskPtr != nil {
				msg += fmt.Sprintf(" — task #%d", *taskPtr)
			}
			if subtaskPtr != nil {
				msg += fmt.Sprintf(" — subtask #%d", *subtaskPtr)
			}
			return mcp.NewToolResultText(msg), nil
		},
	)

	s.AddTool(
		mcp.NewTool("stop_timer",
			mcp.WithDescription("Stop the currently running timer and log the time entry"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			active, err := timeSvc.GetActiveTimer(currentUser.ID)
			if err != nil {
				return mcp.NewToolResultText("Error: " + err.Error()), nil
			}
			if active == nil {
				return mcp.NewToolResultText("No active timer running."), nil
			}

			entry, err := timeSvc.StopTimer(currentUser.ID)
			if err != nil {
				return mcp.NewToolResultText("Error: " + err.Error()), nil
			}
			if entry == nil {
				return mcp.NewToolResultText("Timer stopped (0 minutes logged)."), nil
			}

			return mcp.NewToolResultText(fmt.Sprintf(
				"✓ Timer stopped. Logged %s (%d minutes).",
				app.FormatMinutes(entry.Minutes), entry.Minutes,
			)), nil
		},
	)
}
