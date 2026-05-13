package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/varmiguemunoz/command_pm_app/internal/app"
	"github.com/varmiguemunoz/command_pm_app/internal/domain"
)

type KanbanModel struct {
	project      domain.Project
	states       []domain.State
	tasks        map[uint][]domain.Task
	colCursor    int
	taskCursor   map[uint]int
	loading      bool
	err          error
	moving       bool
	moveCursor   int
	selectedTask *domain.Task
	deleting     bool
	stateSvc     *app.StateService
	taskSvc      *app.TaskService
}

type KanbanLoadedMsg struct {
	States []domain.State
	Tasks  []domain.Task
	Err    error
}

type TaskMovedMsg struct {
	Err error
}

type TaskDeletedMsg struct {
	Err error
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

func (m KanbanModel) moveTaskCmd(taskID uint, newStateID uint) tea.Cmd {
	return func() tea.Msg {
		_, err := m.taskSvc.MoveState(taskID, newStateID)
		if err != nil {
			return TaskMovedMsg{Err: err}
		}
		return TaskMovedMsg{}
	}
}

func (m KanbanModel) deleteTaskCmd(taskID uint) tea.Cmd {
	return func() tea.Msg {
		err := m.taskSvc.Delete(taskID)
		if err != nil {
			return TaskDeletedMsg{Err: err}
		}
		return TaskDeletedMsg{}
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
		m.tasks = make(map[uint][]domain.Task)
		for _, task := range msg.Tasks {
			m.tasks[task.StateID] = append(m.tasks[task.StateID], task)
		}
		m.loading = false
		return m, nil

	case TaskMovedMsg:
		m.moving = false
		m.selectedTask = nil
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.loading = true
		m.tasks = make(map[uint][]domain.Task)
		return m, m.loadDataCmd()

	case TaskDeletedMsg:
		m.deleting = false
		m.selectedTask = nil
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.loading = true
		m.tasks = make(map[uint][]domain.Task)
		return m, m.loadDataCmd()

	case tea.KeyMsg:
		if m.deleting {
			switch msg.String() {
			case "y", "Y":
				if m.selectedTask != nil {
					taskID := m.selectedTask.ID
					return m, m.deleteTaskCmd(taskID)
				}
			case "n", "N", "esc":
				m.deleting = false
				m.selectedTask = nil
			}
			return m, nil
		}

		if m.moving {
			switch msg.String() {
			case "up", "k":
				if m.moveCursor > 0 {
					m.moveCursor--
				}
			case "down", "j":
				if m.moveCursor < len(m.states)-1 {
					m.moveCursor++
				}
			case "enter":
				if m.selectedTask != nil {
					taskID := m.selectedTask.ID
					newStateID := m.states[m.moveCursor].ID
					return m, m.moveTaskCmd(taskID, newStateID)
				}
			case "esc":
				m.moving = false
				m.selectedTask = nil
			}
			return m, nil
		}

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
		case "m":
			if len(m.states) > 0 {
				stateID := m.states[m.colCursor].ID
				tasks := m.tasks[stateID]
				if len(tasks) > 0 {
					task := tasks[m.taskCursor[stateID]]
					m.selectedTask = &task
					m.moveCursor = m.colCursor
					m.moving = true
				}
			}
		case "d":
			if len(m.states) > 0 {
				stateID := m.states[m.colCursor].ID
				tasks := m.tasks[stateID]
				if len(tasks) > 0 {
					task := tasks[m.taskCursor[stateID]]
					m.selectedTask = &task
					m.deleting = true
				}
			}
		case "n":
			if len(m.states) > 0 {
				stateID := m.states[m.colCursor].ID
				project := m.project
				return m, func() tea.Msg {
					return NavigateMsg{To: screenCreateTask, StateID: stateID, Project: project}
				}
			}
		case "enter":
			if len(m.states) > 0 {
				stateID := m.states[m.colCursor].ID
				tasks := m.tasks[stateID]
				if len(tasks) > 0 {
					selected := tasks[m.taskCursor[stateID]]
					project := m.project
					return m, func() tea.Msg {
						return NavigateMsg{To: screenTaskDetail, Task: selected, Project: project}
					}
				}
			}
		case "b":
			project := m.project
			return m, func() tea.Msg {
				return NavigateMsg{To: screenBoardSetup, Project: project, Editing: true}
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

	var footer string

	if m.deleting && m.selectedTask != nil {
		footer = errorStyle.Render(fmt.Sprintf("Delete '%s'? This cannot be undone.", truncate(m.selectedTask.Title, 30))) + "\n"

		footer += normalStyle.Render("y to confirm  •  n / esc to cancel")
	} else if m.moving && m.selectedTask != nil {
		footer = selectedStyle.Render(fmt.Sprintf("Move \"%s\" to:", truncate(m.selectedTask.Title, 30))) + "\n"
		for i, state := range m.states {
			if i == m.moveCursor {
				footer += selectedStyle.Render(fmt.Sprintf("  > %s", state.Name)) + "\n"
			} else {
				footer += normalStyle.Render(fmt.Sprintf("    %s", state.Name)) + "\n"
			}
		}
		footer += "\n" + normalStyle.Render("↑/↓ choose state  •  enter to confirm  •  esc to cancel")
	} else {
		footer = normalStyle.Render("←/→ columns  •  ↑/↓ tasks  •  enter view  •  m move  •  d delete  •  n new task  •  b edit board  •  esc back  •  q quit")
	}

	return header + "\n\n" + board + "\n\n" + footer + "\n"
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
