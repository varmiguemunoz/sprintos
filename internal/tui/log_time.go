package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/domain"
)

type LogTimeModel struct {
	inputs  []textinput.Model
	focused int
	loading bool
	saved   bool
	err     error
	task    *domain.Task
	subtask *domain.Subtask
	project domain.Project
	userID  uint
	timeSvc *app.TimeEntryService
}

type TimeLoggedMsg struct {
	Err error
}

func NewLogTimeModel(
	task *domain.Task,
	subtask *domain.Subtask,
	project domain.Project,
	userID uint,
	timeSvc *app.TimeEntryService,
) LogTimeModel {
	inputs := make([]textinput.Model, 2)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Minutes (e.g. 90)"
	inputs[0].CharLimit = 6
	inputs[0].Focus()

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Note (optional)"
	inputs[1].CharLimit = 200

	return LogTimeModel{
		inputs:  inputs,
		task:    task,
		subtask: subtask,
		project: project,
		userID:  userID,
		timeSvc: timeSvc,
	}
}

func (m LogTimeModel) submitCmd() tea.Cmd {
	return func() tea.Msg {
		raw := strings.TrimSpace(m.inputs[0].Value())
		if raw == "" {
			return TimeLoggedMsg{Err: fmt.Errorf("minutes is required")}
		}

		minutes, err := strconv.Atoi(raw)
		if err != nil || minutes <= 0 {
			return TimeLoggedMsg{Err: fmt.Errorf("minutes must be a positive number")}
		}

		note := strings.TrimSpace(m.inputs[1].Value())

		var taskID *uint
		var subtaskID *uint

		if m.subtask != nil {
			id := m.subtask.ID
			subtaskID = &id
		} else if m.task != nil {
			id := m.task.ID
			taskID = &id
		}

		_, err = m.timeSvc.LogManual(taskID, subtaskID, m.userID, minutes, note)
		if err != nil {
			return TimeLoggedMsg{Err: err}
		}

		return TimeLoggedMsg{}
	}
}

func (m LogTimeModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m LogTimeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(TimeLoggedMsg); ok {
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.saved = true
		if m.subtask != nil {
			subtask := *m.subtask
			task := *m.task
			project := m.project
			return m, func() tea.Msg {
				return NavigateMsg{To: screenSubtaskDetail, Subtask: subtask, Task: task, Project: project}
			}
		}
		task := *m.task
		project := m.project
		return m, func() tea.Msg {
			return NavigateMsg{To: screenTaskDetail, Task: task, Project: project}
		}
	}

	if m.loading {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.subtask != nil {
				subtask := *m.subtask
				task := *m.task
				project := m.project
				return m, func() tea.Msg {
					return NavigateMsg{To: screenSubtaskDetail, Subtask: subtask, Task: task, Project: project}
				}
			}
			task := *m.task
			project := m.project
			return m, func() tea.Msg {
				return NavigateMsg{To: screenTaskDetail, Task: task, Project: project}
			}
		case "tab", "down":
			m.inputs[m.focused].Blur()
			m.focused = (m.focused + 1) % len(m.inputs)
			m.inputs[m.focused].Focus()
		case "shift+tab", "up":
			m.inputs[m.focused].Blur()
			m.focused--
			if m.focused < 0 {
				m.focused = len(m.inputs) - 1
			}
			m.inputs[m.focused].Focus()
		case "enter":
			if m.focused == len(m.inputs)-1 {
				m.loading = true
				return m, m.submitCmd()
			}
			m.inputs[m.focused].Blur()
			m.focused++
			m.inputs[m.focused].Focus()
		}
	}

	var cmd tea.Cmd
	m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
	return m, cmd
}

func (m LogTimeModel) View() string {
	if m.loading {
		return titleStyle.Render("SprintOS — Log Time") +
			"\n\n" + normalStyle.Render("Saving...") + "\n"
	}

	context := ""
	if m.subtask != nil {
		context = fmt.Sprintf("Subtask: %s", m.subtask.Title)
	} else if m.task != nil {
		context = fmt.Sprintf("Task: %s", m.task.Title)
	}

	labels := []string{"Minutes *", "Note"}

	s := titleStyle.Render("SprintOS — Log Time") + "\n\n"
	s += selectedStyle.Render(context) + "\n\n"

	for i, label := range labels {
		if i == m.focused {
			s += selectedStyle.Render(label) + "\n"
		} else {
			s += normalStyle.Render(label) + "\n"
		}
		s += m.inputs[i].View() + "\n\n"
	}

	if m.err != nil {
		s += errorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())) + "\n\n"
	}

	if m.saved {
		s += selectedStyle.Render("✓ Time logged") + "\n\n"
	}

	s += normalStyle.Render("tab/↓ next  •  shift+tab/↑ previous  •  enter save  •  esc back") + "\n"
	return s
}
