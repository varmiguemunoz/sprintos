package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/domain"
)

type boardMode int

const (
	boardModeTemplate boardMode = iota
	boardModeCustom
)

type boardField int

const (
	fieldName  boardField = iota
	fieldColor
	fieldDone
)

type customStateEntry struct {
	Name  string
	Color string
	IsDone bool
}

var colorPalette = []struct{ Name, Hex string }{
	{"Gray", "#6B7280"},
	{"Blue", "#3B82F6"},
	{"Yellow", "#F59E0B"},
	{"Green", "#10B981"},
	{"Red", "#EF4444"},
	{"Purple", "#8B5CF6"},
	{"Pink", "#EC4899"},
	{"Orange", "#F97316"},
}

type BoardSetupModel struct {
	mode           boardMode
	templateCursor int
	templates      []app.TemplatePreview
	customStates   []customStateEntry
	nameInput      textinput.Model
	colorIdx       int
	isDone         bool
	activeField    boardField
	project        domain.Project
	isEditing      bool
	loading        bool
	err            error
	stateSvc       *app.StateService
	taskSvc        *app.TaskService
}

type BoardSetupDoneMsg struct {
	Err error
}

type ExistingStatesLoadedMsg struct {
	States []domain.State
	Err    error
}

func NewBoardSetupModel(
	project domain.Project,
	isEditing bool,
	stateSvc *app.StateService,
	taskSvc *app.TaskService,
) BoardSetupModel {
	input := textinput.New()
	input.Placeholder = "State name (e.g. In Review)"
	input.CharLimit = 50
	input.Focus()

	m := BoardSetupModel{
		project:    project,
		isEditing:  isEditing,
		templates:  app.ListTemplates(),
		nameInput:  input,
		stateSvc:   stateSvc,
		taskSvc:    taskSvc,
	}

	if isEditing {
		m.mode = boardModeCustom
		m.loading = true
	} else {
		m.mode = boardModeTemplate
	}

	return m
}

func (m BoardSetupModel) loadExistingStatesCmd() tea.Cmd {
	return func() tea.Msg {
		states, err := m.stateSvc.ListByProject(m.project.ID)
		return ExistingStatesLoadedMsg{States: states, Err: err}
	}
}

func (m BoardSetupModel) saveCmd() tea.Cmd {
	return func() tea.Msg {
		// Delete all existing states
		existing, _ := m.stateSvc.ListByProject(m.project.ID)
		for _, s := range existing {
			_ = m.stateSvc.Delete(s.ID)
		}

		// Create new states
		var firstStateID uint
		for i, entry := range m.customStates {
			state, err := m.stateSvc.Create(
				entry.Name, entry.Color, uint(i+1), entry.IsDone, m.project.ID,
			)
			if err != nil {
				return BoardSetupDoneMsg{Err: err}
			}
			if i == 0 {
				firstStateID = state.ID
			}
		}

		// Move all tasks to first state (business rule)
		if firstStateID > 0 {
			tasks, _ := m.taskSvc.ListByProject(m.project.ID)
			for _, task := range tasks {
				_, _ = m.taskSvc.MoveState(task.ID, firstStateID)
			}
		}

		return BoardSetupDoneMsg{}
	}
}

func (m BoardSetupModel) applyTemplateCmd(templateKey string) tea.Cmd {
	return func() tea.Msg {
		existing, _ := m.stateSvc.ListByProject(m.project.ID)
		for _, s := range existing {
			_ = m.stateSvc.Delete(s.ID)
		}

		if err := m.stateSvc.ApplyTemplate(m.project.ID, templateKey); err != nil {
			return BoardSetupDoneMsg{Err: err}
		}

		return BoardSetupDoneMsg{}
	}
}

func (m BoardSetupModel) Init() tea.Cmd {
	if m.isEditing {
		return m.loadExistingStatesCmd()
	}
	return textinput.Blink
}

