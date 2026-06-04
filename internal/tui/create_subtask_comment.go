package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/domain"
)

type CreateSubtaskCommentModel struct {
	input              textinput.Model
	loading            bool
	err                error
	subtask            domain.Subtask
	task               domain.Task
	project            domain.Project
	authorID           uint
	subtaskCommentSvc  *app.SubtaskCommentService
}

type SubtaskCommentCreatedMsg struct {
	Err error
}

func NewCreateSubtaskCommentModel(
	subtask domain.Subtask,
	task domain.Task,
	project domain.Project,
	authorID uint,
	subtaskCommentSvc *app.SubtaskCommentService,
) CreateSubtaskCommentModel {
	input := textinput.New()
	input.Placeholder = "Write your comment..."
	input.CharLimit = 1000
	input.Focus()

	return CreateSubtaskCommentModel{
		input:             input,
		subtask:           subtask,
		task:              task,
		project:           project,
		authorID:          authorID,
		subtaskCommentSvc: subtaskCommentSvc,
	}
}

func (m CreateSubtaskCommentModel) submitCmd() tea.Cmd {
	return func() tea.Msg {
		content := m.input.Value()
		if content == "" {
			return SubtaskCommentCreatedMsg{Err: fmt.Errorf("comment cannot be empty")}
		}

		_, err := m.subtaskCommentSvc.Create(content, m.subtask.ID, m.authorID)
		if err != nil {
			return SubtaskCommentCreatedMsg{Err: err}
		}

		return SubtaskCommentCreatedMsg{}
	}
}

func (m CreateSubtaskCommentModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m CreateSubtaskCommentModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(SubtaskCommentCreatedMsg); ok {
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		subtask := m.subtask
		task := m.task
		project := m.project
		return m, func() tea.Msg {
			return NavigateMsg{To: screenSubtaskDetail, Subtask: subtask, Task: task, Project: project}
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
			subtask := m.subtask
			task := m.task
			project := m.project
			return m, func() tea.Msg {
				return NavigateMsg{To: screenSubtaskDetail, Subtask: subtask, Task: task, Project: project}
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

func (m CreateSubtaskCommentModel) View() string {
	if m.loading {
		return titleStyle.Render("SprintOS — Add Comment") +
			"\n\n" + normalStyle.Render("Saving comment...") + "\n"
	}

	s := titleStyle.Render(fmt.Sprintf("SprintOS — Comment on: %s", m.subtask.Title)) + "\n\n"
	s += selectedStyle.Render("Comment") + "\n"
	s += m.input.View() + "\n\n"

	if m.err != nil {
		s += errorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())) + "\n\n"
	}

	s += normalStyle.Render("enter to submit  •  esc to cancel") + "\n"
	return s
}
