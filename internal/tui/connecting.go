package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type ConnectingModel struct {
	dots int
	err  error
}

type connectingTickMsg struct{}

func connectingTickCmd() tea.Cmd {
	return tea.Tick(400*time.Millisecond, func(_ time.Time) tea.Msg {
		return connectingTickMsg{}
	})
}

func NewConnectingModel() ConnectingModel {
	return ConnectingModel{}
}

func (m ConnectingModel) Init() tea.Cmd {
	return connectingTickCmd()
}

func (m ConnectingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case connectingTickMsg:
		m.dots = (m.dots + 1) % 4
		return m, connectingTickCmd()
	case tea.KeyMsg:
		return m, tea.Quit
	}
	return m, nil
}

func (m ConnectingModel) View() string {
	if m.err != nil {
		s := titleStyle.Render("SprintOS") + "\n\n"
		s += errorStyle.Render("Could not connect to the database:") + "\n"
		s += normalStyle.Render(m.err.Error()) + "\n\n"
		s += dimStyle.Render("Check your DATABASE_URL and network connection.") + "\n\n"
		s += renderHintBar("q", "quit") + "\n"
		return s
	}

	dots := ""
	for i := 0; i < m.dots; i++ {
		dots += "."
	}

	return titleStyle.Render("SprintOS") + "\n\n" +
		normalStyle.Render("Connecting"+dots) + "\n"
}
