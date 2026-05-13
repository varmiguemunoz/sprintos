package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/command_pm_app/internal/app"
	"github.com/varmiguemunoz/command_pm_app/internal/domain"
)

type TaskDetailModel struct {
	task       domain.Task
	comments   []domain.Comment
	loading    bool
	err        error
	project    domain.Project
	taskSvc    *app.TaskService
	commentSvc *app.CommentService
}

type TaskDetailLoadedMsg struct {
	Task     domain.Task
	Comments []domain.Comment
	Err      error
}

func NewTaskDetailModel(task domain.Task, project domain.Project, taskSvc *app.TaskService, commentSvc *app.CommentService) TaskDetailModel {
	return TaskDetailModel{
		task:       task,
		project:    project,
		loading:    true,
		taskSvc:    taskSvc,
		commentSvc: commentSvc,
	}
}

func (m TaskDetailModel) loadCmd() tea.Cmd {
	return func() tea.Msg {
		task, err := m.taskSvc.GetByID(m.task.ID)
		if err != nil {
			return TaskDetailLoadedMsg{Err: err}
		}
		comments, err := m.commentSvc.ListByTask(task.ID)
		if err != nil {
			return TaskDetailLoadedMsg{Err: err}
		}
		return TaskDetailLoadedMsg{Task: *task, Comments: comments}
	}
}

func (m TaskDetailModel) Init() tea.Cmd {
	return m.loadCmd()
}

func (m TaskDetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(TaskDetailLoadedMsg); ok {
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.task = msg.Task
		m.comments = msg.Comments
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "a":
			task := m.task
			project := m.project
			return m, func() tea.Msg {
				return NavigateMsg{To: screenAssignUser, Task: task, Project: project}
			}
		case "c":
			task := m.task
			project := m.project
			return m, func() tea.Msg {
				return NavigateMsg{To: screenCreateComment, Task: task, Project: project}
			}
		case "e":
			task := m.task
			project := m.project
			return m, func() tea.Msg {
				return NavigateMsg{To: screenEditTask, Task: task, Project: project}
			}
		case "esc":
			project := m.project
			return m, func() tea.Msg {
				return NavigateMsg{To: screenKanban, Project: project}
			}
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m TaskDetailModel) View() string {
	if m.loading {
		return titleStyle.Render("SprintOS — Task Detail") +
			"\n\n" + normalStyle.Render("Loading...") + "\n"
	}

	if m.err != nil {
		return titleStyle.Render("SprintOS — Task Detail") +
			"\n\n" + errorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())) + "\n"
	}

	s := titleStyle.Render(fmt.Sprintf("SprintOS — %s", m.task.Title)) + "\n\n"

	s += selectedStyle.Render("State: ") + normalStyle.Render(m.task.State.Name) + "\n"

	if m.task.Assignee != nil {
		s += selectedStyle.Render("Assigned to: ") + normalStyle.Render(m.task.Assignee.Name) + "\n"
	} else {
		s += selectedStyle.Render("Assigned to: ") + normalStyle.Render("Unassigned") + "\n"
	}

	if m.task.StartDate != nil {
		s += selectedStyle.Render("Start: ") + normalStyle.Render(m.task.StartDate.Format("2006-01-02")) + "\n"
	}

	if m.task.DueDate != nil {
		s += selectedStyle.Render("Due: ") + normalStyle.Render(m.task.DueDate.Format("2006-01-02")) + "\n"
	}

	if m.task.Description != nil && *m.task.Description != "" {
		s += "\n" + selectedStyle.Render("Description:") + "\n"
		s += normalStyle.Render(*m.task.Description) + "\n"
	}

	s += "\n" + selectedStyle.Render("Comments:") + "\n"
	if len(m.comments) == 0 {
		s += normalStyle.Render("  No comments yet.") + "\n"
	} else {
		for _, c := range m.comments {
			author := "Unknown"
			author = c.Author.Name
			s += normalStyle.Render(fmt.Sprintf("  [%s] %s: %s",
				c.CreatedAt.Format("2006-01-02 15:04"),
				author,
				c.Content,
			)) + "\n"
		}
	}

	s += "\n" + normalStyle.Render(strings.Repeat("─", 40)) + "\n"
	s += normalStyle.Render("a assign  •  c comment  •  e edit  •  esc back to board  •  q quit") + "\n"

	return s
}
