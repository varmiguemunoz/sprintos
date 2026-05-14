package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/domain"
)

type EditTaskModel struct {
	inputs  []textinput.Model
	focused int
	loading bool
	saved   bool
	err     error
	task    domain.Task
	project domain.Project
	taskSvc *app.TaskService
}

type TaskUpdatedMsg struct {
	Err error
}

func NewEditTaskModel(task domain.Task, project domain.Project, taskSvc *app.TaskService) EditTaskModel {
	inputs := make([]textinput.Model, 2)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Task title"
	inputs[0].CharLimit = 150
	inputs[0].SetValue(task.Title)
	inputs[0].Focus()

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Description (optional)"
	inputs[1].CharLimit = 500
	if task.Description != nil {
		inputs[1].SetValue(*task.Description)
	}

	return EditTaskModel{
		inputs:  inputs,
		task:    task,
		project: project,
		taskSvc: taskSvc,
	}
}

func (m EditTaskModel) saveCmd() tea.Cmd {
	return func() tea.Msg {
		title := m.inputs[0].Value()
		description := m.inputs[1].Value()

		if title == "" {
			return TaskUpdatedMsg{Err: fmt.Errorf("task title is required")}
		}

		_, err := m.taskSvc.Update(
			m.task.ID,
			title,
			description,
			m.task.AssignedTo,
			m.task.StartDate,
			m.task.DueDate,
		)
		if err != nil {
			return TaskUpdatedMsg{Err: err}
		}

		return TaskUpdatedMsg{}
	}
}

func (m EditTaskModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m EditTaskModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(TaskUpdatedMsg); ok {
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.saved = true
		m.err = nil
		return m, nil
	}

	if m.loading {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			task := m.task
			project := m.project
			return m, func() tea.Msg {
				return NavigateMsg{To: screenTaskDetail, Task: task, Project: project}
			}
		case "ctrl+c":
			return m, tea.Quit
		case "tab", "down":
			m.inputs[m.focused].Blur()
			m.focused = (m.focused + 1) % len(m.inputs)
			m.inputs[m.focused].Focus()
			m.saved = false
		case "shift+tab", "up":
			m.inputs[m.focused].Blur()
			m.focused--
			if m.focused < 0 {
				m.focused = len(m.inputs) - 1
			}
			m.inputs[m.focused].Focus()
			m.saved = false
		case "enter":
			if m.focused == len(m.inputs)-1 {
				m.loading = true
				return m, m.saveCmd()
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

func (m EditTaskModel) View() string {
	if m.loading {
		return titleStyle.Render("SprintOS — Edit Task") +
			"\n\n" + normalStyle.Render("Saving...") + "\n"
	}

	labels := []string{"Title *", "Description"}

	s := titleStyle.Render(fmt.Sprintf("SprintOS — Edit #%d: %s", m.task.TaskNumber, m.task.Title)) + "\n\n"

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
		s += selectedStyle.Render("✓ Task updated successfully") + "\n\n"
	}

	s += normalStyle.Render("tab/↓ next  •  shift+tab/↑ previous  •  enter save  •  esc back") + "\n"
	return s
}
