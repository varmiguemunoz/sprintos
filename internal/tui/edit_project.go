package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/domain"
)

type EditProjectModel struct {
	nameInput  textinput.Model
	descInput  textarea.Model
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
	ni := textinput.New()
	ni.Placeholder = "Project name"
	ni.CharLimit = 100
	ni.SetValue(project.Name)
	ni.Focus()

	ta := textarea.New()
	ta.Placeholder = "Description (optional)"
	ta.CharLimit = 500
	ta.SetWidth(70)
	ta.SetHeight(4)
	ta.ShowLineNumbers = false
	if project.Description != nil {
		ta.SetValue(*project.Description)
	}

	return EditProjectModel{
		nameInput:  ni,
		descInput:  ta,
		focused:    0,
		project:    project,
		projectSvc: projectSvc,
	}
}

func (m EditProjectModel) saveCmd() tea.Cmd {
	return func() tea.Msg {
		name := m.nameInput.Value()
		description := m.descInput.Value()

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
		case "ctrl+s":
			m.loading = true
			m.saved = false
			return m, m.saveCmd()
		case "tab":
			m.saved = false
			cmd := m.focusNext()
			return m, cmd
		case "shift+tab":
			m.saved = false
			cmd := m.focusPrev()
			return m, cmd
		case "up":
			if m.focused != 1 {
				m.saved = false
				cmd := m.focusPrev()
				return m, cmd
			}
		case "down":
			if m.focused != 1 {
				m.saved = false
				cmd := m.focusNext()
				return m, cmd
			}
		case "enter":
			if m.focused == 0 {
				m.saved = false
				cmd := m.focusNext()
				return m, cmd
			}
		}
	}

	var cmd tea.Cmd
	switch m.focused {
	case 0:
		m.nameInput, cmd = m.nameInput.Update(msg)
	case 1:
		m.descInput, cmd = m.descInput.Update(msg)
	}
	return m, cmd
}

func (m *EditProjectModel) focusNext() tea.Cmd {
	if m.focused == 0 {
		m.nameInput.Blur()
		m.focused = 1
		return m.descInput.Focus()
	}
	m.descInput.Blur()
	m.focused = 0
	m.nameInput.Focus()
	return nil
}

func (m *EditProjectModel) focusPrev() tea.Cmd {
	if m.focused == 1 {
		m.descInput.Blur()
		m.focused = 0
		m.nameInput.Focus()
		return nil
	}
	m.nameInput.Blur()
	m.focused = 1
	return m.descInput.Focus()
}

func (m EditProjectModel) View() string {
	if m.loading {
		return titleStyle.Render("SprintOS — Edit Project") +
			"\n\n" + normalStyle.Render("Saving...") + "\n"
	}

	s := titleStyle.Render(fmt.Sprintf("SprintOS — Edit: %s", m.project.Name)) + "\n\n"

	if m.focused == 0 {
		s += selectedStyle.Render("Project name *") + "\n"
	} else {
		s += dimStyle.Render("Project name *") + "\n"
	}
	s += m.nameInput.View() + "\n\n"

	if m.focused == 1 {
		s += selectedStyle.Render("Description") + "\n"
	} else {
		s += dimStyle.Render("Description") + "\n"
	}
	s += m.descInput.View() + "\n\n"

	if m.err != nil {
		s += errorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())) + "\n\n"
	}

	if m.saved {
		s += successStyle.Render("✓ Project updated successfully") + "\n\n"
	}

	s += dimStyle.Render("tab next  •  shift+tab prev  •  ctrl+s save  •  esc back") + "\n"
	return s
}
