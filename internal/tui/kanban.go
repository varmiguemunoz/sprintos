package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/varmiguemunoz/command_pm_app/internal/app"
	"github.com/varmiguemunoz/command_pm_app/internal/domain"
)

type KanbanModel struct {
	project    domain.Project
	states     []domain.State
	tasks      map[uint][]domain.Task
	colCursor  int
	taskCursor map[uint]int
	loading    bool
	err        error
	stateSvc   *app.StateService
	taskSvc    *app.TaskService
}

type KanbanLoadedMsg struct {
	States []domain.State
	Tasks  []domain.Task
	Err    error
}

func NewKanbanModel(project domain.Project, stateSvc *app.StateService, taskSvc *app.TaskService) KanbanModel {
	return KanbanModel{
		project:    project,
		loading:    true,
		tasks:      make(map[uint][]domain.Task),
		taskCursor: make(map[uint]int),
		stateSvc:   stateSvc,
		taskSvc:    taskSvc,
	}
}

func (m KanbanModel) loadDataCmd() tea.Cmd {
	return func() tea.Msg {
		states, err := m.stateSvc.ListByProject(m.project.ID)
		if err != nil {
			return KanbanLoadedMsg{Err: err}
		}
		tasks, err := m.taskSvc.ListByProject(m.project.ID)
		if err != nil {
			return KanbanLoadedMsg{Err: err}
		}
		return KanbanLoadedMsg{States: states, Tasks: tasks}
	}
}

func (m KanbanModel) Init() tea.Cmd {
	return m.loadDataCmd()
}

func (m KanbanModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case KanbanLoadedMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.loading = false
			return m, nil
		}
		m.states = msg.States
		for _, task := range msg.Tasks {
			m.tasks[task.StateID] = append(m.tasks[task.StateID], task)
		}
		m.loading = false
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			if m.colCursor > 0 {
				m.colCursor--
			}
		case "right", "l":
			if m.colCursor < len(m.states)-1 {
				m.colCursor++
			}
		case "up", "k":
			if len(m.states) > 0 {
				stateID := m.states[m.colCursor].ID
				if m.taskCursor[stateID] > 0 {
					m.taskCursor[stateID]--
				}
			}
		case "down", "j":
			if len(m.states) > 0 {
				stateID := m.states[m.colCursor].ID
				tasks := m.tasks[stateID]
				if m.taskCursor[stateID] < len(tasks)-1 {
					m.taskCursor[stateID]++
				}
			}
		case "esc":
			return m, func() tea.Msg {
				return NavigateMsg{To: screenDashboard}
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m KanbanModel) View() string {
	header := titleStyle.Render(fmt.Sprintf("SprintOS — %s", m.project.Name))

	if m.loading {
		return header + "\n\n" + normalStyle.Render("Loading board...") + "\n"
	}

	if m.err != nil {
		return header + "\n\n" + errorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())) + "\n"
	}

	if len(m.states) == 0 {
		return header + "\n\n" + normalStyle.Render("No states found for this project.") + "\n"
	}

	columns := make([]string, len(m.states))

	for i, state := range m.states {
		isActive := i == m.colCursor
		tasks := m.tasks[state.ID]
		cursor := m.taskCursor[state.ID]

		borderColor := lipgloss.Color("#374151")
		if isActive {
			borderColor = lipgloss.Color("#7C3AED")
		}

		colStyle := lipgloss.NewStyle().
			Width(26).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(0, 1)

		colHeader := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(state.Color)).
			Render(fmt.Sprintf("%s (%d)", state.Name, len(tasks)))

		body := colHeader + "\n\n"

		if len(tasks) == 0 {
			body += normalStyle.Render("  empty") + "\n"
		} else {
			for j, task := range tasks {
				if isActive && j == cursor {
					body += selectedStyle.Render(fmt.Sprintf("> %s", truncate(task.Title, 22))) + "\n"
				} else {
					body += normalStyle.Render(fmt.Sprintf("  %s", truncate(task.Title, 22))) + "\n"
				}
			}
		}

		columns[i] = colStyle.Render(body)
	}

	board := lipgloss.JoinHorizontal(lipgloss.Top, columns...)
	footer := normalStyle.Render("←/→ columns  •  ↑/↓ tasks  •  esc back  •  q quit")

	return header + "\n\n" + board + "\n\n" + footer + "\n"
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
