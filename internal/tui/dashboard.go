package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/command_pm_app/internal/app"
	"github.com/varmiguemunoz/command_pm_app/internal/domain"
)

type DashboardModel struct {
	projects   []domain.Project
	cursor     int
	loading    bool
	err        error
	orgID      uint
	projectSvc *app.ProjectService
}

type ProjectsLoadedMsg struct {
	Projects []domain.Project
	Err      error
}

func NewDashboardModel(orgID uint, projectSvc *app.ProjectService) DashboardModel {
	return DashboardModel{
		loading:    true,
		orgID:      orgID,
		projectSvc: projectSvc,
	}
}

func (m DashboardModel) loadProjectsCmd() tea.Cmd {
	return func() tea.Msg {
		projects, err := m.projectSvc.ListByOrganization(m.orgID)
		return ProjectsLoadedMsg{Projects: projects, Err: err}
	}
}

func (m DashboardModel) Init() tea.Cmd {
	return m.loadProjectsCmd()
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case ProjectsLoadedMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.loading = false
			return m, nil
		}
		m.projects = msg.Projects
		m.loading = false
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.projects)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.projects) > 0 {
				selected := m.projects[m.cursor]
				return m, func() tea.Msg {
					return NavigateMsg{To: screenKanban, Project: selected}
				}
			}
		case "n":
			return m, func() tea.Msg {
				return NavigateMsg{To: screenCreateProject}
			}
		case "s":
			return m, func() tea.Msg {
				return NavigateMsg{To: screenOrgSettings}
			}
		case "L": 
			return m, func() tea.Msg {
				return NavigateMsg{To: screenLogin}
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m DashboardModel) View() string {
	if m.loading {
		return titleStyle.Render("SprintOS") + "\n\n" +
			normalStyle.Render("Loading projects...") + "\n"
	}

	s := titleStyle.Render("SprintOS — Projects") + "\n\n"

	if m.err != nil {
		return s + errorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())) + "\n"
	}

	if len(m.projects) == 0 {
		s += normalStyle.Render("No projects yet.") + "\n"
	} else {
		for i, project := range m.projects {
			if i == m.cursor {
				s += selectedStyle.Render(fmt.Sprintf("> %s", project.Name)) + "\n"
			} else {
				s += normalStyle.Render(fmt.Sprintf("  %s", project.Name)) + "\n"
			}
		}
	}

	s += "\n" + normalStyle.Render("↑/↓ move  •  enter open  •  n new project  •  s settings  •  L logout  •  q quit") + "\n"
	return s
}
