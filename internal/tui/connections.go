package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/domain"
)

type ConnectionsModel struct {
	configs    []domain.NotificationConfig
	cursor     int
	adding     bool
	addMode    notifSetupMode
	addCursor  int
	inputs     []textinput.Model
	loading    bool
	err        error
	saved      bool
	orgID      uint
	notifSvc   *app.NotificationService
}

type ConnectionsLoadedMsg struct {
	Configs []domain.NotificationConfig
	Err     error
}

type ConnectionSavedMsg struct{ Err error }

func NewConnectionsModel(orgID uint, notifSvc *app.NotificationService) ConnectionsModel {
	return ConnectionsModel{
		orgID:    orgID,
		notifSvc: notifSvc,
		loading:  true,
	}
}

func (m ConnectionsModel) loadCmd() tea.Cmd {
	return func() tea.Msg {
		configs, err := m.notifSvc.ListConfigs(m.orgID)
		return ConnectionsLoadedMsg{Configs: configs, Err: err}
	}
}

func (m ConnectionsModel) Init() tea.Cmd {
	return m.loadCmd()
}

func (m ConnectionsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(ConnectionsLoadedMsg); ok {
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.configs = msg.Configs
		return m, nil
	}

	if msg, ok := msg.(ConnectionSavedMsg); ok {
		m.loading = false
		m.adding = false
		m.inputs = nil
		m.addMode = notifModeChoose
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.saved = true
		return m, m.loadCmd()
	}

	if m.loading {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.adding {
			if m.addMode == notifModeChoose {
				switch msg.String() {
				case "up", "k":
					if m.addCursor > 0 { m.addCursor-- }
				case "down", "j":
					if m.addCursor < len(notifChoices)-2 { m.addCursor++ }
				case "enter":
					switch m.addCursor {
					case 0: m.addMode = notifModeSlack; m.inputs = (&NotificationSetupModel{}).setupInputs(notifModeSlack)
					case 1: m.addMode = notifModeDiscord; m.inputs = (&NotificationSetupModel{}).setupInputs(notifModeDiscord)
					case 2: m.addMode = notifModeWhatsApp; m.inputs = (&NotificationSetupModel{}).setupInputs(notifModeWhatsApp)
					}
				case "esc":
					m.adding = false
					m.addMode = notifModeChoose
				}
				return m, nil
			}

			switch msg.String() {
			case "esc":
				m.addMode = notifModeChoose
				m.inputs = nil
			case "enter":
				m.loading = true
				orgID := m.orgID
				notifSvc := m.notifSvc
				mode := m.addMode
				inputs := m.inputs
				return m, func() tea.Msg {
					var channel, webhookURL string
					switch mode {
					case notifModeSlack: channel, webhookURL = "slack", inputs[0].Value()
					case notifModeDiscord: channel, webhookURL = "discord", inputs[0].Value()
					case notifModeWhatsApp: channel, webhookURL = "whatsapp", inputs[0].Value()
					}
					err := notifSvc.SaveConfig(orgID, channel, webhookURL)
					return ConnectionSavedMsg{Err: err}
				}
			case "tab":
				if len(m.inputs) > 1 {
					for i := range m.inputs { m.inputs[i].Blur() }
					focused := 0
					for i, inp := range m.inputs { if inp.Focused() { focused = i } }
					m.inputs[(focused+1)%len(m.inputs)].Focus()
				}
			}

			if len(m.inputs) > 0 {
				focused := 0
				for i, inp := range m.inputs { if inp.Focused() { focused = i } }
				var cmd tea.Cmd
				m.inputs[focused], cmd = m.inputs[focused].Update(msg)
				return m, cmd
			}
			return m, nil
		}

		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 { m.cursor-- }
		case "down", "j":
			if m.cursor < len(m.configs)-1 { m.cursor++ }
		case "n":
			m.adding = true
			m.addMode = notifModeChoose
			m.addCursor = 0
			m.saved = false
			m.err = nil
		case "x":
			if len(m.configs) > 0 {
				id := m.configs[m.cursor].ID
				m.loading = true
				notifSvc := m.notifSvc
				return m, func() tea.Msg {
					notifSvc.DeleteConfig(id)
					return ConnectionSavedMsg{}
				}
			}
		case "esc", "ctrl+c":
			return m, func() tea.Msg { return NavigateMsg{To: screenOrgSettings} }
		}
	}
	return m, nil
}

func (m ConnectionsModel) View() string {
	s := titleStyle.Render("SprintOS — Connections") + "\n\n"

	if m.loading {
		return s + normalStyle.Render("Loading...") + "\n"
	}

	if m.adding {
		if m.addMode == notifModeChoose {
			s += selectedStyle.Render("Choose channel to connect:") + "\n\n"
			choices := []string{"Slack", "Discord", "WhatsApp"}
			for i, ch := range choices {
				if i == m.addCursor {
					s += selectedStyle.Render("> "+ch) + "\n"
				} else {
					s += normalStyle.Render("  "+ch) + "\n"
				}
			}
			s += "\n" + normalStyle.Render("↑/↓ move  •  enter select  •  esc back") + "\n"
			return s
		}

		var label, hint string
		switch m.addMode {
		case notifModeSlack:
			label = "Slack Webhook URL"
			hint = "api.slack.com/apps → Your App → Incoming Webhooks"
		case notifModeDiscord:
			label = "Discord Webhook URL"
			hint = "Discord Server → Edit Channel → Integrations → Webhooks"
		case notifModeWhatsApp:
			label = "Evolution API URL"
			hint = "Requires self-hosted Evolution API"
		}
		s += selectedStyle.Render(label) + "\n"
		s += normalStyle.Render(hint) + "\n\n"
		if len(m.inputs) > 0 {
			s += m.inputs[0].View() + "\n\n"
		}
		if m.err != nil {
			s += errorStyle.Render(m.err.Error()) + "\n\n"
		}
		s += normalStyle.Render("enter to save  •  esc back") + "\n"
		return s
	}

	if len(m.configs) == 0 {
		s += normalStyle.Render("No notification channels connected.") + "\n\n"
	} else {
		s += selectedStyle.Render("Active connections:") + "\n\n"
		for i, c := range m.configs {
			status := "✓ enabled"
			if !c.Enabled { status = "✗ disabled" }
			preview := c.WebhookURL
			if len(preview) > 35 { preview = preview[:35] + "..." }
			line := fmt.Sprintf("%-10s  %-12s  %s", c.Channel, status, preview)
			if i == m.cursor {
				s += selectedStyle.Render("> "+line) + "\n"
			} else {
				s += normalStyle.Render("  "+line) + "\n"
			}
		}
	}

	if m.saved {
		s += "\n" + selectedStyle.Render("✓ Connection saved") + "\n"
	}
	if m.err != nil {
		s += "\n" + errorStyle.Render(m.err.Error()) + "\n"
	}

	s += "\n" + normalStyle.Render("n add  •  x remove selected  •  esc back to settings") + "\n"
	return s
}
