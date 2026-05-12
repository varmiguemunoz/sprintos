package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/command_pm_app/internal/app"
	"github.com/varmiguemunoz/command_pm_app/internal/domain"
)

type CreateProjectModel struct {
	inputs      []textinput.Model
	focused     int
	loading     bool
	err         error
	orgID       uint
	createdByID uint
	projectSvc  *app.ProjectService
	stateSvc    *app.StateService
}

type ProjectCreatedMsg struct {
	Project *domain.Project
	Err     error
}

func NewCreateProjectModel(
	orgID uint,
	createdByID uint,
	projectSvc *app.ProjectService,
	stateSvc *app.StateService,
) CreateProjectModel {
	inputs := make([]textinput.Model, 2)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "My first project"
	inputs[0].CharLimit = 100
	inputs[0].Focus()

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "What is this project about? (optional)"
	inputs[1].CharLimit = 200

	return CreateProjectModel{
		inputs:      inputs,
		focused:     0,
		orgID:       orgID,
		createdByID: createdByID,
		projectSvc:  projectSvc,
		stateSvc:    stateSvc,
	}
}

func (m CreateProjectModel) submitCmd() tea.Cmd {
	return func() tea.Msg {
		name := m.inputs[0].Value()
		description := m.inputs[1].Value()

		if name == "" {
			return ProjectCreatedMsg{Err: fmt.Errorf("project name is required")}
		}

		project, err := m.projectSvc.Create(name, description, m.orgID, m.createdByID)
		if err != nil {
			return ProjectCreatedMsg{Err: err}
		}

		if err := m.stateSvc.ApplyTemplate(project.ID, "standard"); err != nil {
			return ProjectCreatedMsg{Err: err}
		}

		return ProjectCreatedMsg{Project: project}
	}
}

func (m CreateProjectModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m CreateProjectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	if msg,ok := msg.(ProjectCreatedMsg); ok { 
		if msg.Err != nil { 
			m.err = msg.Err
			m.loading = false
			return m, nil
		}
		return m, func() tea.Msg { 
			return NavigateMsg{To: screenDashboard}
		}
	}

	if m.loading {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
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

	case ProjectCreatedMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.loading = false
			return m, nil
		}
		return m, func() tea.Msg {
			return NavigateMsg{To: screenDashboard}
		}
	}

	var cmd tea.Cmd
	m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
	return m, cmd
}

func (m CreateProjectModel) View() string {
	if m.loading {
		return titleStyle.Render("SprintOS — Create Project") +
			"\n\n" + normalStyle.Render("Creating your project...") + "\n"
	}

	labels := []string{"Project name *", "Description"}

	s := titleStyle.Render("SprintOS — Create your first project") + "\n\n"
	s += normalStyle.Render("The standard workflow will be applied automatically:") + "\n"
	s += normalStyle.Render("Backlog → In Progress → In Review → Done") + "\n\n"

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

	s += normalStyle.Render("tab/↓ next field  •  shift+tab/↑ previous  •  enter to confirm  •  esc to quit") + "\n"
	return s
}
