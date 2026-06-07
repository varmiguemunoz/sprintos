package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/domain"
)

type CreateProjectModel struct {
	nameInput   textinput.Model
	descInput   textarea.Model
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
	ni := textinput.New()
	ni.Placeholder = "My first project"
	ni.CharLimit = 100
	ni.Focus()

	ta := textarea.New()
	ta.Placeholder = "What is this project about? (optional)"
	ta.CharLimit = 500
	ta.SetWidth(70)
	ta.SetHeight(4)
	ta.ShowLineNumbers = false

	return CreateProjectModel{
		nameInput:   ni,
		descInput:   ta,
		focused:     0,
		orgID:       orgID,
		createdByID: createdByID,
		projectSvc:  projectSvc,
		stateSvc:    stateSvc,
	}
}

func (m CreateProjectModel) submitCmd() tea.Cmd {
	return func() tea.Msg {
		name := m.nameInput.Value()
		description := m.descInput.Value()

		if name == "" {
			return ProjectCreatedMsg{Err: fmt.Errorf("project name is required")}
		}

		project, err := m.projectSvc.Create(name, description, m.orgID, m.createdByID)
		if err != nil {
			return ProjectCreatedMsg{Err: err}
		}

		return ProjectCreatedMsg{Project: project}
	}
}

func (m CreateProjectModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m CreateProjectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(ProjectCreatedMsg); ok {
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		project := *msg.Project
		return m, func() tea.Msg {
			return NavigateMsg{To: screenBoardSetup, Project: project, Editing: false}
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
		m.nameInput, cmd = m.nameInput.Update(msg)
	case 1:
		m.descInput, cmd = m.descInput.Update(msg)
	}
	return m, cmd
}

func (m *CreateProjectModel) focusNext() tea.Cmd {
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

func (m *CreateProjectModel) focusPrev() tea.Cmd {
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

func (m CreateProjectModel) View() string {
	if m.loading {
		return titleStyle.Render("SprintOS — Create Project") +
			"\n\n" + normalStyle.Render("Creating your project...") + "\n"
	}

	s := titleStyle.Render("SprintOS — Create your first project") + "\n\n"
	s += normalStyle.Render("The standard workflow will be applied automatically:") + "\n"
	s += normalStyle.Render("Backlog → In Progress → In Review → Done") + "\n\n"

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

	s += dimStyle.Render("tab next  •  shift+tab prev  •  ctrl+s save  •  esc quit") + "\n"
	return s
}
