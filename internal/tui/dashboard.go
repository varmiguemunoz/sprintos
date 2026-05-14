package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/domain"
)

type DashboardModel struct {
	projects        []domain.Project
	cursor          int
	loading         bool
	err             error
	orgID           uint
	deleting        bool
	selectedProject *domain.Project
	showHelp        bool
	projectSvc      *app.ProjectService
}

type ProjectsLoadedMsg struct {
	Projects []domain.Project
	Err      error
}

type ProjectDeletedMsg struct {
	Err error
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

func (m DashboardModel) deleteProjectCmd(id uint) tea.Cmd {
	return func() tea.Msg {
		err := m.projectSvc.Delete(id)
		return ProjectDeletedMsg{Err: err}
	}
}

func (m DashboardModel) Init() tea.Cmd {
	return m.loadProjectsCmd()
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(ProjectDeletedMsg); ok {
		m.deleting = false
		m.selectedProject = nil
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.loading = true
		m.projects = nil
		return m, m.loadProjectsCmd()
	}

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
		if m.deleting {
			switch msg.String() {
			case "y", "Y":
				if m.selectedProject != nil {
					id := m.selectedProject.ID
					return m, m.deleteProjectCmd(id)
				}
			case "n", "N", "esc":
				m.deleting = false
				m.selectedProject = nil
			}
			return m, nil
		}

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
		case "e":
			if len(m.projects) > 0 {
				selected := m.projects[m.cursor]
				return m, func() tea.Msg {
					return NavigateMsg{To: screenEditProject, Project: selected}
				}
			}
		case "D":
			if len(m.projects) > 0 {
				p := m.projects[m.cursor]
				m.selectedProject = &p
				m.deleting = true
			}
		case "n":
			return m, func() tea.Msg {
				return NavigateMsg{To: screenCreateProject}
			}
		case "/":
			return m, func() tea.Msg {
				return NavigateMsg{To: screenSearch}
			}
		case "?": 
			m.showHelp = !m.showHelp
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
			//	if project.Description != nil && *project.Description != "" {
			//		s += normalStyle.Render(fmt.Sprintf("  %s", *project.Description)) + "\n"
			// }
			} else {
				s += normalStyle.Render(fmt.Sprintf("  %s", project.Name)) + "\n"
			//	if project.Description != nil && *project.Description != "" {
			//		s += normalStyle.Render(fmt.Sprintf("  %s", *project.Description)) + "\n"
			//	}
			}
		}
	}

	if m.deleting && m.selectedProject != nil {
		s += "\n" + errorStyle.Render(fmt.Sprintf("Delete '%s'? This cannot be undone.", m.selectedProject.Name)) + "\n"
		s += normalStyle.Render("y to confirm  •  n / esc to cancel") + "\n"
	} else {
		s += "\n" + normalStyle.Render("↑/↓ move  •  enter open  •  n new  •  e edit  •  D delete  •  / search  •  s settings  •  ? help  •  q quit") + "\n"
	}

	return s
}
