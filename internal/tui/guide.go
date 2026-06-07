package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const guideHeaderH = 5
const guideFooterH = 2

type guideSection struct {
	icon  string
	short string
	title string
	lines []string
}

type GuideModel struct {
	sections      []guideSection
	activeSection int
	scrollY       int
	windowHeight  int
	windowWidth   int
}

func NewGuideModel() GuideModel {
	m := GuideModel{
		windowHeight: 30,
		windowWidth:  100,
	}
	m.sections = buildGuideSections()
	return m
}

func (m GuideModel) Init() tea.Cmd { return nil }

func (m GuideModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowHeight = msg.Height
		m.windowWidth = msg.Width
		return m, nil
	case tea.KeyMsg:
		vis := m.visible()
		cur := &m.sections[m.activeSection]
		maxScroll := len(cur.lines) - vis
		if maxScroll < 0 {
			maxScroll = 0
		}
		switch msg.String() {
		case "q", "esc":
			return m, func() tea.Msg { return GoBackMsg{} }
		case "down", "j":
			if m.scrollY < maxScroll {
				m.scrollY++
			}
		case "up", "k":
			if m.scrollY > 0 {
				m.scrollY--
			}
		case "pgdown", " ":
			m.scrollY += vis
			if m.scrollY > maxScroll {
				m.scrollY = maxScroll
			}
		case "pgup":
			m.scrollY -= vis
			if m.scrollY < 0 {
				m.scrollY = 0
			}
		case "g":
			m.scrollY = 0
		case "G":
			m.scrollY = maxScroll
		case "tab", "right", "l":
			if m.activeSection < len(m.sections)-1 {
				m.activeSection++
				m.scrollY = 0
			}
		case "shift+tab", "left", "h":
			if m.activeSection > 0 {
				m.activeSection--
				m.scrollY = 0
			}
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			idx := int(msg.String()[0] - '1')
			if idx < len(m.sections) {
				m.activeSection = idx
				m.scrollY = 0
			}
		}
	}
	return m, nil
}

func (m GuideModel) visible() int {
	v := m.windowHeight - guideHeaderH - guideFooterH
	if v < 4 {
		v = 4
	}
	return v
}

func (m GuideModel) View() string {
	sec := m.sections[m.activeSection]
	vis := m.visible()

	title := titleStyle.Render("SprintOS") + dimStyle.Render("  ·  ") + valueStyle.Render("User Guide")

	var tabParts []string
	for i, s := range m.sections {
		label := fmt.Sprintf(" %d %s %s ", i+1, s.icon, s.short)
		if i == m.activeSection {
			tabParts = append(tabParts, selectedStyle.Render(label))
		} else {
			tabParts = append(tabParts, dimStyle.Render(label))
		}
	}
	tabs := strings.Join(tabParts, dimStyle.Render("·"))

	w := m.windowWidth - 2
	if w < 40 {
		w = 40
	}
	sep := dimStyle.Render(strings.Repeat("─", w))

	lines := sec.lines
	start := m.scrollY
	end := start + vis
	if end > len(lines) {
		end = len(lines)
	}
	if start > len(lines) {
		start = len(lines)
	}

	var content []string
	if start > 0 {
		content = append(content, dimStyle.Render(fmt.Sprintf("  ↑ %d lines above", start)))
	} else {
		content = append(content, "")
	}
	content = append(content, lines[start:end]...)
	remaining := len(lines) - end
	if remaining > 0 {
		for len(content) < vis {
			content = append(content, "")
		}
		content = append(content, dimStyle.Render(fmt.Sprintf("  ↓ %d more  ·  j/↓ to scroll", remaining)))
	}

	body := strings.Join(content, "\n")
	footer := sep + "\n" + renderHintBar(
		"↑/↓", "scroll",
		"tab/→", "next",
		"shift+tab/←", "prev",
		"1-9", "jump",
		"g/G", "top/bottom",
		"esc", "back",
	)

	return title + "\n" + tabs + "\n" + sep + "\n" + body + "\n" + footer
}

