package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/domain"
)

type CEODashboardModel struct {
	metrics       *app.DashboardMetrics
	loading       bool
	err           error
	orgID         uint
	filterProject *domain.Project
	filterMode    bool
	filterCursor  int
	lastRefresh   time.Time
	windowWidth   int
	dashboardSvc  *app.DashboardService
	projectSvc    *app.ProjectService
}

type DashboardMetricsLoadedMsg struct {
	Metrics *app.DashboardMetrics
	Err     error
}

type DashboardAutoRefreshMsg time.Time

func dashboardAutoRefreshCmd() tea.Cmd {
	return tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
		return DashboardAutoRefreshMsg(t)
	})
}

func NewCEODashboardModel(orgID uint, dashboardSvc *app.DashboardService, projectSvc *app.ProjectService) CEODashboardModel {
	return CEODashboardModel{
		orgID:        orgID,
		loading:      true,
		windowWidth:  100,
		dashboardSvc: dashboardSvc,
		projectSvc:   projectSvc,
	}
}

func (m CEODashboardModel) loadCmd() tea.Cmd {
	return func() tea.Msg {
		var projectID *uint
		if m.filterProject != nil {
			id := m.filterProject.ID
			projectID = &id
		}
		metrics, err := m.dashboardSvc.GetMetrics(m.orgID, projectID)
		if err != nil {
			return DashboardMetricsLoadedMsg{Err: err}
		}
		return DashboardMetricsLoadedMsg{Metrics: metrics}
	}
}

func (m CEODashboardModel) Init() tea.Cmd {
	return tea.Batch(m.loadCmd(), dashboardAutoRefreshCmd())
}

func (m CEODashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(DashboardMetricsLoadedMsg); ok {
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.metrics = msg.Metrics
		m.lastRefresh = time.Now()
		return m, nil
	}

	if _, ok := msg.(DashboardAutoRefreshMsg); ok {
		m.loading = true
		return m, tea.Batch(m.loadCmd(), dashboardAutoRefreshCmd())
	}

	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.windowWidth = msg.Width
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.filterMode {
			switch msg.String() {
			case "up", "k":
				if m.filterCursor > 0 {
					m.filterCursor--
				}
			case "down", "j":
				if m.metrics != nil && m.filterCursor < len(m.metrics.Projects) {
					m.filterCursor++
				}
			case "enter":
				if m.metrics != nil {
					if m.filterCursor == 0 {
						m.filterProject = nil
					} else {
						p := m.metrics.Projects[m.filterCursor-1]
						m.filterProject = &p
					}
				}
				m.filterMode = false
				m.loading = true
				return m, m.loadCmd()
			case "esc":
				m.filterMode = false
			}
			return m, nil
		}

		switch msg.String() {
		case "f":
			m.filterMode = true
			m.filterCursor = 0
		case "r":
			m.loading = true
			return m, m.loadCmd()
		case "p":
			return m, func() tea.Msg {
				return NavigateMsg{To: screenDashboard}
			}
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m CEODashboardModel) View() string {
	if m.loading {
		return titleStyle.Render("SprintOS") + "\n\n" + normalStyle.Render("Loading metrics...") + "\n"
	}

	if m.err != nil {
		return titleStyle.Render("SprintOS") + "\n\n" + errorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())) + "\n"
	}

	if m.metrics == nil || !m.metrics.HasAnyData {
		return m.viewEmpty()
	}

	if m.filterMode {
		return m.viewFilterPicker()
	}

	return m.viewDashboard()
}

func (m CEODashboardModel) viewEmpty() string {
	art := dimStyle.Render("  ┌──────┐  ┌──────┐  ┌──────┐") + "\n" +
		dimStyle.Render("  │ TODO │  │  WIP │  │ DONE │") + "\n" +
		dimStyle.Render("  │      │  │      │  │      │") + "\n" +
		dimStyle.Render("  │  ·   │  │  ·   │  ") + successStyle.Render("│  ✓   │") + "\n" +
		dimStyle.Render("  │  ·   │  │      │  ") + successStyle.Render("│  ✓   │") + "\n" +
		dimStyle.Render("  └──────┘  └──────┘  └──────┘")

	s := "\n" + art + "\n\n"
	s += titleStyle.Render("  Welcome to SprintOS") + "\n\n"
	s += normalStyle.Render("  Your engineering command center. Track sprints, ship faster, stay aligned.") + "\n\n"
	s += selectedStyle.Render("  Getting started:") + "\n"
	s += hintKeyStyle.Render("  1.") + normalStyle.Render(" Create your first project") + "\n"
	s += hintKeyStyle.Render("  2.") + normalStyle.Render(" Set up your board states") + "\n"
	s += hintKeyStyle.Render("  3.") + normalStyle.Render(" Add tasks and assign your team") + "\n\n"
	s += dimStyle.Render(strings.Repeat("─", 54)) + "\n"
	s += renderHintBar("p", "go to projects", "q", "quit") + "\n"
	return s
}

