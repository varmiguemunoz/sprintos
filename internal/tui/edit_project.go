package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/command_pm_app/internal/app"
	"github.com/varmiguemunoz/command_pm_app/internal/domain"
)

type EditProjectModel struct {
	inputs     []textinput.Model
	focused    int
	loading    bool
	saved      bool
	err        error
	project    domain.Project
	projectSvc *app.ProjectService
}

type ProjectUpdatedMsg struct {
	Err error
}

func NewEditProjectModel(project domain.Project, projectSvc *app.ProjectService) EditProjectModel {
	inputs := make([]textinput.Model, 2)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Project name"
	inputs[0].CharLimit = 100
	inputs[0].SetValue(project.Name)
	inputs[0].Focus()

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Description (optional)"
	inputs[1].CharLimit = 200
	if project.Description != nil {
		inputs[1].SetValue(*project.Description)
	}

	return EditProjectModel{
		inputs:     inputs,
		project:    project,
		projectSvc: projectSvc,
	}
}

func (m EditProjectModel) saveCmd() tea.Cmd {
	return func() tea.Msg {
		name := m.inputs[0].Value()
		description := m.inputs[1].Value()

		if name == "" {
			return ProjectUpdatedMsg{Err: fmt.Errorf("project name is required")}
		}

		_, err := m.projectSvc.Update(m.project.ID, name, description)
		if err != nil {
			return ProjectUpdatedMsg{Err: err}
		}

		return ProjectUpdatedMsg{}
	}
}

func (m EditProjectModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m EditProjectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(ProjectUpdatedMsg); ok {
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.saved = true
		m.err = nil
		return m, nil
	}

	if m.loading {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg {
				return NavigateMsg{To: screenDashboard}
			}
		case "ctrl+c":
			return m, tea.Quit
		case "tab", "down":
			m.inputs[m.focused].Blur()
			m.focused = (m.focused + 1) % len(m.inputs)
			m.inputs[m.focused].Focus()
			m.saved = false
		case "shift+tab", "up":
			m.inputs[m.focused].Blur()
			m.focused--
			if m.focused < 0 {
				m.focused = len(m.inputs) - 1
			}
			m.inputs[m.focused].Focus()
			m.saved = false
		case "enter":
			if m.focused == len(m.inputs)-1 {
				m.loading = true
				return m, m.saveCmd()
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

func (m EditProjectModel) View() string {
	if m.loading {
		return titleStyle.Render("SprintOS — Edit Project") +
			"\n\n" + normalStyle.Render("Saving...") + "\n"
	}

	labels := []string{"Project name *", "Description"}

	s := titleStyle.Render(fmt.Sprintf("SprintOS — Edit: %s", m.project.Name)) + "\n\n"

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

	if m.saved {
		s += selectedStyle.Render("✓ Project updated successfully") + "\n\n"
	}

	s += normalStyle.Render("tab/↓ next field  •  shift+tab/↑ previous  •  enter to save  •  esc back") + "\n"
	return s
}