func buildGuideSections() []guideSection {
	return []guideSection{
		sectionGettingStarted(),
		sectionDashboard(),
		sectionKanban(),
		sectionTaskDetail(),
		sectionSprints(),
		sectionTimerPomodoro(),
		sectionMenuBar(),
		sectionExportReport(),
		sectionMCPAI(),
		sectionCLI(),
	}
}

func gl(parts ...string) string {
	return strings.Join(parts, "")
}

func guideKey(k string) string  { return hintKeyStyle.Render(k) }
func guideVal(s string) string  { return valueStyle.Render(s) }
func guideHead(s string) string { return selectedStyle.Render(s) }
func guideDim(s string) string  { return dimStyle.Render(s) }
func guideNorm(s string) string { return normalStyle.Render(s) }
func guideTip(s string) string  { return dimStyle.Render("  💡 " + s) }
func guideWarn(s string) string { return warningStyle.Render("  ⚠  " + s) }
func guideOk(s string) string   { return successStyle.Render("  ✓  " + s) }
func guideSep() string          { return dimStyle.Render("  " + strings.Repeat("─", 60)) }

func kv(key, desc string) string {
	return "  " + hintKeyStyle.Render(fmt.Sprintf("%-14s", key)) + normalStyle.Render(desc)
}

func sectionGettingStarted() guideSection {
	lines := []string{
		"",
		guideHead("  🚀  Getting Started"),
		"",
		guideNorm("  SprintOS is a keyboard-driven project manager that runs entirely"),
		guideNorm("  in your terminal. One binary. No configuration required."),
		"",
		guideSep(),
		guideHead("  Installation"),
		"",
		guideNorm("  macOS (Homebrew)"),
		guideVal("    brew install varmiguemunoz/sprintos/sprintos"),
		"",
		guideNorm("  Windows (Scoop)"),
		guideVal("    scoop bucket add sprintos https://github.com/varmiguemunoz/scoop-sprintos"),
		guideVal("    scoop install sprintos"),
		"",
		guideNorm("  Linux"),
		guideVal("    curl -fsSL https://raw.githubusercontent.com/varmiguemunoz/sprintos/main/install.sh | sh"),
		"",
		guideSep(),
		guideHead("  First Launch"),
		"",
		guideNorm("  1. Set your database URL in a .env file:"),
		guideVal("       DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=require"),
		"",
		guideNorm("  2. Launch the app:"),
		guideVal("       sprintos start"),
		"",
		guideNorm("  3. The first-launch wizard walks you through 3 steps:"),
		guideOk("Login with GitHub OAuth"),
		guideOk("Create your organization (name, prefix, WhatsApp number)"),
		guideOk("Set up your first project board"),
		"",
		guideSep(),
		guideHead("  After Updates"),
		"",
		guideNorm("  Run with --migrate once to apply schema changes:"),
		guideVal("    sprintos start --migrate"),
		"",
		guideTip("Normal launches skip migrations entirely for fast startup."),
		"",
		guideSep(),
		guideHead("  Navigation Basics"),
		"",
		guideNorm("  All screens are keyboard-driven. The hint bar at the"),
		guideNorm("  bottom of each screen shows the available shortcuts."),
		"",
		kv("?", "Open this guide from any main screen"),
		kv("esc", "Go back to the previous screen"),
		kv("q", "Quit the application"),
		kv("ctrl+c", "Force quit"),
		"",
		guideTip("Press ? from the Kanban board, Projects, or Org Settings."),
		"",
	}
	return guideSection{icon: "🚀", short: "Start", title: "Getting Started", lines: lines}
}

