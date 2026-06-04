package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/domain"
)

type EditSubtaskCommentModel struct {
	input             textinput.Model
	loading           bool
	saved             bool
	err               error
	comment           domain.SubtaskComment
	subtask           domain.Subtask
	task              domain.Task
	project           domain.Project
	subtaskCommentSvc *app.SubtaskCommentService
}

type SubtaskCommentUpdatedMsg struct {
	Err error
}

func NewEditSubtaskCommentModel(
	comment domain.SubtaskComment,
	subtask domain.Subtask,
	task domain.Task,
	project domain.Project,
	subtaskCommentSvc *app.SubtaskCommentService,
) EditSubtaskCommentModel {
	input := textinput.New()
	input.Placeholder = "Edit your comment..."
	input.CharLimit = 1000
	input.SetValue(comment.Content)
	input.Focus()

	return EditSubtaskCommentModel{
		input:             input,
		comment:           comment,
		subtask:           subtask,
		task:              task,
		project:           project,
		subtaskCommentSvc: subtaskCommentSvc,
	}
}

func (m EditSubtaskCommentModel) saveCmd() tea.Cmd {
	return func() tea.Msg {
		content := m.input.Value()
		if content == "" {
			return SubtaskCommentUpdatedMsg{Err: fmt.Errorf("comment cannot be empty")}
		}

		_, err := m.subtaskCommentSvc.Update(m.comment.ID, content)
		if err != nil {
			return SubtaskCommentUpdatedMsg{Err: err}
		}

		return SubtaskCommentUpdatedMsg{}
	}
}

func (m EditSubtaskCommentModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m EditSubtaskCommentModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(SubtaskCommentUpdatedMsg); ok {
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.saved = true
		m.err = nil
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
			return m, m.saveCmd()
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m EditSubtaskCommentModel) View() string {
	if m.loading {
		return titleStyle.Render("SprintOS — Edit Comment") +
			"\n\n" + normalStyle.Render("Saving...") + "\n"
	}

	s := titleStyle.Render(fmt.Sprintf("SprintOS — Edit Comment on: %s", m.subtask.Title)) + "\n\n"
	s += selectedStyle.Render("Comment") + "\n"
	s += m.input.View() + "\n\n"

	if m.err != nil {
		s += errorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())) + "\n\n"
	}

	if m.saved {
		s += selectedStyle.Render("✓ Comment updated") + "\n\n"
	}

	s += normalStyle.Render("enter to save  •  esc to cancel") + "\n"
	return s
}
