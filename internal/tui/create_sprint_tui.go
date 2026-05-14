package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/domain"
)

type CreateSprintTUIModel struct {
	inputs    []textinput.Model
	focused   int
	loading   bool
	err       error
	project   domain.Project
	sprintSvc *app.SprintService
}

type SprintCreatedMsg struct {
	Sprint *domain.Sprint
	Err    error
}

func NewCreateSprintTUIModel(project domain.Project, sprintSvc *app.SprintService) CreateSprintTUIModel {
	inputs := make([]textinput.Model, 4)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Sprint 1"
	inputs[0].CharLimit = 100
	inputs[0].Focus()

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Sprint goal (optional)"
	inputs[1].CharLimit = 200

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "2025-06-01"
	inputs[2].CharLimit = 10

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "2025-06-14"
	inputs[3].CharLimit = 10

	return CreateSprintTUIModel{
		inputs:    inputs,
		project:   project,
		sprintSvc: sprintSvc,
	}
}

func (m CreateSprintTUIModel) submitCmd() tea.Cmd {
	return func() tea.Msg {
		name := m.inputs[0].Value()
		goal := m.inputs[1].Value()
		startStr := m.inputs[2].Value()
		endStr := m.inputs[3].Value()

		if name == "" {
			return SprintCreatedMsg{Err: fmt.Errorf("sprint name is required")}
		}

		start, err := time.Parse("2006-01-02", startStr)
		if err != nil {
			return SprintCreatedMsg{Err: fmt.Errorf("invalid start date — use YYYY-MM-DD")}
		}
		end, err := time.Parse("2006-01-02", endStr)
		if err != nil {
			return SprintCreatedMsg{Err: fmt.Errorf("invalid end date — use YYYY-MM-DD")}
		}
		if end.Before(start) {
			return SprintCreatedMsg{Err: fmt.Errorf("end date must be after start date")}
		}

		sprint, err := m.sprintSvc.Create(name, goal, m.project.ID, start, end)
		if err != nil {
			return SprintCreatedMsg{Err: err}
		}
		return SprintCreatedMsg{Sprint: sprint}
	}
}

func (m CreateSprintTUIModel) Init() tea.Cmd { return textinput.Blink }

func (m CreateSprintTUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(SprintCreatedMsg); ok {
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		project := m.project
		return m, func() tea.Msg {
			return NavigateMsg{To: screenSprintView, Project: project}
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
				return NavigateMsg{To: screenSprintView, Project: project}
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

func (m CreateSprintTUIModel) View() string {
	if m.loading {
		return titleStyle.Render("SprintOS — Create Sprint") + "\n\n" +
			normalStyle.Render("Creating sprint...") + "\n"
	}

	labels := []string{"Sprint name *", "Goal (optional)", "Start date * (YYYY-MM-DD)", "End date * (YYYY-MM-DD)"}
	s := titleStyle.Render(fmt.Sprintf("SprintOS — New Sprint: %s", m.project.Name)) + "\n\n"

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

	s += normalStyle.Render("tab/↓ next  •  shift+tab/↑ prev  •  enter confirm  •  esc back") + "\n"
	return s
}