func sectionDashboard() guideSection {
	lines := []string{
		"",
		guideHead("  📊  Executive Dashboard"),
		"",
		guideNorm("  The first screen after login. A live snapshot of your entire"),
		guideNorm("  organization's health across all projects."),
		"",
		guideSep(),
		guideHead("  KPI Metrics Shown"),
		"",
		kv("Sprint complete", "% of current sprint tasks finished"),
		kv("Cycle time", "Avg days from task creation to done"),
		kv("Throughput", "Tasks completed in the last 7 days"),
		kv("On-time rate", "% tasks completed before due date"),
		kv("Overdue", "Tasks past their due date"),
		"",
		guideSep(),
		guideHead("  Charts & Tables"),
		"",
		kv("Velocity trend", "Bar chart of last 4 sprints (done / total)"),
		kv("State distribution", "How tasks are spread across board columns"),
		kv("Team workload", "Assigned vs completed per member"),
		kv("Recently completed", "Last 8 finished tasks with assignee + age"),
		"",
		guideSep(),
		guideHead("  Keyboard Shortcuts"),
		"",
		kv("f", "Filter metrics by a specific project"),
		kv("r", "Manual refresh now"),
		kv("p", "Go to the Projects list"),
		kv("?", "Open this guide"),
		kv("q", "Quit"),
		"",
		guideTip("The dashboard auto-refreshes every 10 minutes in the background."),
		guideTip("Use r to force an immediate refresh at any time."),
		"",
		guideSep(),
		guideHead("  Projects List"),
		"",
		guideNorm("  Navigate your organization's projects (shown after pressing p)."),
		"",
		kv("↑/↓  j/k", "Navigate projects"),
		kv("enter", "Open project's Kanban board"),
		kv("d", "Go to Executive Dashboard"),
		kv("n", "Create a new project"),
		kv("e", "Edit selected project (name, description)"),
		kv("D", "Delete project — confirms with y / n"),
		kv("/", "Fuzzy search across all tasks"),
		kv("s", "Organization settings"),
		kv("?", "Open this guide"),
		kv("q", "Quit"),
		"",
	}
	return guideSection{icon: "📊", short: "Dashboard", title: "Executive Dashboard & Projects", lines: lines}
}

func sectionKanban() guideSection {
	lines := []string{
		"",
		guideHead("  📌  Kanban Board"),
		"",
		guideNorm("  The main workspace. Tasks are organized in columns by state."),
		guideNorm("  The default workflow is: Backlog → In Progress → In Review → Done."),
		"",
		guideSep(),
		guideHead("  Reading the Board"),
		"",
		kv("!!", "Critical priority task"),
		kv("↑", "High priority task"),
		kv("✗  (red)", "Task is overdue"),
		kv("⚠  (yellow)", "Due within 48 hours"),
		kv("MA", "Assignee initials (top-right of task card)"),
		"",
		guideSep(),
		guideHead("  Keyboard Shortcuts"),
		"",
		kv("← →  h l", "Move between columns"),
		kv("↑ ↓  j k", "Move between tasks in the active column"),
		kv("enter", "Open task detail"),
		kv("n  or  +", "Create a new task in the current column"),
		kv("m", "Move selected task to another state (dialog)"),
		kv("d", "Delete selected task — confirms with y / n"),
		kv("R", "Enter column reorder mode"),
		kv("v", "Switch to Sprint View"),
		kv("b", "Edit board layout (add/rename/delete states)"),
		kv("E", "Export PDF executive report"),
		kv("/", "Fuzzy search all tasks"),
		kv("?", "Open this guide"),
		kv("esc", "Back to Projects list"),
		kv("q", "Quit"),
		"",
		guideSep(),
		guideHead("  🔀  Column Reorder Mode  (press R)"),
		"",
		guideNorm("  The selected column turns amber. Move it freely:"),
		"",
		kv("← /  h", "Move this column to the left"),
		kv("→ /  l", "Move this column to the right"),
		kv("enter", "Save the new order to the database"),
		kv("esc", "Cancel — discard changes, no DB write"),
		"",
		guideTip("Positions are persisted immediately after enter."),
		guideTip("All team members see the new order on their next board load."),
		"",
		guideSep(),
		guideHead("  Board Setup  (press b)"),
		"",
		guideNorm("  Add, rename, reorder, or delete board columns."),
		guideNorm("  Choose from preset templates:"),
		"",
		kv("Standard", "Backlog → In Progress → In Review → Done"),
		kv("Simple", "Todo → Done"),
		"",
		guideNorm("  Or build a fully custom workflow for your team."),
		"",
	}
	return guideSection{icon: "📌", short: "Kanban", title: "Kanban Board", lines: lines}
}

