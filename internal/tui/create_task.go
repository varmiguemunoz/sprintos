package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/domain"
)

type CreateTaskModel struct {
	titleInput   textinput.Model
	descInput    textarea.Model
	dueDateInput textinput.Model
	focused      int
	loading      bool
	err          error
	stateID      uint
	projectID    uint
	createdByID  uint
	project      domain.Project
	taskSvc      *app.TaskService
}

type TaskCreatedMsg struct {
	Err error
}

func NewCreateTaskModel(
	stateID uint,
	projectID uint,
	createdByID uint,
	project domain.Project,
	taskSvc *app.TaskService,
) CreateTaskModel {
	ti := textinput.New()
	ti.Placeholder = "Task title"
	ti.CharLimit = 150
	ti.Focus()

	ta := textarea.New()
	ta.Placeholder = "Description (optional)"
	ta.CharLimit = 1000
	ta.SetWidth(70)
	ta.SetHeight(5)
	ta.ShowLineNumbers = false

	due := textinput.New()
	due.Placeholder = "Due date: YYYY-MM-DD (optional)"
	due.CharLimit = 10

	return CreateTaskModel{
		titleInput:   ti,
		descInput:    ta,
		dueDateInput: due,
		focused:      0,
		stateID:      stateID,
		projectID:    projectID,
		createdByID:  createdByID,
		project:      project,
		taskSvc:      taskSvc,
	}
}

func (m CreateTaskModel) submitCmd() tea.Cmd {
	return func() tea.Msg {
		title := m.titleInput.Value()
		description := m.descInput.Value()
		dueDateStr := m.dueDateInput.Value()

		if title == "" {
			return TaskCreatedMsg{Err: fmt.Errorf("title is required")}
		}

		var dueDate *time.Time
		if dueDateStr != "" {
			parsed, err := time.Parse("2006-01-02", dueDateStr)
			if err != nil {
				return TaskCreatedMsg{Err: fmt.Errorf("invalid due date format, use YYYY-MM-DD")}
			}
			dueDate = &parsed
		}

		_, err := m.taskSvc.Create(
			title,
			description,
			m.stateID,
			m.projectID,
			m.createdByID,
			nil,
			nil,
			dueDate,
		)
		if err != nil {
			return TaskCreatedMsg{Err: err}
		}

		return TaskCreatedMsg{}
	}
}

func (m CreateTaskModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m CreateTaskModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(TaskCreatedMsg); ok {
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		project := m.project
		return m, func() tea.Msg {
			return NavigateMsg{To: screenKanban, Project: project}
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
			project := m.project
			return m, func() tea.Msg {
				return NavigateMsg{To: screenKanban, Project: project}
			}
		case "ctrl+s":
			m.loading = true
			return m, m.submitCmd()
		case "tab":
			cmd := m.focusNext()
			return m, cmd
		case "shift+tab":
			cmd := m.focusPrev()
			return m, cmd
		case "up":
			if m.focused != 1 {
				cmd := m.focusPrev()
				return m, cmd
			}
		case "down":
			if m.focused != 1 {
				cmd := m.focusNext()
				return m, cmd
			}
		case "enter":
			if m.focused == 2 {
				m.loading = true
				return m, m.submitCmd()
			}
			if m.focused == 0 {
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
	case 2:
		m.dueDateInput, cmd = m.dueDateInput.Update(msg)
	}
	return m, cmd
}

func (m *CreateTaskModel) focusNext() tea.Cmd {
	switch m.focused {
	case 0:
		m.titleInput.Blur()
		m.focused = 1
		return m.descInput.Focus()
	case 1:
		m.descInput.Blur()
		m.focused = 2
		m.dueDateInput.Focus()
	case 2:
		m.dueDateInput.Blur()
		m.focused = 0
		m.titleInput.Focus()
	}
	return nil
}

func (m *CreateTaskModel) focusPrev() tea.Cmd {
	switch m.focused {
	case 0:
		m.titleInput.Blur()
		m.focused = 2
		m.dueDateInput.Focus()
	case 1:
		m.descInput.Blur()
		m.focused = 0
		m.titleInput.Focus()
	case 2:
		m.dueDateInput.Blur()
		m.focused = 1
		return m.descInput.Focus()
	}
	return nil
}

func (m CreateTaskModel) View() string {
	if m.loading {
		return titleStyle.Render("SprintOS — Create Task") +
			"\n\n" + normalStyle.Render("Creating task...") + "\n"
	}

	s := titleStyle.Render("SprintOS — Create Task") + "\n\n"

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

	if m.focused == 2 {
		s += selectedStyle.Render("Due Date") + "\n"
	} else {
		s += dimStyle.Render("Due Date") + "\n"
	}
	s += m.dueDateInput.View() + "\n\n"

	if m.err != nil {
		s += errorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())) + "\n\n"
	}

	s += dimStyle.Render("tab next  •  shift+tab prev  •  enter confirm  •  ctrl+s save  •  esc cancel") + "\n"
	return s
}