func (m BoardSetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(ExistingStatesLoadedMsg); ok {
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		for _, s := range msg.States {
			m.customStates = append(m.customStates, customStateEntry{
				Name:   s.Name,
				Color:  s.Color,
				IsDone: s.IsDone,
			})
		}
		return m, textinput.Blink
	}

	if msg, ok := msg.(BoardSetupDoneMsg); ok {
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		project := m.project
		return m, func() tea.Msg {
			return NavigateMsg{To: screenDashboard, Project: project}
		}
	}

	if m.loading {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.mode == boardModeTemplate {
			switch msg.String() {
			case "up", "k":
				if m.templateCursor > 0 {
					m.templateCursor--
				}
			case "down", "j":
				// +1 for the Custom option
				if m.templateCursor < len(m.templates) {
					m.templateCursor++
				}
			case "enter":
				if m.templateCursor == len(m.templates) {
					m.mode = boardModeCustom
					m.nameInput.Focus()
					return m, textinput.Blink
				}
				m.loading = true
				key := m.templates[m.templateCursor].Key
				return m, m.applyTemplateCmd(key)
			case "esc", "ctrl+c":
				return m, tea.Quit
			}
			return m, nil
		}

		// Custom mode
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.mode == boardModeCustom && !m.isEditing {
				m.mode = boardModeTemplate
				return m, nil
			}
			project := m.project
			return m, func() tea.Msg {
				return NavigateMsg{To: screenDashboard, Project: project}
			}
		case "tab":
			m.activeField = (m.activeField + 1) % 3
			if m.activeField == fieldName {
				m.nameInput.Focus()
			} else {
				m.nameInput.Blur()
			}
		case "shift+tab":
			if m.activeField == 0 {
				m.activeField = fieldDone
			} else {
				m.activeField--
			}
			if m.activeField == fieldName {
				m.nameInput.Focus()
			} else {
				m.nameInput.Blur()
			}
		case "left", "h":
			if m.activeField == fieldColor && m.colorIdx > 0 {
				m.colorIdx--
			}
		case "right", "l":
			if m.activeField == fieldColor && m.colorIdx < len(colorPalette)-1 {
				m.colorIdx++
			}
		case "y", "Y":
			if m.activeField == fieldDone {
				m.isDone = true
			}
		case "n", "N":
			if m.activeField == fieldDone {
				m.isDone = false
			}
		case "enter":
			if m.activeField == fieldDone || m.activeField == fieldColor {
				name := strings.TrimSpace(m.nameInput.Value())
				if name == "" {
					m.err = fmt.Errorf("state name cannot be empty")
					return m, nil
				}
				m.customStates = append(m.customStates, customStateEntry{
					Name:   name,
					Color:  colorPalette[m.colorIdx].Hex,
					IsDone: m.isDone,
				})
				m.nameInput.SetValue("")
				m.nameInput.Focus()
				m.activeField = fieldName
				m.isDone = false
				m.err = nil
				return m, nil
			}
		case "ctrl+s":
			if len(m.customStates) == 0 {
				m.err = fmt.Errorf("add at least one state before saving")
				return m, nil
			}
			m.loading = true
			return m, m.saveCmd()
		case "backspace":
			if m.activeField == fieldName && m.nameInput.Value() == "" && len(m.customStates) > 0 {
				m.customStates = m.customStates[:len(m.customStates)-1]
				return m, nil
			}
		}
	}

	if m.mode == boardModeCustom && m.activeField == fieldName {
		var cmd tea.Cmd
		m.nameInput, cmd = m.nameInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m BoardSetupModel) View() string {
	if m.loading {
		return titleStyle.Render("SprintOS — Board Setup") +
			"\n\n" + normalStyle.Render("Setting up your board...") + "\n"
	}

	if m.mode == boardModeTemplate {
		s := titleStyle.Render(fmt.Sprintf("SprintOS — Board Setup: %s", m.project.Name)) + "\n\n"
		s += normalStyle.Render("Choose a template:") + "\n\n"

		for i, tmpl := range m.templates {
			preview := strings.Join(tmpl.States, " → ")
			if i == m.templateCursor {
				s += selectedStyle.Render(fmt.Sprintf("> %s", tmpl.Name)) + "\n"
				s += normalStyle.Render(fmt.Sprintf("  %s", preview)) + "\n\n"
			} else {
				s += normalStyle.Render(fmt.Sprintf("  %s", tmpl.Name)) + "\n"
				s += normalStyle.Render(fmt.Sprintf("  %s", preview)) + "\n\n"
			}
		}

		if m.templateCursor == len(m.templates) {
			s += selectedStyle.Render("> Custom") + "\n"
			s += normalStyle.Render("  Build your own board") + "\n\n"
		} else {
			s += normalStyle.Render("  Custom") + "\n"
			s += normalStyle.Render("  Build your own board") + "\n\n"
		}

		s += normalStyle.Render("↑/↓ move  •  enter to select") + "\n"
		return s
	}

	// Custom mode
	s := titleStyle.Render(fmt.Sprintf("SprintOS — Custom Board: %s", m.project.Name)) + "\n\n"

	if len(m.customStates) > 0 {
		s += selectedStyle.Render("States added:") + "\n"
		for i, st := range m.customStates {
			dot := lipgloss.NewStyle().Foreground(lipgloss.Color(st.Color)).Render("■")
			done := ""
			if st.IsDone {
				done = " (done)"
			}
			s += normalStyle.Render(fmt.Sprintf("  %d. %s %s%s", i+1, dot, st.Name, done)) + "\n"
		}
		s += "\n"
	}

	s += selectedStyle.Render("Add a state:") + "\n\n"

	if m.activeField == fieldName {
		s += selectedStyle.Render("Name *") + "\n"
	} else {
		s += normalStyle.Render("Name *") + "\n"
	}
	s += m.nameInput.View() + "\n\n"

	if m.activeField == fieldColor {
		s += selectedStyle.Render("Color") + "\n"
	} else {
		s += normalStyle.Render("Color") + "\n"
	}
	colorRow := ""
	for i, c := range colorPalette {
		swatch := lipgloss.NewStyle().Foreground(lipgloss.Color(c.Hex)).Render("■ " + c.Name)
		if i == m.colorIdx {
			colorRow += selectedStyle.Render("[" + swatch + "]") + " "
		} else {
			colorRow += swatch + " "
		}
	}
	s += colorRow + "\n\n"

	if m.activeField == fieldDone {
		s += selectedStyle.Render("Mark as Done state? ") + normalStyle.Render(fmt.Sprintf("[y/n] current: %v", m.isDone)) + "\n\n"
	} else {
		s += normalStyle.Render(fmt.Sprintf("Mark as Done state? [y/n] current: %v", m.isDone)) + "\n\n"
	}

	if m.err != nil {
		s += errorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())) + "\n\n"
	}

	s += normalStyle.Render("tab next field  •  enter add state  •  backspace remove last  •  ctrl+s save  •  esc back") + "\n"
	return s
}
