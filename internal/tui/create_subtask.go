package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/domain"
)

type CreateSubtaskModel struct {
	inputs      []textinput.Model
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
	inputs := make([]textinput.Model, 2)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Subtask title"
	inputs[0].CharLimit = 150
	inputs[0].Focus()

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Description (optional)"
	inputs[1].CharLimit = 500

	return CreateSubtaskModel{
		inputs:      inputs,
		task:        task,
		project:     project,
		createdByID: createdByID,
		subtaskSvc:  subtaskSvc,
	}
}

func (m CreateSubtaskModel) submitCmd() tea.Cmd {
	return func() tea.Msg {
		title := m.inputs[0].Value()
		description := m.inputs[1].Value()

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
		case "tab", "down":
			m.inputs[m.focused].Blur()
			m.focused = (m.focused + 1) % len(m.inputs)
			m.inputs[m.focused].Focus()
		case "shift+tab", "up":
			m.inputs[m.focused].Blur()
			m.focused--
			if m.focused < 0 {
				m.focused = len(m.inputs) - 1
			}
			m.inputs[m.focused].Focus()
		case "enter":
			if m.focused == len(m.inputs)-1 {
				m.loading = true
				return m, m.submitCmd()
			}
			m.inputs[m.focused].Blur()
			m.focused++
			m.inputs[m.focused].Focus()
		}
	}

	var cmd tea.Cmd
	m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
	return m, cmd
}

func (m CreateSubtaskModel) View() string {
	if m.loading {
		return titleStyle.Render("SprintOS — Create Subtask") +
			"\n\n" + normalStyle.Render("Saving...") + "\n"
	}

	labels := []string{"Title *", "Description"}

	s := titleStyle.Render(fmt.Sprintf("SprintOS — New Subtask for: %s", m.task.Title)) + "\n\n"

	for i, label := range labels {
		if i == m.focused {
			s += selectedStyle.Render(label) + "\n"
		} else {
			s += normalStyle.Render(label) + "\n"
		}
		s += m.inputs[i].View() + "\n\n"
	}

	if m.err != nil {
		s += errorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())) + "\n\n"
	}

	s += normalStyle.Render("tab/↓ next  •  shift+tab/↑ previous  •  enter save  •  esc back") + "\n"
	return s
}