func sectionTaskDetail() guideSection {
	lines := []string{
		"",
		guideHead("  🗒️  Task Detail"),
		"",
		guideNorm("  Full view of a task: metadata, time tracking, subtasks,"),
		guideNorm("  and comments. Open with enter from the Kanban board."),
		"",
		guideSep(),
		guideHead("  Keyboard Shortcuts"),
		"",
		kv("a", "Assign / reassign / unassign a team member"),
		kv("c", "Add a comment"),
		kv("e", "Edit task (title, description)"),
		kv("m", "Move task to a different state (inline picker)"),
		kv("s", "Create a subtask"),
		kv("T", "Start / stop time tracker for this task"),
		kv("l", "Log time manually (minutes + optional note)"),
		kv("↑ ↓  j k", "Navigate subtasks"),
		kv("enter", "Open selected subtask detail"),
		kv("d", "Delete the selected subtask"),
		kv("esc", "Back to Kanban board"),
		kv("q", "Quit"),
		"",
		guideSep(),
		guideHead("  What's Displayed"),
		"",
		kv("Details box", "State · Assignee · Priority · Due date · Description"),
		kv("Time tracked", "Total minutes logged + live running timer"),
		kv("Subtasks", "List with [✓] checkmarks and completion ratio"),
		kv("Comments", "Thread with author name and timestamp"),
		"",
		guideTip("When a timer is running, the Time box shows a live counter:  ● 01:23:45"),
		"",
		guideSep(),
		guideHead("  ⏱️  Time Tracking"),
		"",
		guideNorm("  Press T to toggle the timer. Only one timer runs at a time."),
		guideNorm("  Time is auto-saved when you stop — rounded up to 1 minute minimum."),
		"",
		guideNorm("  Press l to open the manual log form:"),
		"",
		kv("Minutes", "Time to log, e.g. 90 for 1h 30m"),
		kv("Note", "Optional description of what was done"),
		"",
		kv("tab", "Next field"),
		kv("ctrl+s", "Save"),
		kv("esc", "Cancel"),
		"",
		guideSep(),
		guideHead("  🧩  Subtask Detail"),
		"",
		guideNorm("  Each subtask has its own time tracking and comment thread."),
		"",
		kv("e", "Edit subtask (title, description)"),
		kv("T", "Start / stop time tracker"),
		kv("l", "Log time manually"),
		kv("c", "Add a comment to the subtask"),
		kv("esc", "Back to parent task"),
		"",
		guideSep(),
		guideHead("  Create / Edit Task Form"),
		"",
		kv("Title  *", "Required — short descriptive name"),
		kv("Description", "Optional — multi-line, press enter for new lines"),
		kv("Due Date", "Optional — format YYYY-MM-DD  (e.g. 2026-12-31)"),
		"",
		kv("tab", "Next field"),
		kv("shift+tab", "Previous field"),
		kv("ctrl+s", "Save from any field"),
		kv("enter", "Advance or submit on last textinput field"),
		kv("esc", "Cancel"),
		"",
	}
	return guideSection{icon: "🗒️", short: "Tasks", title: "Task Detail & Time Tracking", lines: lines}
}

