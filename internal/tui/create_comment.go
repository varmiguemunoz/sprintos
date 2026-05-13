package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/domain"
)

type CreateCommentModel struct {
	input      textinput.Model
	loading    bool
	err        error
	task       domain.Task
	project    domain.Project
	authorID   uint
	commentSvc *app.CommentService
}

type CommentCreatedMsg struct {
	Err error
}

func NewCreateCommentModel(
	task domain.Task,
	project domain.Project,
	authorID uint,
	commentSvc *app.CommentService,
) CreateCommentModel {
	input := textinput.New()
	input.Placeholder = "Write your comment..."
	input.CharLimit = 1000
	input.Focus()

	return CreateCommentModel{
		input:      input,
		task:       task,
		project:    project,
		authorID:   authorID,
		commentSvc: commentSvc,
	}
}

func (m CreateCommentModel) submitCmd() tea.Cmd {
	return func() tea.Msg {
		content := m.input.Value()
		if content == "" {
			return CommentCreatedMsg{Err: fmt.Errorf("comment cannot be empty")}
		}

		_, err := m.commentSvc.Create(content, m.task.ID, m.authorID)
		if err != nil {
			return CommentCreatedMsg{Err: err}
		}

		return CommentCreatedMsg{}
	}
}

func (m CreateCommentModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m CreateCommentModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(CommentCreatedMsg); ok {
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		task := m.task
		project := m.project
		return m, func() tea.Msg {
			return NavigateMsg{To: screenTaskDetail, Task: task, Project: project}
		}
	}

	if m.loading {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			task := m.task
			project := m.project
			return m, func() tea.Msg {
				return NavigateMsg{To: screenTaskDetail, Task: task, Project: project}
			}
		case "enter":
			m.loading = true
			return m, m.submitCmd()
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m CreateCommentModel) View() string {
	if m.loading {
		return titleStyle.Render("SprintOS — Add Comment") +
			"\n\n" + normalStyle.Render("Saving comment...") + "\n"
	}

	s := titleStyle.Render(fmt.Sprintf("SprintOS — Comment on: %s", m.task.Title)) + "\n\n"
	s += selectedStyle.Render("Comment") + "\n"
	s += m.input.View() + "\n\n"

	if m.err != nil {
		s += errorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())) + "\n\n"
	}

	s += normalStyle.Render("enter to submit  •  esc to cancel") + "\n"
	return s
}
