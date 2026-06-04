package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/domain"
)

type SubtaskDetailModel struct {
	subtask           domain.Subtask
	comments          []domain.SubtaskComment
	selectedComment   int
	activeTimer       *domain.ActiveTimer
	totalMinutes      int
	loading           bool
	err               error
	task              domain.Task
	project           domain.Project
	subtaskSvc        *app.SubtaskService
	subtaskCommentSvc *app.SubtaskCommentService
	timeSvc           *app.TimeEntryService
	currentUserID     uint
}

type SubtaskDetailLoadedMsg struct {
	Subtask      domain.Subtask
	Comments     []domain.SubtaskComment
	ActiveTimer  *domain.ActiveTimer
	TotalMinutes int
	Err          error
}

type SubtaskTimerToggledMsg struct {
	ActiveTimer *domain.ActiveTimer
	Err         error
}

type SubtaskDetailTickMsg time.Time

func subtaskDetailTickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return SubtaskDetailTickMsg(t)
	})
}

func NewSubtaskDetailModel(
	subtask domain.Subtask,
	task domain.Task,
	project domain.Project,
	subtaskSvc *app.SubtaskService,
	subtaskCommentSvc *app.SubtaskCommentService,
	timeSvc *app.TimeEntryService,
	currentUserID uint,
) SubtaskDetailModel {
	return SubtaskDetailModel{
		subtask:           subtask,
		task:              task,
		project:           project,
		loading:           true,
		subtaskSvc:        subtaskSvc,
		subtaskCommentSvc: subtaskCommentSvc,
		timeSvc:           timeSvc,
		currentUserID:     currentUserID,
	}
}

func (m SubtaskDetailModel) loadCmd() tea.Cmd {
	return func() tea.Msg {
		subtask, err := m.subtaskSvc.GetByID(m.subtask.ID)
		if err != nil {
			return SubtaskDetailLoadedMsg{Err: err}
		}
		comments, err := m.subtaskCommentSvc.ListBySubtask(subtask.ID)
		if err != nil {
			return SubtaskDetailLoadedMsg{Err: err}
		}
		timer, _ := m.timeSvc.GetActiveTimer(m.currentUserID)
		totalMinutes := m.timeSvc.GetTotalMinutesForSubtask(subtask.ID)
		return SubtaskDetailLoadedMsg{
			Subtask:      *subtask,
			Comments:     comments,
			ActiveTimer:  timer,
			TotalMinutes: totalMinutes,
		}
	}
}

func (m SubtaskDetailModel) toggleTimerCmd() tea.Cmd {
	return func() tea.Msg {
		if m.activeTimer != nil && m.activeTimer.SubtaskID != nil && *m.activeTimer.SubtaskID == m.subtask.ID {
			_, err := m.timeSvc.StopTimer(m.currentUserID)
			if err != nil {
				return SubtaskTimerToggledMsg{Err: err}
			}
			return SubtaskTimerToggledMsg{ActiveTimer: nil}
		}
		subtaskID := m.subtask.ID
		timer, err := m.timeSvc.StartTimer(nil, &subtaskID, m.currentUserID)
		if err != nil {
			return SubtaskTimerToggledMsg{Err: err}
		}
		return SubtaskTimerToggledMsg{ActiveTimer: timer}
	}
}

func (m SubtaskDetailModel) Init() tea.Cmd {
	return m.loadCmd()
}