func sectionSprints() guideSection {
	lines := []string{
		"",
		guideHead("  🏃  Sprint View"),
		"",
		guideNorm("  Plan, monitor, and manage all sprints for a project."),
		guideNorm("  Access it from the Kanban board by pressing v."),
		"",
		guideSep(),
		guideHead("  Sprint List"),
		"",
		guideNorm("  Each sprint shows: name · date range · task count"),
		guideNorm("  · days remaining · status badge."),
		"",
		guideNorm("  Status badges:"),
		kv("● active", "Currently in progress"),
		kv("upcoming", "Starts in the future"),
		kv("completed", "Finished"),
		"",
		guideNorm("  Selecting a sprint expands it to show all its tasks."),
		"",
		guideSep(),
		guideHead("  Keyboard Shortcuts"),
		"",
		kv("↑ ↓  j k", "Navigate sprints"),
		kv("c", "Create a new sprint"),
		kv("e", "Edit selected sprint (name, goal, dates)"),
		kv("D", "Delete selected sprint — confirms with y / n"),
		kv("p", "Enter planning mode"),
		kv("esc", "Back to Kanban board"),
		kv("q", "Quit"),
		"",
		guideSep(),
		guideHead("  📋  Planning Mode  (press p)"),
		"",
		guideNorm("  Two columns: Backlog (left) and Sprint Tasks (right)."),
		guideNorm("  Move tasks from backlog into the selected sprint."),
		"",
		kv("↑ ↓  j k", "Navigate backlog tasks"),
		kv("enter", "Add selected task to the sprint"),
		kv("esc", "Exit planning mode"),
		"",
		guideSep(),
		guideHead("  Create / Edit Sprint Form"),
		"",
		kv("Sprint name  *", "Required — e.g. Sprint 3 / Q3 Auth Sprint"),
		kv("Goal", "Optional — one-line sprint objective"),
		kv("Start date  *", "Required — format YYYY-MM-DD"),
		kv("End date  *", "Required — must be after start date"),
		"",
		kv("tab / ↓", "Next field"),
		kv("shift+tab / ↑", "Previous field"),
		kv("enter", "Advance or save on last field"),
		kv("esc", "Cancel"),
		"",
		guideTip("Complete a sprint via CLI: sprintos sprint complete --id 1 --backlog-state 1"),
		guideTip("Unfinished tasks are automatically moved back to Backlog on completion."),
		"",
		guideSep(),
		guideHead("  Sprint CLI Commands"),
		"",
		guideVal("  sprintos sprint create --name 'Sprint 1' --project 1 --start 2026-06-01 --end 2026-06-14"),
		guideVal("  sprintos sprint list --project 1"),
		guideVal("  sprintos sprint assign --sprint 1 --task 5"),
		guideVal("  sprintos sprint velocity --id 1"),
		guideVal("  sprintos sprint complete --id 1 --backlog-state 1"),
		guideVal("  sprintos sprint snapshot   # daily burndown — add to cron"),
		"",
	}
	return guideSection{icon: "🏃", short: "Sprints", title: "Sprint Management", lines: lines}
}

func sectionTimerPomodoro() guideSection {
	lines := []string{
		"",
		guideHead("  ⏱️  Timer & 🍅 Pomodoro"),
		"",
		guideSep(),
		guideHead("  Task Timer"),
		"",
		guideNorm("  Track exactly how much time you spend on each task."),
		guideNorm("  One timer runs at a time — attempting to start a second"),
		guideNorm("  one shows a notification asking you to stop the first."),
		"",
		guideHead("  Starting the timer"),
		"",
		kv("From Task Detail", "Press T to toggle on/off"),
		kv("From Menu Bar", "Select a task → click ▶ Start Timer"),
		"",
		guideHead("  While the timer runs"),
		"",
		guideNorm("  The Task Detail screen shows a live counter:"),
		guideVal("    ●  01:23:45"),
		"",
		guideNorm("  The macOS menu bar shows the elapsed time:"),
		guideVal("    ⏱  01:23"),
		"",
		guideHead("  Stopping the timer"),
		"",
		kv("From Task Detail", "Press T again"),
		kv("From Menu Bar", "Click ■ Stop Timer"),
		"",
		guideNorm("  Time is auto-saved. Rounded up to the nearest minute"),
		guideNorm("  (minimum 1 minute logged)."),
		"",
		guideSep(),
		guideHead("  🔔  Timer Notifications"),
		"",
		guideNorm("  A macOS notification fires when a timer starts — whether"),
		guideNorm("  triggered from the TUI or from the menu bar."),
		"",
		guideNorm("  Real-time sync: if you start a timer in the TUI, the"),
		guideNorm("  menu bar detects it within 1 second and updates automatically."),
		"",
		guideSep(),
		guideHead("  🍅  Pomodoro Timer  (macOS menu bar)"),
		"",
		guideNorm("  Focus sessions with auto-restart and macOS notifications."),
		"",
		kv("Start 15 min", "15-minute focus session"),
		kv("Start 30 min", "30-minute focus session"),
		kv("Start 45 min", "45-minute focus session"),
		kv("■ Stop", "Cancel the current session at any time"),
		"",
		guideHead("  How a Pomodoro session works:"),
		"",
		guideNorm("  1. Choose a session length."),
		guideNorm("     Menu bar shows countdown: 🍅 24:35"),
		"",
		guideNorm("  2. When time is up, a notification fires:"),
		guideVal(`     "Time's up! Take a break."`),
		"",
		guideNorm("  3. A 15-second grace period starts: ⚠ 00:12"),
		"",
		guideNorm("  4. If you don't press Stop, the session auto-restarts"),
		guideNorm("     with another notification."),
		"",
		guideTip("Pomodoro and task timer are independent — run both simultaneously."),
		"",
	}
	return guideSection{icon: "⏱️", short: "Timer", title: "Timer & Pomodoro", lines: lines}
}

