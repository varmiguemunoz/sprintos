package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/domain"
)

type SprintViewModel struct {
	project      domain.Project
	sprints      []domain.Sprint
	sprintTasks  map[uint][]domain.Task
	cursor       int
	loading      bool
	planningMode bool
	backlog      []domain.Task
	backlogCursor int
	err          error
	sprintSvc    *app.SprintService
	taskSvc      *app.TaskService
	stateSvc     *app.StateService
}

type SprintDataLoadedMsg struct {
	Sprints     []domain.Sprint
	SprintTasks map[uint][]domain.Task
	Backlog     []domain.Task
	Err         error
}

func NewSprintViewModel(project domain.Project, sprintSvc *app.SprintService, taskSvc *app.TaskService, stateSvc *app.StateService) SprintViewModel {
	return SprintViewModel{
		project:     project,
		loading:     true,
		sprintTasks: make(map[uint][]domain.Task),
		sprintSvc:   sprintSvc,
		taskSvc:     taskSvc,
		stateSvc:    stateSvc,
	}
}

func (m SprintViewModel) loadCmd() tea.Cmd {
	return func() tea.Msg {
		sprints, err := m.sprintSvc.ListByProject(m.project.ID)
		if err != nil {
			return SprintDataLoadedMsg{Err: err}
		}
		sprintTasks := make(map[uint][]domain.Task)
		for _, sp := range sprints {
			tasks, _ := m.sprintSvc.ListTasks(sp.ID)
			sprintTasks[sp.ID] = tasks
		}
		allTasks, _ := m.taskSvc.ListByProject(m.project.ID)
		var backlog []domain.Task
		for _, t := range allTasks {
			if t.SprintID == nil {
				backlog = append(backlog, t)
			}
		}
		return SprintDataLoadedMsg{Sprints: sprints, SprintTasks: sprintTasks, Backlog: backlog}
	}
}

func (m SprintViewModel) Init() tea.Cmd {
	return m.loadCmd()
}

func (m SprintViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(SprintDataLoadedMsg); ok {
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.sprints = msg.Sprints
		m.sprintTasks = msg.SprintTasks
		m.backlog = msg.Backlog
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.planningMode {
			switch msg.String() {
			case "esc":
				m.planningMode = false
			case "up", "k":
				if m.backlogCursor > 0 {
					m.backlogCursor--
				}
			case "down", "j":
				if m.backlogCursor < len(m.backlog)-1 {
					m.backlogCursor++
				}
			case "enter":
				if len(m.sprints) > 0 && m.cursor < len(m.sprints) && len(m.backlog) > 0 {
					sprintID := m.sprints[m.cursor].ID
					taskID := m.backlog[m.backlogCursor].ID
					_ = m.sprintSvc.AddTask(sprintID, taskID)
					m.loading = true
					return m, m.loadCmd()
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.sprints)-1 {
				m.cursor++
			}
		case "c":
			project := m.project
			return m, func() tea.Msg {
				return NavigateMsg{To: screenCreateSprintTUI, Project: project}
			}
		case "p":
			m.planningMode = true
			m.backlogCursor = 0
		case "esc":
			project := m.project
			return m, func() tea.Msg {
				return NavigateMsg{To: screenKanban, Project: project}
			}
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m SprintViewModel) View() string {
	header := titleStyle.Render(fmt.Sprintf("SprintOS — %s — Sprints", m.project.Name))

	if m.loading {
		return header + "\n\n" + normalStyle.Render("Loading sprints...") + "\n"
	}

	if m.err != nil {
		return header + "\n\n" + errorStyle.Render(m.err.Error()) + "\n"
	}

	if len(m.sprints) == 0 {
		s := header + "\n\n"
		s += normalStyle.Render("No sprints yet. Press c to create a sprint") + "\n"
		s += normalStyle.Render("esc back • c create • p plan mode • q quit") + "\n"
		return s
	}

	if m.planningMode {
		return m.viewPlanning()
	}

	s := header + "\n\n"
	now := time.Now()

	for i, sprint := range m.sprints {
		tasks := m.sprintTasks[sprint.ID]
		completed := 0
		for _, t := range tasks {
			if t.CompletedAt != nil {
				completed++
			}
		}

		status := "upcoming"
		statusStyle := normalStyle
		if sprint.CompletedAt != nil {
			status = "completed"
			statusStyle = normalStyle
		} else if sprint.StartDate.Before(now) && sprint.EndDate.After(now) {
			status = "● active"
			statusStyle = selectedStyle
		}

		daysLeft := ""
		if sprint.CompletedAt == nil && sprint.EndDate.After(now) {
			d := int(sprint.EndDate.Sub(now).Hours() / 24)
			daysLeft = fmt.Sprintf(" (%d days left)", d)
		}

		header := fmt.Sprintf("%s  %s → %s  %d/%d tasks%s",
			sprint.Name,
			sprint.StartDate.Format("Jan 2"),
			sprint.EndDate.Format("Jan 2"),
			completed, len(tasks),
			daysLeft,
		)

		if i == m.cursor {
			s += selectedStyle.Render("> "+header) + "  " + statusStyle.Render(status) + "\n"
			for _, t := range tasks {
				done := " "
				if t.CompletedAt != nil {
					done = "✓"
				}
				s += normalStyle.Render(fmt.Sprintf("    [%s] #%d %s", done, t.TaskNumber, truncate(t.Title, 40))) + "\n"
			}
			if len(tasks) == 0 {
				s += normalStyle.Render("    No tasks in this sprint. Press p to plan.") + "\n"
			}
		} else {
			s += normalStyle.Render("  "+header) + "  " + statusStyle.Render(status) + "\n"
		}
	}

	s += "\n" + normalStyle.Render("↑/↓ sprint • c create sprint  • esc back  •  q quit") + "\n"
	return s
}

func (m SprintViewModel) viewPlanning() string {
	if len(m.sprints) == 0 || m.cursor >= len(m.sprints) {
		return titleStyle.Render("No sprints available") + "\n"
	}
	sprint := m.sprints[m.cursor]
	s := titleStyle.Render(fmt.Sprintf("Planning: %s", sprint.Name)) + "\n\n"

	left := selectedStyle.Render("Backlog") + "\n"
	if len(m.backlog) == 0 {
		left += normalStyle.Render("  (empty)") + "\n"
	}
	for i, t := range m.backlog {
		line := fmt.Sprintf("  #%d %s", t.TaskNumber, truncate(t.Title, 22))
		if i == m.backlogCursor {
			left += selectedStyle.Render("> "+strings.TrimPrefix(line, "  ")) + "\n"
		} else {
			left += normalStyle.Render(line) + "\n"
		}
	}

	right := selectedStyle.Render("Sprint tasks") + "\n"
	tasks := m.sprintTasks[sprint.ID]
	if len(tasks) == 0 {
		right += normalStyle.Render("  (empty)") + "\n"
	}
	for _, t := range tasks {
		right += normalStyle.Render(fmt.Sprintf("  #%d %s", t.TaskNumber, truncate(t.Title, 22))) + "\n"
	}

	leftBox := lipgloss.NewStyle().Width(32).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#6B7280")).Padding(0, 1).Render(left)
	rightBox := lipgloss.NewStyle().Width(32).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#7C3AED")).Padding(0, 1).Render(right)

	s += lipgloss.JoinHorizontal(lipgloss.Top, leftBox, "  ", rightBox)
	s += "\n\n" + normalStyle.Render("↑/↓ move  •  enter add to sprint  •  esc exit planning") + "\n"
	return s
}
