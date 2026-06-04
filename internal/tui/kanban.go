package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/domain"
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
	showHelp     bool
	windowWidth  int
	windowHeight int
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
		project:      project,
		loading:      true,
		windowWidth:  100,
		windowHeight: 30,
		tasks:        make(map[uint][]domain.Task),
		taskCursor:   make(map[uint]int),
		stateSvc:     stateSvc,
		taskSvc:      taskSvc,
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

	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
		return m, nil

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
		case "n", "+":
			if len(m.states) > 0 {
				stateID := m.states[m.colCursor].ID
				project := m.project
				return m, func() tea.Msg {
					return NavigateMsg{To: screenCreateTask, StateID: stateID, Project: project}
				}
			}
		case "?":
			m.showHelp = !m.showHelp
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
		case "v":
			project := m.project
			return m, func() tea.Msg {
				return NavigateMsg{To: screenSprintView, Project: project}
			}
		case "/":
			return m, func() tea.Msg {
				return NavigateMsg{To: screenSearch}
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
	header := titleStyle.Render("SprintOS — ") + valueStyle.Render(m.project.Name)

	if m.loading {
		return header + "\n\n" + normalStyle.Render("Loading board...") + "\n"
	}

	if m.err != nil {
		return header + "\n\n" + errorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())) + "\n"
	}

	if len(m.states) == 0 {
		return header + "\n\n" + cardStyle.Render(
			dimStyle.Render("No columns yet — press ")+hintKeyStyle.Render("b")+" to set up your board",
		) + "\n"
	}

	if m.showHelp {
		return header + "\n\n" + m.renderHelp()
	}

	if m.moving && m.selectedTask != nil {
		return header + "\n\n" + m.renderMoveDialog()
	}

	availW := m.windowWidth * 95 / 100
	if availW < 60 {
		availW = 60
	}
	availH := m.windowHeight * 95 / 100
	if availH < 12 {
		availH = 12
	}

	numCols := len(m.states)
	colContentW := (availW - numCols*5 + 1) / numCols
	if colContentW < 18 {
		colContentW = 18
	}

	// availH - 2 (header+blank) - 2 (blank+footer) - 2 (col top+bottom border) = col content height
	colContentH := availH - 6
	if colContentH < 6 {
		colContentH = 6
	}

	maxColTasks := 1
	for _, state := range m.states {
		if n := len(m.tasks[state.ID]); n > maxColTasks {
			maxColTasks = n
		}
	}

	columns := make([]string, numCols)
	for i, state := range m.states {
		columns[i] = m.renderColumn(state, i, colContentW, colContentH, maxColTasks)
	}

	board := lipgloss.JoinHorizontal(lipgloss.Top, columns...)

	var footer string
	if m.deleting && m.selectedTask != nil {
		warn := errorStyle.Render(fmt.Sprintf("⚠  Delete '%s'?", truncate(m.selectedTask.Title, 40))) +
			"\n" + dimStyle.Render("   This cannot be undone.")
		footer = cardStyle.Render(warn) + "\n" +
			renderHintBar("y", "confirm", "n", "cancel", "esc", "cancel")
	} else {
		footer = renderHintBar(
			"←/→", "cols",
			"↑/↓", "tasks",
			"enter", "view",
			"n", "new",
			"m", "move",
			"d", "delete",
			"v", "sprints",
			"b", "board",
			"?", "help",
			"esc", "back",
		)
	}

	return header + "\n\n" + board + "\n\n" + footer + "\n"
}

func (m KanbanModel) renderColumn(state domain.State, idx int, contentW int, contentH int, maxColTasks int) string {
	tasks := m.tasks[state.ID]
	cursor := m.taskCursor[state.ID]
	isActive := idx == m.colCursor

	borderColor := lipgloss.Color("#374151")
	if isActive {
		borderColor = lipgloss.Color("#7C3AED")
	}

	colStyle := lipgloss.NewStyle().
		Width(contentW).
		Height(contentH).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1)

	stateNameStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(state.Color))

	countStyle := dimStyle
	if isActive {
		countStyle = selectedStyle
	}
	if state.IsDone {
		countStyle = successStyle
	}

	headerLine := stateNameStyle.Render(truncate(state.Name, contentW-6)) +
		"  " + countStyle.Render(fmt.Sprintf("(%d)", len(tasks)))

	barW := contentW
	if barW < 1 {
		barW = 1
	}
	filled := 0
	if maxColTasks > 0 {
		filled = len(tasks) * barW / maxColTasks
	}

	barFillStyle := dimStyle
	if isActive {
		barFillStyle = selectedStyle
	}
	if state.IsDone {
		barFillStyle = successStyle
	}

	progBar := barFillStyle.Render(strings.Repeat("█", filled)) +
		dimStyle.Render(strings.Repeat("░", barW-filled))

	divider := dimStyle.Render(strings.Repeat("─", contentW))

	body := headerLine + "\n" + progBar + "\n" + divider + "\n"

	// 3 rows consumed by header, bar, divider — rest is for tasks
	maxVisible := contentH - 3
	if maxVisible < 1 {
		maxVisible = 1
	}

	if len(tasks) == 0 {
		body += dimStyle.Render("  empty") + "\n"
	} else {
		startIdx := 0
		if cursor >= maxVisible {
			startIdx = cursor - maxVisible + 1
		}
		endIdx := startIdx + maxVisible
		if endIdx > len(tasks) {
			endIdx = len(tasks)
		}

		if startIdx > 0 {
			body += dimStyle.Render(fmt.Sprintf("  ↑ %d above", startIdx)) + "\n"
			if endIdx-startIdx > 1 {
				endIdx--
			}
		}

		for j := startIdx; j < endIdx; j++ {
			body += m.renderTaskRow(tasks[j], j == cursor && isActive, state.IsDone, contentW) + "\n"
		}

		if endIdx < len(tasks) {
			body += dimStyle.Render(fmt.Sprintf("  ↓ %d more", len(tasks)-endIdx)) + "\n"
		}
	}

	return colStyle.Render(body)
}