func sectionMenuBar() guideSection {
	lines := []string{
		"",
		guideHead("  🍎  macOS Menu Bar"),
		"",
		guideNorm("  The ⚡ menu bar app starts automatically with sprintos start."),
		guideNorm("  Track time and monitor timers without opening a terminal."),
		"",
		guideSep(),
		guideHead("  Timer Section"),
		"",
		kv("Select Task ▶", "Browse all non-completed tasks across projects"),
		kv("", "Navigate pages with Previous / Next inside the submenu"),
		kv("▶ Start Timer", "Start tracking time for the selected task"),
		kv("■ Stop Timer", "Stop the running timer and save time"),
		kv("Status label", "Shows: Active: 01:23 — Task Title"),
		"",
		guideSep(),
		guideHead("  Rules & Behavior"),
		"",
		guideWarn("Only one timer can run at a time."),
		guideNorm("  Attempting to start a second timer shows a notification:"),
		guideVal(`  "A timer is already running. Stop it first."`),
		"",
		guideOk("Real-time sync with the TUI."),
		guideNorm("  If you start a timer in the terminal, the menu bar detects"),
		guideNorm("  it within 1 second, enables Stop, and shows the task name."),
		"",
		guideOk("Task selector only shows non-Done tasks."),
		guideNorm("  Completed tasks are filtered out automatically."),
		"",
		guideSep(),
		guideHead("  Pomodoro Section"),
		"",
		kv("Start 15 / 30 / 45", "Start a focus session of that length"),
		kv("■ Stop Pomodoro", "Cancel the current focus session"),
		"",
		guideNorm("  While a Pomodoro runs, the menu bar title shows the"),
		guideNorm("  countdown live: 🍅 24:35"),
		"",
		guideNorm("  After the grace period, it auto-restarts: 🔄 notification fires."),
		"",
		guideTip("The task timer and Pomodoro are completely independent."),
		"",
	}
	return guideSection{icon: "🍎", short: "Menu Bar", title: "macOS Menu Bar", lines: lines}
}

func sectionExportReport() guideSection {
	lines := []string{
		"",
		guideHead("  📤  Export PDF Report"),
		"",
		guideNorm("  Generate an executive-grade PDF with charts and KPIs."),
		guideNorm("  Saved automatically to ~/Desktop/sprintos-report-YYYY-MM-DD.pdf"),
		"",
		guideNorm("  Open the export wizard from the Kanban board:"),
		kv("E", "Open the 2-step export wizard"),
		"",
		guideSep(),
		guideHead("  Step 1 — Project Selection"),
		"",
		kv("↑ / ↓", "Navigate options"),
		kv("space", "Toggle a project on / off"),
		kv("enter", "Proceed to Step 2"),
		kv("esc", "Cancel"),
		"",
		guideNorm("  'All projects' is pre-selected. Use space to pick individual ones."),
		"",
		guideSep(),
		guideHead("  Step 2 — Time Range"),
		"",
		kv("Last 7 days", "Compact weekly snapshot"),
		kv("Last 30 days", "Monthly view"),
		kv("Last 90 days", "Quarterly review"),
		kv("Custom range", "Enter From and To dates manually"),
		"",
		kv("↑ / ↓", "Navigate options"),
		kv("enter", "Generate PDF (or confirm custom dates)"),
		kv("tab", "Switch between date fields (custom range only)"),
		kv("esc", "Go back to Step 1"),
		"",
		guideSep(),
		guideHead("  What the PDF Contains"),
		"",
		guideOk("Cover — org name · date range · generated timestamp"),
		guideOk("Executive Summary — 6 KPI numbers"),
		guideOk("Project Health table — per-project breakdown"),
		guideOk("Weekly Velocity chart — bar chart of completed tasks/week"),
		guideOk("Team Performance table — tasks done + hours per member"),
		guideOk("Priority Risk — critical/high open tasks + overdue list"),
		"",
		guideSep(),
		guideHead("  The 6 KPIs on the Summary Page"),
		"",
		kv("Tasks Created", "New tasks opened in the period"),
		kv("Completed", "Tasks finished in the period"),
		kv("On-Time Rate", "% finished before their due date"),
		kv("Hours Logged", "Total tracked time across all selected projects"),
		kv("Avg Cycle Time", "Mean days from creation to completion"),
		kv("Overdue", "Open tasks past their due date"),
		"",
	}
	return guideSection{icon: "📤", short: "Export", title: "PDF Export", lines: lines}
}