func (m SubtaskDetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(SubtaskDetailLoadedMsg); ok {
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.subtask = msg.Subtask
		m.comments = msg.Comments
		m.activeTimer = msg.ActiveTimer
		m.totalMinutes = msg.TotalMinutes
		if m.selectedComment >= len(m.comments) {
			m.selectedComment = 0
		}
		if m.activeTimer != nil && m.activeTimer.SubtaskID != nil && *m.activeTimer.SubtaskID == m.subtask.ID {
			return m, subtaskDetailTickCmd()
		}
		return m, nil
	}

	if msg, ok := msg.(SubtaskDetailTickMsg); ok {
		_ = msg
		return m, subtaskDetailTickCmd()
	}

	if msg, ok := msg.(SubtaskTimerToggledMsg); ok {
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.activeTimer = msg.ActiveTimer
		if m.activeTimer != nil {
			return m, subtaskDetailTickCmd()
		}
		m.loading = true
		return m, m.loadCmd()
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "e":
			subtask := m.subtask
			task := m.task
			project := m.project
			return m, func() tea.Msg {
				return NavigateMsg{To: screenEditSubtask, Subtask: subtask, Task: task, Project: project}
			}
		case "c":
			subtask := m.subtask
			task := m.task
			project := m.project
			return m, func() tea.Msg {
				return NavigateMsg{To: screenCreateSubtaskComment, Subtask: subtask, Task: task, Project: project}
			}
		case "T":
			return m, m.toggleTimerCmd()
		case "l":
			subtask := m.subtask
			task := m.task
			project := m.project
			return m, func() tea.Msg {
				return NavigateMsg{To: screenLogTime, Subtask: subtask, Task: task, Project: project}
			}
		case "up", "k":
			if len(m.comments) > 0 && m.selectedComment > 0 {
				m.selectedComment--
			}
		case "down", "j":
			if len(m.comments) > 0 && m.selectedComment < len(m.comments)-1 {
				m.selectedComment++
			}
		case "enter":
			if len(m.comments) > 0 {
				comment := m.comments[m.selectedComment]
				subtask := m.subtask
				task := m.task
				project := m.project
				return m, func() tea.Msg {
					return NavigateMsg{To: screenEditSubtaskComment, SubtaskComment: comment, Subtask: subtask, Task: task, Project: project}
				}
			}
		case "esc":
			task := m.task
			project := m.project
			return m, func() tea.Msg {
				return NavigateMsg{To: screenTaskDetail, Task: task, Project: project}
			}
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m SubtaskDetailModel) View() string {
	if m.loading {
		return titleStyle.Render("SprintOS — Subtask Detail") +
			"\n\n" + normalStyle.Render("Loading...") + "\n"
	}

	if m.err != nil {
		return titleStyle.Render("SprintOS — Subtask Detail") +
			"\n\n" + errorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())) + "\n"
	}

	done := "[ ]"
	if m.subtask.Done {
		done = "[x]"
	}

	s := titleStyle.Render(fmt.Sprintf("SprintOS — %s %s", done, m.subtask.Title)) + "\n\n"

	if m.subtask.Description != nil && *m.subtask.Description != "" {
		s += selectedStyle.Render("Description:") + "\n"
		s += normalStyle.Render(*m.subtask.Description) + "\n\n"
	}

	timerRunning := m.activeTimer != nil && m.activeTimer.SubtaskID != nil && *m.activeTimer.SubtaskID == m.subtask.ID
	s += selectedStyle.Render("Time Tracked:") + "\n"
	s += normalStyle.Render(fmt.Sprintf("  Total: %s", app.FormatMinutes(m.totalMinutes))) + "\n"
	if timerRunning {
		s += selectedStyle.Render(fmt.Sprintf("  ● Timer running: %s", app.FormatElapsed(m.activeTimer.StartedAt))) + "\n"
	}
	s += normalStyle.Render("  T start/stop timer  •  l log time manually") + "\n"

	s += "\n" + selectedStyle.Render("Comments:") + "\n"
	if len(m.comments) == 0 {
		s += normalStyle.Render("  No comments yet.") + "\n"
	} else {
		for i, c := range m.comments {
			line := fmt.Sprintf("  [%s] %s: %s",
				c.CreatedAt.Format("2006-01-02 15:04"),
				c.Author.Name,
				c.Content,
			)
			if i == m.selectedComment {
				s += selectedStyle.Render(line) + "\n"
			} else {
				s += normalStyle.Render(line) + "\n"
			}
		}
	}

	s += "\n" + normalStyle.Render(strings.Repeat("─", 40)) + "\n"
	s += normalStyle.Render("c comment  •  e edit  •  T timer  •  l log time  •  ↑/↓ select comment  •  enter edit comment  •  esc back  •  q quit") + "\n"

	return s
}
