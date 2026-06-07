package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/domain"
)

type CreateSubtaskModel struct {
	titleInput  textinput.Model
	descInput   textarea.Model
	focused     int
	loading     bool
	err         error
	task        domain.Task
	project     domain.Project
	createdByID uint
	subtaskSvc  *app.SubtaskService
}

type SubtaskCreatedMsg struct {
	Err error
}

func NewCreateSubtaskModel(
	task domain.Task,
	project domain.Project,
	createdByID uint,
	subtaskSvc *app.SubtaskService,
) CreateSubtaskModel {
	ti := textinput.New()
	ti.Placeholder = "Subtask title"
	ti.CharLimit = 150
	ti.Focus()

	ta := textarea.New()
	ta.Placeholder = "Description (optional)"
	ta.CharLimit = 1000
	ta.SetWidth(70)
	ta.SetHeight(5)
	ta.ShowLineNumbers = false

	return CreateSubtaskModel{
		titleInput:  ti,
		descInput:   ta,
		focused:     0,
		task:        task,
		project:     project,
		createdByID: createdByID,
		subtaskSvc:  subtaskSvc,
	}
}

func (m CreateSubtaskModel) submitCmd() tea.Cmd {
	return func() tea.Msg {
		title := m.titleInput.Value()
		description := m.descInput.Value()

		if title == "" {
			return SubtaskCreatedMsg{Err: fmt.Errorf("subtask title is required")}
		}

		_, err := m.subtaskSvc.Create(title, description, m.task.ID, m.createdByID)
		if err != nil {
			return SubtaskCreatedMsg{Err: err}
		}

		return SubtaskCreatedMsg{}
	}
}

func (m CreateSubtaskModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m CreateSubtaskModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(SubtaskCreatedMsg); ok {
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
		case "ctrl+s":
			m.loading = true
			return m, m.submitCmd()
		case "tab":
			cmd := m.focusNext()
			return m, cmd
		case "shift+tab":
			cmd := m.focusPrev()
			return m, cmd
		case "up":
			if m.focused != 1 {
				cmd := m.focusPrev()
				return m, cmd
			}
		case "down":
			if m.focused != 1 {
				cmd := m.focusNext()
				return m, cmd
			}
		case "enter":
			if m.focused == 0 {
				cmd := m.focusNext()
				return m, cmd
			}
		}
	}

	var cmd tea.Cmd
	switch m.focused {
	case 0:
		m.titleInput, cmd = m.titleInput.Update(msg)
	case 1:
		m.descInput, cmd = m.descInput.Update(msg)
	}
	return m, cmd
}

func (m *CreateSubtaskModel) focusNext() tea.Cmd {
	if m.focused == 0 {
		m.titleInput.Blur()
		m.focused = 1
		return m.descInput.Focus()
	}
	m.descInput.Blur()
	m.focused = 0
	m.titleInput.Focus()
	return nil
}

func (m *CreateSubtaskModel) focusPrev() tea.Cmd {
	if m.focused == 1 {
		m.descInput.Blur()
		m.focused = 0
		m.titleInput.Focus()
		return nil
	}
	m.titleInput.Blur()
	m.focused = 1
	return m.descInput.Focus()
}

func (m CreateSubtaskModel) View() string {
	if m.loading {
		return titleStyle.Render("SprintOS — Create Subtask") +
			"\n\n" + normalStyle.Render("Saving...") + "\n"
	}

	s := titleStyle.Render(fmt.Sprintf("SprintOS — New Subtask for: %s", m.task.Title)) + "\n\n"

	if m.focused == 0 {
		s += selectedStyle.Render("Title *") + "\n"
	} else {
		s += dimStyle.Render("Title *") + "\n"
	}
	s += m.titleInput.View() + "\n\n"

	if m.focused == 1 {
		s += selectedStyle.Render("Description") + "\n"
	} else {
		s += dimStyle.Render("Description") + "\n"
	}
	s += m.descInput.View() + "\n\n"

	if m.err != nil {
		s += errorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())) + "\n\n"
	}

	s += dimStyle.Render("tab next  •  shift+tab prev  •  ctrl+s save  •  esc back") + "\n"
	return s
}
