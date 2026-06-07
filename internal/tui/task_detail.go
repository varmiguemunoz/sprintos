package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gen2brain/beeep"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/domain"
)

type TaskDetailModel struct {
	task            domain.Task
	comments        []domain.Comment
	subtasks        []domain.Subtask
	states          []domain.State
	selectedSubtask int
	activeTimer     *domain.ActiveTimer
	totalMinutes    int
	loading         bool
	err             error
	moving          bool
	moveCursor      int
	project         domain.Project
	windowWidth     int
	taskSvc         *app.TaskService
	commentSvc      *app.CommentService
	subtaskSvc      *app.SubtaskService
	timeSvc         *app.TimeEntryService
	stateSvc        *app.StateService
	currentUserID   uint
}

type TaskDetailLoadedMsg struct {
	Task         domain.Task
	Comments     []domain.Comment
	Subtasks     []domain.Subtask
	States       []domain.State
	ActiveTimer  *domain.ActiveTimer
	TotalMinutes int
	Err          error
}

type SubtaskDeletedMsg struct {
	Err error
}

type TaskTimerToggledMsg struct {
	ActiveTimer *domain.ActiveTimer
	Err         error
}

type TaskDetailTickMsg time.Time

func taskDetailTickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TaskDetailTickMsg(t)
	})
}

func NewTaskDetailModel(
	task domain.Task,
	project domain.Project,
	taskSvc *app.TaskService,
	commentSvc *app.CommentService,
	subtaskSvc *app.SubtaskService,
	timeSvc *app.TimeEntryService,
	stateSvc *app.StateService,
	currentUserID uint,
) TaskDetailModel {
	return TaskDetailModel{
		task:          task,
		project:       project,
		loading:       true,
		windowWidth:   90,
		taskSvc:       taskSvc,
		commentSvc:    commentSvc,
		subtaskSvc:    subtaskSvc,
		timeSvc:       timeSvc,
		stateSvc:      stateSvc,
		currentUserID: currentUserID,
	}
}

func (m TaskDetailModel) loadCmd() tea.Cmd {
	return func() tea.Msg {
		task, err := m.taskSvc.GetByID(m.task.ID)
		if err != nil {
			return TaskDetailLoadedMsg{Err: err}
		}
		comments, err := m.commentSvc.ListByTask(task.ID)
		if err != nil {
			return TaskDetailLoadedMsg{Err: err}
		}
		subtasks, err := m.subtaskSvc.ListByTask(task.ID)
		if err != nil {
			return TaskDetailLoadedMsg{Err: err}
		}
		states, _ := m.stateSvc.ListByProject(task.ProjectID)
		timer, _ := m.timeSvc.GetActiveTimer(m.currentUserID)
		var totalMinutes int
		if len(subtasks) > 0 {
			totalMinutes = m.timeSvc.GetTotalMinutesForTaskWithSubtasks(task.ID)
		} else {
			totalMinutes = m.timeSvc.GetTotalMinutesForTask(task.ID)
		}
		return TaskDetailLoadedMsg{
			Task:         *task,
			Comments:     comments,
			Subtasks:     subtasks,
			States:       states,
			ActiveTimer:  timer,
			TotalMinutes: totalMinutes,
		}
	}
}

func (m TaskDetailModel) deleteSubtaskCmd(id uint) tea.Cmd {
	return func() tea.Msg {
		if err := m.subtaskSvc.Delete(id); err != nil {
			return SubtaskDeletedMsg{Err: err}
		}
		return SubtaskDeletedMsg{}
	}
}

func (m TaskDetailModel) moveTaskCmd(newStateID uint) tea.Cmd {
	return func() tea.Msg {
		_, err := m.taskSvc.MoveState(m.task.ID, newStateID)
		return TaskMovedMsg{Err: err}
	}
}