func (m CEODashboardModel) viewFilterPicker() string {
	s := titleStyle.Render("SprintOS — Executive Dashboard") + "\n\n"
	s += sectionHeader("Filter by project") + "\n\n"

	if m.filterCursor == 0 {
		s += highlightStyle.Render("  All Projects") + "\n"
	} else {
		s += normalStyle.Render("  All Projects") + "\n"
	}

	if m.metrics != nil {
		for i, p := range m.metrics.Projects {
			if m.filterCursor == i+1 {
				s += highlightStyle.Render(fmt.Sprintf("  %s", p.Name)) + "\n"
			} else {
				s += normalStyle.Render(fmt.Sprintf("  %s", p.Name)) + "\n"
			}
		}
	}

	s += "\n" + renderHintBar("↑/↓", "move", "enter", "select", "esc", "cancel") + "\n"
	return s
}

func (m CEODashboardModel) viewDashboard() string {
	mt := m.metrics
	w := m.windowWidth
	if w < 80 {
		w = 80
	}

	filterLabel := "All Projects"
	if m.filterProject != nil {
		filterLabel = m.filterProject.Name
	}
	refreshLabel := ""
	if !m.lastRefresh.IsZero() {
		refreshLabel = "  " + dimStyle.Render("refreshed "+m.lastRefresh.Format("15:04:05"))
	}

	header := titleStyle.Render("SprintOS — Executive Dashboard") +
		"   " + labelStyle.Render(filterLabel) + refreshLabel + "\n"

	kpiRow := m.renderKPIRow(mt, w)
	midRow := m.renderMidRow(mt, w)
	bottomRow := m.renderBottomRow(mt, w)

	hint := "\n" + dimStyle.Render(strings.Repeat("─", w-2)) + "\n"
	hint += renderHintBar("f", "filter", "r", "refresh", "p", "projects", "q", "quit") + "\n"

	return header + "\n" + kpiRow + "\n" + midRow + "\n" + bottomRow + hint
}

func (m CEODashboardModel) renderKPIRow(mt *app.DashboardMetrics, w int) string {
	cardW := (w - 9) / 5
	if cardW < 10 {
		cardW = 10
	}

	sprintVal := fmt.Sprintf("%.0f%%", mt.SprintCompletionRate)
	sprintSub := ""
	if !mt.HasSprints {
		sprintVal = "N/A"
		sprintSub = "no sprints yet"
	} else if mt.ActiveSprintName != "" {
		sprintSub = "● " + truncate(mt.ActiveSprintName, cardW-2)
	}

	cycleVal := fmt.Sprintf("%.1f days", mt.AvgCycleTimeDays)
	if mt.AvgCycleTimeDays == 0 {
		cycleVal = "N/A"
	}

	onTimeVal := fmt.Sprintf("%.0f%%", mt.OnTimeDeliveryRate)

	overdueVal := fmt.Sprintf("%d", mt.OverdueCount)
	overdueStyle := selectedStyle
	if mt.OverdueCount > 0 {
		overdueStyle = errorStyle
		overdueVal = fmt.Sprintf("%d ⚠", mt.OverdueCount)
	}

	cards := []string{
		m.kpiCard("Sprint complete", sprintVal, sprintSub, selectedStyle, mt.HasSprints, cardW),
		m.kpiCard("Cycle time", cycleVal, "", selectedStyle, true, cardW),
		m.kpiCard("Throughput", fmt.Sprintf("%d/week", mt.WeeklyThroughput), "", successStyle, true, cardW),
		m.kpiCard("On-time", onTimeVal, "", selectedStyle, true, cardW),
		m.kpiCard("Overdue", overdueVal, "", overdueStyle, true, cardW),
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, cards...)
}

func (m CEODashboardModel) kpiCard(label, value, sub string, valStyle lipgloss.Style, active bool, width int) string {
	style := cardStyle.Width(width)
	if active {
		style = activeCardStyle.Width(width)
	}

	body := labelStyle.Render(strings.ToUpper(truncate(label, width))) + "\n\n"
	body += valStyle.Render(truncate(value, width))
	if sub != "" {
		body += "\n" + dimStyle.Render(truncate(sub, width))
	}

	return style.Render(body)
}

