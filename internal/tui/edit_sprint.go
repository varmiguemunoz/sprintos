package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/domain"
)

type EditSprintModel struct {
	inputs    []textinput.Model
	focused   int
	loading   bool
	saved     bool
	err       error
	sprint    domain.Sprint
	project   domain.Project
	sprintSvc *app.SprintService
}

type SprintUpdatedMsg struct {
	Err error
}

func NewEditSprintModel(sprint domain.Sprint, project domain.Project, sprintSvc *app.SprintService) EditSprintModel {
	inputs := make([]textinput.Model, 4)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Sprint name"
	inputs[0].CharLimit = 100
	inputs[0].SetValue(sprint.Name)
	inputs[0].Focus()

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Goal (optional)"
	inputs[1].CharLimit = 200
	if sprint.Goal != nil {
		inputs[1].SetValue(*sprint.Goal)
	}

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "YYYY-MM-DD"
	inputs[2].CharLimit = 10
	inputs[2].SetValue(sprint.StartDate.Format("2006-01-02"))

	inputs[3] = textinput.New()
	inputs[3].Placeholder = "YYYY-MM-DD"
	inputs[3].CharLimit = 10
	inputs[3].SetValue(sprint.EndDate.Format("2006-01-02"))

	return EditSprintModel{
		inputs:    inputs,
		sprint:    sprint,
		project:   project,
		sprintSvc: sprintSvc,
	}
}

func (m EditSprintModel) saveCmd() tea.Cmd {
	return func() tea.Msg {
		name := m.inputs[0].Value()
		goal := m.inputs[1].Value()
		startStr := m.inputs[2].Value()
		endStr := m.inputs[3].Value()

		if name == "" {
			return SprintUpdatedMsg{Err: fmt.Errorf("sprint name is required")}
		}

		start, err := time.Parse("2006-01-02", startStr)
		if err != nil {
			return SprintUpdatedMsg{Err: fmt.Errorf("invalid start date — use YYYY-MM-DD")}
		}
		end, err := time.Parse("2006-01-02", endStr)
		if err != nil {
			return SprintUpdatedMsg{Err: fmt.Errorf("invalid end date — use YYYY-MM-DD")}
		}
		if end.Before(start) {
			return SprintUpdatedMsg{Err: fmt.Errorf("end date must be after start date")}
		}

		_, err = m.sprintSvc.Update(m.sprint.ID, name, goal, start, end)
		if err != nil {
			return SprintUpdatedMsg{Err: err}
		}

		return SprintUpdatedMsg{}
	}
}

func (m EditSprintModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m EditSprintModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(SprintUpdatedMsg); ok {
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

func (m EditSprintModel) View() string {
	if m.loading {
		return titleStyle.Render("SprintOS — Edit Sprint") +
			"\n\n" + normalStyle.Render("Saving...") + "\n"
	}

	labels := []string{"Sprint name *", "Goal (optional)", "Start date * (YYYY-MM-DD)", "End date * (YYYY-MM-DD)"}
	s := titleStyle.Render(fmt.Sprintf("SprintOS — Edit Sprint: %s", m.sprint.Name)) + "\n\n"

	for i, label := range labels {
		if i == m.focused {
			s += selectedStyle.Render(label) + "\n"
		} else {
			s += dimStyle.Render(label) + "\n"
		}
		s += m.inputs[i].View() + "\n\n"
	}

	if m.err != nil {
		s += errorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())) + "\n\n"
	}

	if m.saved {
		s += successStyle.Render("✓ Sprint updated successfully") + "\n\n"
	}

	s += dimStyle.Render("tab/↓ next  •  shift+tab/↑ prev  •  enter save  •  esc back") + "\n"
	return s
}
