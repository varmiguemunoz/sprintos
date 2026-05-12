package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/command_pm_app/internal/app"
	"github.com/varmiguemunoz/command_pm_app/internal/domain"
)

type CreateTaskModel struct {
	inputs      []textinput.Model
	focused     int
	loading     bool
	err         error
	stateID     uint
	projectID   uint
	createdByID uint
	project     domain.Project
	taskSvc     *app.TaskService
}

type TaskCreatedMsg struct {
	Err error
}

func NewCreateTaskModel(
	stateID uint,
	projectID uint,
	createdByID uint,
	project domain.Project,
	taskSvc *app.TaskService,
) CreateTaskModel {
	inputs := make([]textinput.Model, 2)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Task title"
	inputs[0].CharLimit = 150
	inputs[0].Focus()

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Description (optional)"
	inputs[1].CharLimit = 500

	return CreateTaskModel{
		inputs:      inputs,
		focused:     0,
		stateID:     stateID,
		projectID:   projectID,
		createdByID: createdByID,
		project:     project,
		taskSvc:     taskSvc,
	}
}

func (m CreateTaskModel) submitCmd() tea.Cmd {
	return func() tea.Msg {
		title := m.inputs[0].Value()
		description := m.inputs[1].Value()

		if title == "" {
			return TaskCreatedMsg{Err: fmt.Errorf("title is required")}
		}

		_, err := m.taskSvc.Create(
			title,
			description,
			m.stateID,
			m.projectID,
			m.createdByID,
			nil,
			nil,
			nil,
		)
		if err != nil {
			return TaskCreatedMsg{Err: err}
		}

		return TaskCreatedMsg{}
	}
}

func (m CreateTaskModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m CreateTaskModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(TaskCreatedMsg); ok {
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		project := m.project
		return m, func() tea.Msg {
			return NavigateMsg{To: screenKanban, Project: project}
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
			project := m.project
			return m, func() tea.Msg {
				return NavigateMsg{To: screenKanban, Project: project}
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

func (m CreateTaskModel) View() string {
	if m.loading {
		return titleStyle.Render("SprintOS — Create Task") +
			"\n\n" + normalStyle.Render("Creating task...") + "\n"
	}

	labels := []string{"Title *", "Description"}

	s := titleStyle.Render("SprintOS — Create Task") + "\n\n"

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

	s += normalStyle.Render("tab/↓ next field  •  shift+tab/↑ previous  •  enter to confirm  •  esc to cancel") + "\n"
	return s
}