func (m KanbanModel) renderTaskRow(task domain.Task, isSelected bool, isDone bool, colW int) string {
	titleW := colW - 9
	if titleW < 4 {
		titleW = 4
	}

	pri := priorityIndicator(task.Priority)
	due := dueDateIndicator(task.DueDate)
	av := "  "
	if task.Assignee != nil {
		av = assigneeInitials(task.Assignee.Name)
	}

	title := truncate(task.Title, titleW)

	if isSelected {
		line := fmt.Sprintf("▶ %s%-*s%s %s", pri, titleW, title, due, av)
		return highlightStyle.Render(line)
	}

	if isDone {
		return dimStyle.Render(fmt.Sprintf("  %s%-*s%s %s", pri, titleW, title, due, av))
	}

	return "  " + pri + normalStyle.Render(fmt.Sprintf("%-*s", titleW, title)) + due + " " + dimStyle.Render(av)
}

func (m KanbanModel) renderMoveDialog() string {
	s := sectionHeader(fmt.Sprintf("Move \"%s\" to:", truncate(m.selectedTask.Title, 34))) + "\n\n"
	for i, state := range m.states {
		if i == m.moveCursor {
			s += highlightStyle.Render(fmt.Sprintf("  ▶ %s", state.Name)) + "\n"
		} else {
			s += normalStyle.Render(fmt.Sprintf("    %s", state.Name)) + "\n"
		}
	}
	s += "\n" + renderHintBar("↑/↓", "choose", "enter", "confirm", "esc", "cancel") + "\n"
	return s
}

func (m KanbanModel) renderHelp() string {
	s := sectionHeader("Keyboard shortcuts") + "\n\n"

	s += selectedStyle.Render("  Navigation") + "\n"
	s += labelStyle.Render("  ←/→      ") + normalStyle.Render("Move between columns") + "\n"
	s += labelStyle.Render("  ↑/↓      ") + normalStyle.Render("Move between tasks") + "\n\n"

	s += selectedStyle.Render("  Actions") + "\n"
	s += labelStyle.Render("  enter    ") + normalStyle.Render("View task detail") + "\n"
	s += labelStyle.Render("  n / +    ") + normalStyle.Render("Create new task in current column") + "\n"
	s += labelStyle.Render("  m        ") + normalStyle.Render("Move task to another state") + "\n"
	s += labelStyle.Render("  d        ") + normalStyle.Render("Delete task") + "\n"
	s += labelStyle.Render("  b        ") + normalStyle.Render("Edit board layout") + "\n\n"

	s += selectedStyle.Render("  Other") + "\n"
	s += labelStyle.Render("  v        ") + normalStyle.Render("Sprint view") + "\n"
	s += labelStyle.Render("  /        ") + normalStyle.Render("Search") + "\n"
	s += labelStyle.Render("  ?        ") + normalStyle.Render("Toggle this help") + "\n"
	s += labelStyle.Render("  esc      ") + normalStyle.Render("Back to projects") + "\n"
	s += labelStyle.Render("  q        ") + normalStyle.Render("Quit") + "\n\n"

	s += renderHintBar("?", "close help") + "\n"
	return s
}

func truncate(s string, max int) string {
	if len([]rune(s)) <= max {
		return s
	}
	runes := []rune(s)
	return string(runes[:max-1]) + "…"
}

func assigneeInitials(name string) string {
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "  "
	}
	r0 := []rune(parts[0])
	if len(parts) == 1 {
		if len(r0) >= 2 {
			return strings.ToUpper(string(r0[:2]))
		}
		return strings.ToUpper(string(r0)) + " "
	}
	r1 := []rune(parts[1])
	return strings.ToUpper(string(r0[:1])) + strings.ToUpper(string(r1[:1]))
}

func priorityIndicator(priority string) string {
	switch priority {
	case "high":
		return warningStyle.Render("↑ ")
	case "critical":
		return errorStyle.Render("!!")
	default:
		return "  "
	}
}

func dueDateIndicator(due *time.Time) string {
	if due == nil {
		return "  "
	}
	now := time.Now()
	if due.Before(now) {
		return errorStyle.Render("✗ ")
	}
	if due.Before(now.Add(48 * time.Hour)) {
		return warningStyle.Render("⚠ ")
	}
	return "  "
}