func (m CEODashboardModel) renderMidRow(mt *app.DashboardMetrics, w int) string {
	colW := (w - 7) / 2
	if colW < 20 {
		colW = 20
	}
	barW := colW - 18
	if barW < 4 {
		barW = 4
	}

	leftContent := sectionHeader("Velocity trend (last 4 sprints)") + "\n"
	if !mt.HasSprints {
		leftContent += "\n" + normalStyle.Render("  No sprints yet.") + "\n"
		leftContent += dimStyle.Render("  Create a sprint to track velocity.") + "\n"
	} else if len(mt.VelocityTrend) == 0 {
		leftContent += "\n" + normalStyle.Render("  No sprint data.") + "\n"
	} else {
		maxC := 1
		for _, v := range mt.VelocityTrend {
			if v.Completed > maxC {
				maxC = v.Completed
			}
		}
		for _, v := range mt.VelocityTrend {
			bar := renderBar(v.Completed, maxC, barW, selectedStyle)
			leftContent += fmt.Sprintf("\n  %-10s %s %s",
				truncate(v.SprintName, 10),
				bar,
				dimStyle.Render(fmt.Sprintf(" %d", v.Completed)),
			)
		}
		if len(mt.VelocityTrend) >= 2 {
			first := mt.VelocityTrend[0].Completed
			last := mt.VelocityTrend[len(mt.VelocityTrend)-1].Completed
			if first > 0 {
				pct := float64(last-first) / float64(first) * 100
				if pct >= 0 {
					leftContent += "\n\n  " + successStyle.Render(fmt.Sprintf("↑ +%.0f%%", pct)) + dimStyle.Render(" vs first sprint")
				} else {
					leftContent += "\n\n  " + errorStyle.Render(fmt.Sprintf("↓ %.0f%%", pct)) + dimStyle.Render(" vs first sprint")
				}
			}
		}
	}

	rightContent := sectionHeader("State distribution") + "\n"
	if len(mt.StateDistribution) == 0 {
		rightContent += "\n" + normalStyle.Render("  No tasks yet.") + "\n"
	} else {
		maxC := 1
		for _, sc := range mt.StateDistribution {
			if sc.Count > maxC {
				maxC = sc.Count
			}
		}
		for _, sc := range mt.StateDistribution {
			barStyle := selectedStyle
			if sc.Count == maxC {
				barStyle = successStyle
			}
			bar := renderBar(sc.Count, maxC, barW, barStyle)
			rightContent += fmt.Sprintf("\n  %-12s %s %s",
				truncate(sc.StateName, 12),
				bar,
				dimStyle.Render(fmt.Sprintf(" %d", sc.Count)),
			)
		}
	}

	left := cardStyle.Width(colW).Render(leftContent)
	right := cardStyle.Width(colW).Render(rightContent)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, " ", right)
}

func (m CEODashboardModel) renderBottomRow(mt *app.DashboardMetrics, w int) string {
	colW := (w - 7) / 2
	if colW < 20 {
		colW = 20
	}
	barW := colW - 22
	if barW < 4 {
		barW = 4
	}

	leftContent := sectionHeader("Team workload") + "\n"
	if len(mt.TeamWorkload) == 0 {
		leftContent += "\n" + normalStyle.Render("  No assignments yet.") + "\n"
	} else {
		maxA := 1
		for _, mw := range mt.TeamWorkload {
			if mw.AssignedCount > maxA {
				maxA = mw.AssignedCount
			}
		}
		for _, mw := range mt.TeamWorkload {
			bar := renderBar(mw.AssignedCount, maxA, barW, selectedStyle)
			leftContent += fmt.Sprintf("\n  %-10s %s %s",
				truncate(mw.UserName, 10),
				bar,
				dimStyle.Render(fmt.Sprintf(" %d (%d done)", mw.AssignedCount, mw.CompletedCount)),
			)
		}
	}

	rightContent := sectionHeader("Recently completed") + "\n"
	if len(mt.RecentlyCompleted) == 0 {
		rightContent += "\n" + normalStyle.Render("  No completed tasks yet.") + "\n"
	} else {
		for _, t := range mt.RecentlyCompleted {
			assignee := "—"
			if t.Assignee != nil {
				assignee = truncate(t.Assignee.Name, 8)
			}
			age := formatAge(*t.CompletedAt)
			rightContent += "\n  " +
				successStyle.Render("✓") + " " +
				normalStyle.Render(truncate(t.Title, colW-22)) +
				"  " + dimStyle.Render(assignee) +
				"  " + dimStyle.Render(age)
		}
	}

	left := cardStyle.Width(colW).Render(leftContent)
	right := cardStyle.Width(colW).Render(rightContent)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, " ", right)
}

func formatAge(t time.Time) string {
	d := time.Since(t)
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	return fmt.Sprintf("%dd ago", int(d.Hours()/24))
}