func (m TaskDetailModel) toggleTimerCmd() tea.Cmd {
	return func() tea.Msg {
		if m.activeTimer != nil && m.activeTimer.TaskID != nil && *m.activeTimer.TaskID == m.task.ID {
			_, err := m.timeSvc.StopTimer(m.currentUserID)
			if err != nil {
				return TaskTimerToggledMsg{Err: err}
			}
			return TaskTimerToggledMsg{ActiveTimer: nil}
		}
		taskID := m.task.ID
		timer, err := m.timeSvc.StartTimer(&taskID, nil, m.currentUserID)
		if err != nil {
			return TaskTimerToggledMsg{Err: err}
		}
		_ = beeep.Notify("SprintOS", "Timer started — "+m.task.Title, "")
		return TaskTimerToggledMsg{ActiveTimer: timer}
	}
}

func (m TaskDetailModel) Init() tea.Cmd {
	return m.loadCmd()
}

func (m TaskDetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(TaskDetailLoadedMsg); ok {
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.task = msg.Task
		m.comments = msg.Comments
		m.subtasks = msg.Subtasks
		m.states = msg.States
		m.activeTimer = msg.ActiveTimer
		m.totalMinutes = msg.TotalMinutes
		if m.selectedSubtask >= len(m.subtasks) {
			m.selectedSubtask = 0
		}
		if m.activeTimer != nil && m.activeTimer.TaskID != nil && *m.activeTimer.TaskID == m.task.ID {
			return m, taskDetailTickCmd()
		}
		return m, nil
	}

	if msg, ok := msg.(TaskMovedMsg); ok {
		m.moving = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.loading = true
		return m, m.loadCmd()
	}

	if msg, ok := msg.(TaskDetailTickMsg); ok {
		_ = msg
		return m, taskDetailTickCmd()
	}

	if msg, ok := msg.(SubtaskDeletedMsg); ok {
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.loading = true
		return m, m.loadCmd()
	}

	if msg, ok := msg.(TaskTimerToggledMsg); ok {
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.activeTimer = msg.ActiveTimer
		if m.activeTimer != nil {
			return m, taskDetailTickCmd()
		}
		m.loading = true
		return m, m.loadCmd()
	}

	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.windowWidth = msg.Width
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
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
				if len(m.states) > 0 {
					newStateID := m.states[m.moveCursor].ID
					return m, m.moveTaskCmd(newStateID)
				}
			case "esc":
				m.moving = false
			}
			return m, nil
		}

		switch msg.String() {
		case "a":
			task := m.task
			project := m.project
			return m, func() tea.Msg {
				return NavigateMsg{To: screenAssignUser, Task: task, Project: project}
			}
		case "c":
			task := m.task
			project := m.project
			return m, func() tea.Msg {
				return NavigateMsg{To: screenCreateComment, Task: task, Project: project}
			}
		case "e":
			task := m.task
			project := m.project
			return m, func() tea.Msg {
				return NavigateMsg{To: screenEditTask, Task: task, Project: project}
			}
		case "s":
			task := m.task
			project := m.project
			return m, func() tea.Msg {
				return NavigateMsg{To: screenCreateSubtask, Task: task, Project: project}
			}
		case "m":
			if len(m.states) > 0 {
				for i, st := range m.states {
					if st.ID == m.task.StateID {
						m.moveCursor = i
						break
					}
				}
				m.moving = true
			}
		case "T":
			if len(m.subtasks) == 0 {
				return m, m.toggleTimerCmd()
			}
		case "l":
			if len(m.subtasks) == 0 {
				task := m.task
				project := m.project
				return m, func() tea.Msg {
					return NavigateMsg{To: screenLogTime, Task: task, Project: project}
				}
			}
		case "up", "k":
			if len(m.subtasks) > 0 && m.selectedSubtask > 0 {
				m.selectedSubtask--
			}
		case "down", "j":
			if len(m.subtasks) > 0 && m.selectedSubtask < len(m.subtasks)-1 {
				m.selectedSubtask++
			}
		case "enter":
			if len(m.subtasks) > 0 {
				subtask := m.subtasks[m.selectedSubtask]
				task := m.task
				project := m.project
				return m, func() tea.Msg {
					return NavigateMsg{To: screenSubtaskDetail, Subtask: subtask, Task: task, Project: project}
				}
			}
		case "d":
			if len(m.subtasks) > 0 {
				id := m.subtasks[m.selectedSubtask].ID
				if m.selectedSubtask > 0 {
					m.selectedSubtask--
				}
				return m, m.deleteSubtaskCmd(id)
			}
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

func (m TaskDetailModel) renderMoveDialog() string {
	s := titleStyle.Render("SprintOS — ") + valueStyle.Render(m.task.Title) + "\n\n"
	s += sectionHeader(fmt.Sprintf("Move \"%s\" to:", truncate(m.task.Title, 34))) + "\n\n"
	for i, st := range m.states {
		if i == m.moveCursor {
			s += highlightStyle.Render(fmt.Sprintf("  ▶ %s", st.Name)) + "\n"
		} else {
			s += normalStyle.Render(fmt.Sprintf("    %s", st.Name)) + "\n"
		}
	}
	s += "\n" + renderHintBar("↑/↓", "choose", "enter", "confirm", "esc", "cancel") + "\n"
	return s
}

func (m TaskDetailModel) View() string {
	if m.loading {
		return titleStyle.Render("SprintOS — Task Detail") +
			"\n\n" + normalStyle.Render("Loading...") + "\n"
	}

	if m.err != nil {
		return titleStyle.Render("SprintOS — Task Detail") +
			"\n\n" + errorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())) + "\n"
	}

	if m.moving {
		return m.renderMoveDialog()
	}

	w := m.windowWidth
	if w < 70 {
		w = 70
	}
	boxW := w - 4

	titleLine := titleStyle.Render("SprintOS — ") + valueStyle.Render(m.task.Title)
	if m.task.State.Name != "" {
		titleLine += "  " + selectedStyle.Render("["+m.task.State.Name+"]")
	}
	if m.task.DueDate != nil && m.task.DueDate.Before(time.Now()) && m.task.CompletedAt == nil {
		titleLine += "  " + errorStyle.Render("[Overdue]")
	}

	s := titleLine + "\n\n"

	s += m.renderDetailsBox(boxW)
	s += "\n"
	s += m.renderTimeBox(boxW)
	s += "\n"
	s += m.renderSubtasksBox(boxW)
	s += "\n"
	s += m.renderCommentsBox(boxW)
	s += "\n"
	s += dimStyle.Render(strings.Repeat("─", w-2)) + "\n"
	s += renderHintBar(
		"a", "assign",
		"c", "comment",
		"e", "edit",
		"m", "move",
		"s", "subtask",
		"T", "timer",
		"l", "log time",
		"esc", "board",
		"q", "quit",
	) + "\n"

	return s
}