func sectionMCPAI() guideSection {
	lines := []string{
		"",
		guideHead("  🤖  AI & MCP Integration"),
		"",
		guideNorm("  SprintOS exposes an MCP server that AI agents can connect to"),
		guideNorm("  and control your entire board through natural language."),
		"",
		guideSep(),
		guideHead("  Setup (from inside the TUI)"),
		"",
		guideNorm("  1. Go to Organization Settings (s from Projects screen)"),
		guideNorm("  2. Press m to open MCP Setup"),
		guideNorm("  3. Select your AI tool from the list:"),
		"",
		kv("Claude Desktop", "~/Library/Application Support/Claude/claude_desktop_config.json"),
		kv("Cursor", "~/.cursor/mcp.json"),
		kv("Windsurf", "~/.codeium/windsurf/mcp_config.json"),
		kv("Zed", "~/.config/zed/settings.json"),
		"",
		guideNorm("  4. SprintOS writes the correct config automatically."),
		guideNorm("  5. Restart your AI tool to activate."),
		"",
		guideSep(),
		guideHead("  Setup (from the CLI)"),
		"",
		guideVal("  sprintos mcp"),
		"",
		guideSep(),
		guideHead("  Available MCP Tools"),
		"",
		kv("list_projects", "List all projects in the organization"),
		kv("list_states", "List board columns for a project"),
		kv("list_tasks", "List all tasks with IDs, states, assignees"),
		kv("get_task_detail", "Full task detail with comments and subtasks"),
		kv("create_task", "Create with title · state · priority · due date"),
		kv("update_task", "Edit title, description, or due date"),
		kv("move_task", "Move a task to a different state"),
		kv("assign_task", "Assign or unassign a team member"),
		kv("add_comment", "Add a comment to a task"),
		kv("delete_task", "Delete a task permanently"),
		kv("list_members", "List all organization members"),
		kv("list_overdue_tasks", "All tasks past their due date"),
		kv("analyze_stale_tasks", "Tasks stuck in a state with suggested actions"),
		kv("summarize_project", "Project health: counts, overdue, workload"),
		kv("generate_sprint", "Create multiple tasks from a structured list"),
		"",
		guideSep(),
		guideHead("  Example AI Prompts"),
		"",
		guideVal(`  "Generate a sprint for the TaoFlow project based on this PRD: ..."`),
		guideVal(`  "Which tasks have been in the same state for more than 5 days?"`),
		guideVal(`  "Create 5 tasks for the auth module and put them in Backlog"`),
		guideVal(`  "Move task #15 to In Review and assign it to Miguel"`),
		guideVal(`  "Show me the health summary for project 2"`),
		"",
		guideTip("Your AI agent sees task numbers, states, priorities, assignees, and dates."),
		"",
	}
	return guideSection{icon: "🤖", short: "MCP / AI", title: "AI & MCP Integration", lines: lines}
}

