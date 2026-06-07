package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/domain"
	"github.com/varmiguemunoz/sprintos/internal/report"
)

type exportStep int

const (
	exportStepProjects exportStep = iota
	exportStepRange
	exportStepGenerating
	exportStepDone
)

type timeRangeOption int

const (
	rangeWeek timeRangeOption = iota
	rangeMonth
	rangeThreeMonths
	rangeCustom
)

type ExportReportModel struct {
	step            exportStep
	projects        []domain.Project
	selectedAll     bool
	checkedProjects map[uint]bool
	projectCursor   int

	rangeOption  timeRangeOption
	rangeCursor  int
	customFrom   textinput.Model
	customTo     textinput.Model
	focusedInput int

	spinner   spinner.Model
	savedPath string
	err       error

	currentOrgID uint
	reportSvc    *app.ReportService
	projectSvc   *app.ProjectService
	windowWidth  int
}

type exportProjectsLoadedMsg struct {
	Projects []domain.Project
	Err      error
}

type exportDoneMsg struct {
	Path string
	Err  error
}

func NewExportReportModel(
	orgID uint,
	reportSvc *app.ReportService,
	projectSvc *app.ProjectService,
) ExportReportModel {
	fromInput := textinput.New()
	fromInput.Placeholder = "YYYY-MM-DD"
	fromInput.CharLimit = 10
	fromInput.Width = 14

	toInput := textinput.New()
	toInput.Placeholder = "YYYY-MM-DD"
	toInput.CharLimit = 10
	toInput.Width = 14

	s := spinner.New()
	s.Spinner = spinner.Dot

	return ExportReportModel{
		step:            exportStepProjects,
		selectedAll:     true,
		checkedProjects: make(map[uint]bool),
		customFrom:      fromInput,
		customTo:        toInput,
		spinner:         s,
		currentOrgID:    orgID,
		reportSvc:       reportSvc,
		projectSvc:      projectSvc,
		windowWidth:     90,
	}
}

func (m ExportReportModel) Init() tea.Cmd {
	return m.loadProjectsCmd()
}

func (m ExportReportModel) loadProjectsCmd() tea.Cmd {
	return func() tea.Msg {
		projects, err := m.projectSvc.ListByOrganization(m.currentOrgID)
		return exportProjectsLoadedMsg{Projects: projects, Err: err}
	}
}

func (m ExportReportModel) generateCmd() tea.Cmd {
	return func() tea.Msg {
		var projectIDs []uint
		if !m.selectedAll {
			for id, checked := range m.checkedProjects {
				if checked {
					projectIDs = append(projectIDs, id)
				}
			}
		}

		from, to, err := m.resolveTimeRange()
		if err != nil {
			return exportDoneMsg{Err: err}
		}

		params := app.ReportParams{
			OrgID:      m.currentOrgID,
			ProjectIDs: projectIDs,
			From:       from,
			To:         to,
		}

		data, err := m.reportSvc.Generate(params)
		if err != nil {
			return exportDoneMsg{Err: fmt.Errorf("data error: %w", err)}
		}

		home, _ := os.UserHomeDir()
		filename := fmt.Sprintf("sprintos-report-%s.pdf", time.Now().Format("2006-01-02"))
		destPath := filepath.Join(home, "Desktop", filename)

		if err := report.Generate(data, destPath); err != nil {
			return exportDoneMsg{Err: fmt.Errorf("PDF error: %w", err)}
		}

		return exportDoneMsg{Path: destPath}
	}
}

func (m ExportReportModel) resolveTimeRange() (time.Time, time.Time, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	switch m.rangeOption {
	case rangeWeek:
		return today.Add(-7 * 24 * time.Hour), today, nil
	case rangeMonth:
		return today.AddDate(0, -1, 0), today, nil
	case rangeThreeMonths:
		return today.AddDate(0, -3, 0), today, nil
	case rangeCustom:
		from, err := time.ParseInLocation("2006-01-02", m.customFrom.Value(), now.Location())
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid 'from' date: use YYYY-MM-DD")
		}
		to, err := time.ParseInLocation("2006-01-02", m.customTo.Value(), now.Location())
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid 'to' date: use YYYY-MM-DD")
		}
		if to.Before(from) {
			return time.Time{}, time.Time{}, fmt.Errorf("'to' date must be after 'from' date")
		}
		return from, to, nil
	}
	return today.Add(-30 * 24 * time.Hour), today, nil
}