func (m TaskDetailModel) renderDetailsBox(boxW int) string {
	var lines []string

	row := func(label, val string) string {
		return labelStyle.Render(fmt.Sprintf("  %-12s", label)) + valueStyle.Render(val)
	}
	colorRow := func(label string, val string, s lipgloss.Style) string {
		return labelStyle.Render(fmt.Sprintf("  %-12s", label)) + s.Render(val)
	}

	lines = append(lines, row("State", m.task.State.Name))

	if m.task.Assignee != nil {
		lines = append(lines, row("Assigned", m.task.Assignee.Name))
	} else {
		lines = append(lines, colorRow("Assigned", "Unassigned", dimStyle))
	}

	if m.task.DueDate != nil {
		dueStr := m.task.DueDate.Format("2006-01-02")
		if m.task.DueDate.Before(time.Now()) && m.task.CompletedAt == nil {
			lines = append(lines, colorRow("Due", dueStr+" · overdue", errorStyle))
		} else {
			lines = append(lines, row("Due", dueStr))
		}
	}

	if m.task.StartDate != nil {
		lines = append(lines, row("Start", m.task.StartDate.Format("2006-01-02")))
	}

	switch m.task.Priority {
	case "high":
		lines = append(lines, colorRow("Priority", "High", warningStyle))
	case "critical":
		lines = append(lines, colorRow("Priority", "Critical", errorStyle))
	}

	if m.task.Description != nil && *m.task.Description != "" {
		lines = append(lines, dimStyle.Render("  "+strings.Repeat("─", boxW-6)))
		lines = append(lines, normalStyle.Render("  "+*m.task.Description))
	}

	content := sectionHeader("Details") + "\n" + strings.Join(lines, "\n")
	return cardStyle.Width(boxW).Render(content)
}

