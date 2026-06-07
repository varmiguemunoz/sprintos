package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/domain"
)

type EditTaskModel struct {
	titleInput textinput.Model
	descInput  textarea.Model
	focused    int
	loading    bool
	saved      bool
	err        error
	task       domain.Task
	project    domain.Project
	taskSvc    *app.TaskService
}

type TaskUpdatedMsg struct {
	Err error
}

func NewEditTaskModel(task domain.Task, project domain.Project, taskSvc *app.TaskService) EditTaskModel {
	ti := textinput.New()
	ti.Placeholder = "Task title"
	ti.CharLimit = 150
	ti.SetValue(task.Title)
	ti.Focus()

	ta := textarea.New()
	ta.Placeholder = "Description (optional)"
	ta.CharLimit = 1000
	ta.SetWidth(70)
	ta.SetHeight(5)
	ta.ShowLineNumbers = false
	if task.Description != nil {
		ta.SetValue(*task.Description)
	}

	return EditTaskModel{
		titleInput: ti,
		descInput:  ta,
		focused:    0,
		task:       task,
		project:    project,
		taskSvc:    taskSvc,
	}
}

func (m EditTaskModel) saveCmd() tea.Cmd {
	return func() tea.Msg {
		title := m.titleInput.Value()
		description := m.descInput.Value()

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
		case "ctrl+s":
			m.loading = true
			m.saved = false
			return m, m.saveCmd()
		case "tab":
			cmd := m.focusNext()
			return m, cmd
		case "shift+tab":
			cmd := m.focusPrev()
			return m, cmd
		case "up":
			if m.focused != 1 {
				m.saved = false
				cmd := m.focusPrev()
				return m, cmd
			}
		case "down":
			if m.focused != 1 {
				m.saved = false
				cmd := m.focusNext()
				return m, cmd
			}
		case "enter":
			if m.focused == 0 {
				m.saved = false
				cmd := m.focusNext()
				return m, cmd
			}
		}
	}

	var cmd tea.Cmd
	switch m.focused {
	case 0:
		m.titleInput, cmd = m.titleInput.Update(msg)
	case 1:
		m.descInput, cmd = m.descInput.Update(msg)
	}
	return m, cmd
}

func (m *EditTaskModel) focusNext() tea.Cmd {
	if m.focused == 0 {
		m.titleInput.Blur()
		m.focused = 1
		return m.descInput.Focus()
	}
	m.descInput.Blur()
	m.focused = 0
	m.titleInput.Focus()
	return nil
}

func (m *EditTaskModel) focusPrev() tea.Cmd {
	if m.focused == 1 {
		m.descInput.Blur()
		m.focused = 0
		m.titleInput.Focus()
		return nil
	}
	m.titleInput.Blur()
	m.focused = 1
	return m.descInput.Focus()
}

func (m EditTaskModel) View() string {
	if m.loading {
		return titleStyle.Render("SprintOS — Edit Task") +
			"\n\n" + normalStyle.Render("Saving...") + "\n"
	}

	s := titleStyle.Render(fmt.Sprintf("SprintOS — Edit #%d: %s", m.task.TaskNumber, m.task.Title)) + "\n\n"

	if m.focused == 0 {
		s += selectedStyle.Render("Title *") + "\n"
	} else {
		s += dimStyle.Render("Title *") + "\n"
	}
	s += m.titleInput.View() + "\n\n"

	if m.focused == 1 {
		s += selectedStyle.Render("Description") + "\n"
	} else {
		s += dimStyle.Render("Description") + "\n"
	}
	s += m.descInput.View() + "\n\n"

	if m.err != nil {
		s += errorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())) + "\n\n"
	}

	if m.saved {
		s += successStyle.Render("✓ Task updated successfully") + "\n\n"
	}

	s += dimStyle.Render("tab next  •  shift+tab prev  •  ctrl+s save  •  esc back") + "\n"
	return s
}
