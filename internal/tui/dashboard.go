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
	windowWidth     int
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
		loading:     true,
		orgID:       orgID,
		windowWidth: 90,
		projectSvc:  projectSvc,
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

	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		return m, nil

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
		case "d":
			return m, func() tea.Msg {
				return NavigateMsg{To: screenCEODashboard}
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
			return m, func() tea.Msg {
				return NavigateMsg{To: screenGuide}
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

	w := m.windowWidth
	if w < 60 {
		w = 60
	}
	boxW := w - 4

	count := len(m.projects)
	countLabel := ""
	if count == 1 {
		countLabel = dimStyle.Render("1 project")
	} else if count > 1 {
		countLabel = dimStyle.Render(fmt.Sprintf("%d projects", count))
	}

	header := titleStyle.Render("SprintOS — ") + valueStyle.Render("Projects")
	if countLabel != "" {
		header += "   " + countLabel
	}

	s := header + "\n"

	if m.err != nil {
		s += "\n" + errorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())) + "\n"
		return s
	}

	s += "\n"

	if len(m.projects) == 0 {
		empty := dimStyle.Render("No projects yet — press ") +
			hintKeyStyle.Render("n") +
			dimStyle.Render(" to create your first one")
		s += cardStyle.Width(boxW).Render(empty) + "\n"
	} else {
		for i, project := range m.projects {
			style := cardStyle.Width(boxW)
			prefix := dimStyle.Render(" · ")
			nameStyle := normalStyle
			descStyle := dimStyle

			if i == m.cursor {
				style = activeCardStyle.Width(boxW)
				prefix = selectedStyle.Render(" ▶ ")
				nameStyle = valueStyle
				descStyle = normalStyle
			}

			cardContent := prefix + nameStyle.Render(project.Name)
			if project.Description != nil && *project.Description != "" {
				cardContent += "\n   " + descStyle.Render(truncate(*project.Description, boxW-6))
			}

			s += style.Render(cardContent) + "\n"
		}
	}

	s += "\n"

	if m.deleting && m.selectedProject != nil {
		warn := errorStyle.Render(fmt.Sprintf("⚠  Delete '%s'?", m.selectedProject.Name)) +
			"\n" + dimStyle.Render("   This cannot be undone.")
		s += cardStyle.Width(boxW).Render(warn) + "\n\n"
		s += renderHintBar("y", "confirm", "n", "cancel", "esc", "cancel") + "\n"
	} else {
		s += renderHintBar(
			"↑/↓", "move",
			"enter", "open",
			"n", "new",
			"e", "edit",
			"D", "delete",
			"d", "dashboard",
			"/", "search",
			"s", "settings",
			"?", "guide",
			"q", "quit",
		) + "\n"
	}

	return s
}