func (m TaskDetailModel) renderTimeBox(boxW int) string {
	timerRunning := m.activeTimer != nil && m.activeTimer.TaskID != nil && *m.activeTimer.TaskID == m.task.ID

	var lines []string
	if len(m.subtasks) > 0 {
		lines = append(lines, dimStyle.Render("  ⊘ Tracked via subtasks"))
		lines = append(lines, labelStyle.Render("  Total  ")+valueStyle.Render(app.FormatMinutes(m.totalMinutes)))
	} else {
		lines = append(lines, labelStyle.Render("  Total  ")+valueStyle.Render(app.FormatMinutes(m.totalMinutes)))
		if timerRunning {
			lines = append(lines, successStyle.Render("  ● ")+selectedStyle.Render(app.FormatElapsed(m.activeTimer.StartedAt)))
		}
		lines = append(lines, dimStyle.Render("  ")+renderHintBar("T", "start/stop timer", "l", "log manually"))
	}

	content := sectionHeader("Time tracked") + "\n" + strings.Join(lines, "\n")
	return cardStyle.Width(boxW).Render(content)
}

func (m TaskDetailModel) renderSubtasksBox(boxW int) string {
	done := 0
	for _, st := range m.subtasks {
		if st.Done {
			done++
		}
	}

	header := "Subtasks"
	if len(m.subtasks) > 0 {
		header = fmt.Sprintf("Subtasks  %s", dimStyle.Render(fmt.Sprintf("%d/%d done", done, len(m.subtasks))))
	}

	var lines []string
	if len(m.subtasks) == 0 {
		lines = append(lines, dimStyle.Render("  No subtasks yet — press s to add one"))
	} else {
		for i, st := range m.subtasks {
			checkmark := dimStyle.Render("[ ]")
			title := normalStyle.Render(st.Title)
			if st.Done {
				checkmark = successStyle.Render("[✓]")
				title = dimStyle.Render(st.Title)
			}
			line := fmt.Sprintf("  %s %s", checkmark, title)
			if i == m.selectedSubtask {
				line = highlightStyle.Render(fmt.Sprintf("  %s ▶ %s", checkmark, st.Title))
			}
			lines = append(lines, line)
		}
		if len(m.subtasks) > 0 {
			lines = append(lines, dimStyle.Render("  ")+renderHintBar("↑/↓", "select", "enter", "open", "d", "delete"))
		}
	}

	content := sectionHeader(header) + "\n" + strings.Join(lines, "\n")
	return cardStyle.Width(boxW).Render(content)
}

func (m TaskDetailModel) renderCommentsBox(boxW int) string {
	var lines []string
	if len(m.comments) == 0 {
		lines = append(lines, dimStyle.Render("  No comments yet"))
	} else {
		for _, c := range m.comments {
			ts := dimStyle.Render("[" + c.CreatedAt.Format("2006-01-02 15:04") + "]")
			author := selectedStyle.Render(c.Author.Name + ":")
			body := normalStyle.Render(c.Content)
			lines = append(lines, fmt.Sprintf("  %s %s %s", ts, author, body))
		}
	}

	content := sectionHeader("Comments") + "\n" + strings.Join(lines, "\n")
	return cardStyle.Width(boxW).Render(content)
}