func (m ExportReportModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(exportProjectsLoadedMsg); ok {
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.projects = msg.Projects
		return m, nil
	}

	if msg, ok := msg.(exportDoneMsg); ok {
		m.step = exportStepDone
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.savedPath = msg.Path
		}
		return m, nil
	}

	if msg, ok := msg.(spinner.TickMsg); ok {
		if m.step == exportStepGenerating {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		_ = msg
		return m, nil
	}

	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.windowWidth = msg.Width
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.step {
		case exportStepProjects:
			return m.updateProjectStep(msg)
		case exportStepRange:
			return m.updateRangeStep(msg)
		case exportStepDone:
			return m.updateDoneStep(msg)
		}
	}

	return m, nil
}

func (m ExportReportModel) updateProjectStep(msg tea.KeyMsg) (ExportReportModel, tea.Cmd) {
	totalItems := len(m.projects) + 1

	switch msg.String() {
	case "up", "k":
		if m.projectCursor > 0 {
			m.projectCursor--
		}
	case "down", "j":
		if m.projectCursor < totalItems-1 {
			m.projectCursor++
		}
	case " ":
		if m.projectCursor == 0 {
			m.selectedAll = !m.selectedAll
		} else {
			idx := m.projectCursor - 1
			if idx < len(m.projects) {
				pid := m.projects[idx].ID
				m.checkedProjects[pid] = !m.checkedProjects[pid]
				m.selectedAll = false
			}
		}
	case "enter":
		if m.selectedAll || m.hasAnyChecked() {
			m.step = exportStepRange
		}
	case "esc":
		return m, func() tea.Msg {
			return NavigateMsg{To: screenDashboard}
		}
	case "q", "ctrl+c":
		return m, tea.Quit
	}

	return m, nil
}

func (m ExportReportModel) updateRangeStep(msg tea.KeyMsg) (ExportReportModel, tea.Cmd) {
	if m.rangeOption == rangeCustom {
		if m.focusedInput == 0 {
			if msg.String() == "tab" || msg.String() == "enter" {
				m.customFrom.Blur()
				m.customTo.Focus()
				m.focusedInput = 1
				return m, nil
			}
			if msg.String() == "esc" {
				m.step = exportStepProjects
				return m, nil
			}
			var cmd tea.Cmd
			m.customFrom, cmd = m.customFrom.Update(msg)
			return m, cmd
		}
		if msg.String() == "shift+tab" {
			m.customTo.Blur()
			m.customFrom.Focus()
			m.focusedInput = 0
			return m, nil
		}
		if msg.String() == "enter" {
			return m.startGeneration()
		}
		if msg.String() == "esc" {
			m.step = exportStepProjects
			return m, nil
		}
		var cmd tea.Cmd
		m.customTo, cmd = m.customTo.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "up", "k":
		if m.rangeCursor > 0 {
			m.rangeCursor--
		}
		m.rangeOption = timeRangeOption(m.rangeCursor)
	case "down", "j":
		if m.rangeCursor < 3 {
			m.rangeCursor++
		}
		m.rangeOption = timeRangeOption(m.rangeCursor)
	case "enter":
		if m.rangeOption == rangeCustom {
			m.customFrom.Focus()
			m.focusedInput = 0
			return m, textinput.Blink
		}
		return m.startGeneration()
	case "esc":
		m.step = exportStepProjects
	case "q", "ctrl+c":
		return m, tea.Quit
	}

	return m, nil
}

func (m ExportReportModel) startGeneration() (ExportReportModel, tea.Cmd) {
	m.step = exportStepGenerating
	m.err = nil
	return m, tea.Batch(m.spinner.Tick, m.generateCmd())
}