func sectionCLI() guideSection {
	lines := []string{
		"",
		guideHead("  ⌨️  CLI Commands"),
		"",
		guideNorm("  All commands also work without the TUI — useful for"),
		guideNorm("  scripts, CI pipelines, and automation."),
		"",
		guideSep(),
		guideHead("  Core"),
		"",
		guideVal("  sprintos start               # Launch TUI + macOS menu bar"),
		guideVal("  sprintos start --migrate      # Run DB migrations then launch"),
		guideVal("  sprintos --help               # List all commands"),
		guideVal("  sprintos --version            # Show current version"),
		"",
		guideSep(),
		guideHead("  Tasks"),
		"",
		guideVal(`  sprintos task create "Fix login bug"`),
		guideVal(`  sprintos task create "Deploy v2" --project 1 --state "In Progress" --priority high`),
		guideVal("  sprintos task ls"),
		guideVal("  sprintos task ls --state backlog --format json"),
		guideVal("  sprintos task move 5 done"),
		guideVal("  sprintos task show 5"),
		"",
		guideNorm("  Priority values: low · medium · high · critical"),
		"",
		guideSep(),
		guideHead("  Sprints"),
		"",
		guideVal("  sprintos sprint create --name 'Sprint 1' --project 1 --start 2026-06-01 --end 2026-06-14"),
		guideVal("  sprintos sprint list --project 1"),
		guideVal("  sprintos sprint assign --sprint 1 --task 5"),
		guideVal("  sprintos sprint velocity --id 1"),
		guideVal("  sprintos sprint complete --id 1 --backlog-state 1"),
		guideVal("  sprintos sprint snapshot       # burndown — add to cron"),
		"",
		guideSep(),
		guideHead("  Reports & Export"),
		"",
		guideVal("  sprintos report --project 1 --completed 30"),
		guideVal("  sprintos export --project 1 --format csv --output tasks.csv"),
		guideVal("  sprintos export --format json | jq '.[].Title'"),
		guideVal("  sprintos standup"),
		guideVal("  sprintos review --days 5 --notify"),
		guideVal("  sprintos my-tasks"),
		"",
		guideSep(),
		guideHead("  API Keys & Server"),
		"",
		guideVal("  sprintos api-key create --name 'zapier'"),
		guideVal("  sprintos api-key list"),
		guideVal("  sprintos api-key revoke --id 3"),
		guideVal("  sprintos serve --port 8090"),
		"",
		guideSep(),
		guideHead("  GitHub Integration"),
		"",
		guideVal("  sprintos github setup          # Connect a repo to a project"),
		guideVal("  sprintos github list           # List connected repos"),
		guideVal("  sprintos serve --port 8090     # Start webhook receiver"),
		"",
		guideNorm("  Include the task ID in PR title or branch name:"),
		guideVal("    feat: TSK-42 Add rate limiting"),
		guideVal("    feature/TSK-42-add-rate-limiting"),
		"",
		guideNorm("  PR opened  → task moves to In Review"),
		guideNorm("  PR merged  → task moves to Done"),
		"",
		guideSep(),
		guideHead("  Notifications"),
		"",
		guideVal("  sprintos notify config         # Add Slack or Discord channel"),
		guideVal("  sprintos notify list"),
		guideVal("  sprintos notify test"),
		"",
		guideSep(),
		guideHead("  Repo Init & Invitations"),
		"",
		guideVal("  sprintos init                  # Creates .sprintos with project ID"),
		guideVal("  sprintos join --token abc123   # Accept a team invitation"),
		"",
		guideNorm("  After init, all task commands auto-read the project — no --project flag needed."),
		"",
		guideSep(),
		guideHead("  REST API  (base: http://localhost:8090)"),
		"",
		guideNorm("  All endpoints require:  Authorization: Bearer <api-key>"),
		guideNorm("  Rate limit: 60 requests / minute per key."),
		"",
		kv("GET  /api/projects", "List projects"),
		kv("POST /api/projects", "Create project"),
		kv("GET  /api/tasks", "List tasks (filter: project_id, state_id)"),
		kv("POST /api/tasks", "Create task"),
		kv("PATCH /api/tasks/:id", "Update task"),
		kv("POST /api/tasks/:id/move", "Move task to state"),
		kv("DELETE /api/tasks/:id", "Delete task"),
		kv("GET  /api/states", "List board states"),
		kv("GET  /api/members", "List org members"),
		kv("GET  /api/health", "Health check (no auth)"),
		kv("GET  /api/docs", "OpenAPI spec (no auth)"),
		"",
	}
	return guideSection{icon: "⌨️", short: "CLI / API", title: "CLI Commands & REST API", lines: lines}
}
