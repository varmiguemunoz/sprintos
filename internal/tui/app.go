package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/command_pm_app/internal/infrastructure/auth"
)

type screen int

const (
	screenLogin screen = iota
	screenDashboard
)

type AppModel struct {
	activeScreen screen
	currentModel tea.Model
}

func (m AppModel) Init() tea.Cmd {
	return m.currentModel.Init()
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, cmd := m.currentModel.Update(msg)
	m.currentModel = updated
	return m, cmd
}

func (m AppModel) View() string {
	return m.currentModel.View()
}

func Start() error {
	auth.SetupProviders()

	model := AppModel{
		activeScreen: screenLogin,
		currentModel: NewLoginModel(),
	}

	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running app: %w", err)
	}

	return nil
}