func (m ExportReportModel) updateDoneStep(msg tea.KeyMsg) (ExportReportModel, tea.Cmd) {
	switch msg.String() {
	case "enter", "esc":
		return m, func() tea.Msg {
			return NavigateMsg{To: screenDashboard}
		}
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m ExportReportModel) View() string {
	header := titleStyle.Render("SprintOS — ") + valueStyle.Render("Export Report")

	switch m.step {
	case exportStepProjects:
		return header + "\n\n" + m.viewProjectStep()
	case exportStepRange:
		return header + "\n\n" + m.viewRangeStep()
	case exportStepGenerating:
		return header + "\n\n" + m.viewGeneratingStep()
	case exportStepDone:
		return header + "\n\n" + m.viewDoneStep()
	}
	return header
}

func (m ExportReportModel) viewProjectStep() string {
	s := sectionHeader("Step 1 of 2 — Select Projects") + "\n\n"

	allCheck := "[ ]"
	if m.selectedAll {
		allCheck = successStyle.Render("[✓]")
	}
	line := fmt.Sprintf("  %s  All projects", allCheck)
	if m.projectCursor == 0 {
		line = highlightStyle.Render(fmt.Sprintf("  ▶ %s  All projects", allCheck))
	}
	s += line + "\n"

	if m.err != nil {
		s += "\n" + errorStyle.Render("  Error loading projects: "+m.err.Error()) + "\n"
	}

	for i, p := range m.projects {
		check := "[ ]"
		if m.checkedProjects[p.ID] {
			check = successStyle.Render("[✓]")
		}
		row := fmt.Sprintf("  %s  %s", check, p.Name)
		if m.projectCursor == i+1 {
			row = highlightStyle.Render(fmt.Sprintf("  ▶ %s  %s", check, p.Name))
		}
		s += row + "\n"
	}

	s += "\n"
	s += dimStyle.Render(strings.Repeat("─", 50)) + "\n"
	s += renderHintBar("↑/↓", "move", "space", "select", "enter", "next", "esc", "back") + "\n"
	return s
}

func (m ExportReportModel) viewRangeStep() string {
	s := sectionHeader("Step 2 of 2 — Select Time Range") + "\n\n"

	options := []string{"Last 7 days", "Last 30 days", "Last 90 days", "Custom range"}
	for i, opt := range options {
		check := "○"
		if m.rangeCursor == i {
			check = selectedStyle.Render("●")
		}
		row := fmt.Sprintf("  %s  %s", check, opt)
		if m.rangeCursor == i {
			row = highlightStyle.Render(row)
		}
		s += row + "\n"
	}

	if m.rangeOption == rangeCustom {
		s += "\n"
		fromLabel := dimStyle.Render("  From: ")
		toLabel := dimStyle.Render("  To:   ")
		if m.focusedInput == 0 {
			fromLabel = selectedStyle.Render("  From: ")
		} else {
			toLabel = selectedStyle.Render("  To:   ")
		}
		s += fromLabel + m.customFrom.View() + "\n"
		s += toLabel + m.customTo.View() + "\n"
		s += "\n" + dimStyle.Render("  tab to switch fields, enter to confirm") + "\n"
	}

	s += "\n"
	s += dimStyle.Render(strings.Repeat("─", 50)) + "\n"
	s += renderHintBar("↑/↓", "choose", "enter", "generate", "esc", "back") + "\n"
	return s
}

func (m ExportReportModel) viewGeneratingStep() string {
	return "\n  " + m.spinner.View() + selectedStyle.Render("  Generating PDF report…") + "\n\n" +
		dimStyle.Render("  Querying data and building the document. This may take a moment.") + "\n"
}

func (m ExportReportModel) viewDoneStep() string {
	if m.err != nil {
		return "\n" + errorStyle.Render("  ✗  Export failed") + "\n\n" +
			normalStyle.Render("  "+m.err.Error()) + "\n\n" +
			renderHintBar("enter", "back") + "\n"
	}
	return "\n" + successStyle.Render("  ✓  Report saved successfully") + "\n\n" +
		labelStyle.Render("  Path  ") + valueStyle.Render(m.savedPath) + "\n\n" +
		dimStyle.Render("  Open the file from your Desktop.") + "\n\n" +
		renderHintBar("enter", "back to board") + "\n"
}

func (m ExportReportModel) hasAnyChecked() bool {
	for _, v := range m.checkedProjects {
		if v {
			return true
		}
	}
	return false
}
