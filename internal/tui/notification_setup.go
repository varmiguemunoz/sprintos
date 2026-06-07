package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/sprintos/internal/app"
)

type notifSetupMode int

const (
	notifModeChoose notifSetupMode = iota
	notifModeSlack
	notifModeDiscord
	notifModeWhatsApp
)

type NotificationSetupModel struct {
	mode     notifSetupMode
	cursor   int
	inputs   []textinput.Model
	loading  bool
	err      error
	orgID    uint
	notifSvc *app.NotificationService
}

type NotifSetupDoneMsg struct{ Err error }

var notifChoices = []string{"Slack", "Discord", "WhatsApp", "Skip for now"}

func NewNotificationSetupModel(orgID uint, notifSvc *app.NotificationService) NotificationSetupModel {
	return NotificationSetupModel{
		mode:     notifModeChoose,
		orgID:    orgID,
		notifSvc: notifSvc,
	}
}

func (m NotificationSetupModel) setupInputs(mode notifSetupMode) []textinput.Model {
	switch mode {
	case notifModeSlack:
		i := textinput.New()
		i.Placeholder = "https://hooks.slack.com/services/T.../B.../xxx"
		i.CharLimit = 300
		i.Focus()
		return []textinput.Model{i}
	case notifModeDiscord:
		i := textinput.New()
		i.Placeholder = "https://discord.com/api/webhooks/xxx/yyy"
		i.CharLimit = 300
		i.Focus()
		return []textinput.Model{i}
	case notifModeWhatsApp:
		i1 := textinput.New()
		i1.Placeholder = "http://your-evolution-api:8080"
		i1.CharLimit = 200
		i1.Focus()
		i2 := textinput.New()
		i2.Placeholder = "+1234567890"
		i2.CharLimit = 20
		return []textinput.Model{i1, i2}
	}
	return nil
}

func (m NotificationSetupModel) submitCmd() tea.Cmd {
	return func() tea.Msg {
		var channel, webhookURL string
		switch m.mode {
		case notifModeSlack:
			channel = "slack"
			webhookURL = m.inputs[0].Value()
		case notifModeDiscord:
			channel = "discord"
			webhookURL = m.inputs[0].Value()
		case notifModeWhatsApp:
			channel = "whatsapp"
			webhookURL = m.inputs[0].Value()
		}
		if err := m.notifSvc.SaveConfig(m.orgID, channel, webhookURL); err != nil {
			return NotifSetupDoneMsg{Err: err}
		}
		return NotifSetupDoneMsg{}
	}
}

func (m NotificationSetupModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m NotificationSetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(NotifSetupDoneMsg); ok {
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		return m, func() tea.Msg { return NavigateMsg{To: screenMCPSetup} }
	}

	if m.loading {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.mode == notifModeChoose {
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(notifChoices)-1 {
					m.cursor++
				}
			case "enter":
				switch m.cursor {
				case 0:
					m.mode = notifModeSlack
					m.inputs = m.setupInputs(notifModeSlack)
				case 1:
					m.mode = notifModeDiscord
					m.inputs = m.setupInputs(notifModeDiscord)
				case 2:
					m.mode = notifModeWhatsApp
					m.inputs = m.setupInputs(notifModeWhatsApp)
				case 3:
					return m, func() tea.Msg { return NavigateMsg{To: screenMCPSetup} }
				}
			case "esc", "ctrl+c":
				return m, func() tea.Msg { return NavigateMsg{To: screenMCPSetup} }
			}
			return m, nil
		}

		// Form mode
		switch msg.String() {
		case "esc":
			m.mode = notifModeChoose
			m.inputs = nil
			m.err = nil
		case "ctrl+c":
			return m, tea.Quit
		case "tab", "down":
			if len(m.inputs) > 1 {
				for i := range m.inputs {
					m.inputs[i].Blur()
				}
				next := (m.focusedInput() + 1) % len(m.inputs)
				m.inputs[next].Focus()
			}
		case "shift+tab", "up":
			if len(m.inputs) > 1 {
				for i := range m.inputs {
					m.inputs[i].Blur()
				}
				prev := (m.focusedInput() - 1 + len(m.inputs)) % len(m.inputs)
				m.inputs[prev].Focus()
			}
		case "enter":
			m.loading = true
			return m, m.submitCmd()
		}
	}

	if len(m.inputs) > 0 {
		focused := m.focusedInput()
		var cmd tea.Cmd
		m.inputs[focused], cmd = m.inputs[focused].Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m NotificationSetupModel) focusedInput() int {
	for i, inp := range m.inputs {
		if inp.Focused() {
			return i
		}
	}
	return 0
}

func (m NotificationSetupModel) View() string {
	if m.loading {
		return titleStyle.Render("SprintOS") + "\n" +
			normalStyle.Render("Step 4 of 5  ●●●●○  Connecting notifications...") + "\n"
	}

	s := titleStyle.Render("SprintOS") + "\n"
	s += normalStyle.Render("Step 4 of 5  ●●●●○  Connect notifications") + "\n\n"

	if m.mode == notifModeChoose {
		s += normalStyle.Render("Stay in the loop. Choose where to receive task updates:") + "\n\n"
		for i, ch := range notifChoices {
			if i == m.cursor {
				s += selectedStyle.Render("> "+ch) + "\n"
			} else {
				s += normalStyle.Render("  "+ch) + "\n"
			}
		}
		s += "\n" + normalStyle.Render("↑/↓ move  •  enter select  •  esc skip") + "\n"
		return s
	}

	var labels []string
	var hint string
	switch m.mode {
	case notifModeSlack:
		s += selectedStyle.Render("Slack Webhook URL") + "\n"
		labels = []string{"Webhook URL"}
		hint = "Get it at: api.slack.com/apps → Your App → Incoming Webhooks"
	case notifModeDiscord:
		s += selectedStyle.Render("Discord Webhook URL") + "\n"
		labels = []string{"Webhook URL"}
		hint = "Get it at: Discord Server → Edit Channel → Integrations → Webhooks"
	case notifModeWhatsApp:
		s += selectedStyle.Render("WhatsApp via Evolution API") + "\n"
		labels = []string{"Evolution API URL", "Phone number (with country code)"}
		hint = "Requires a self-hosted Evolution API instance"
	}

	s += normalStyle.Render(hint) + "\n\n"
	for i, label := range labels {
		if i == m.focusedInput() {
			s += selectedStyle.Render(label) + "\n"
		} else {
			s += normalStyle.Render(label) + "\n"
		}
		s += m.inputs[i].View() + "\n\n"
	}

	if m.err != nil {
		s += errorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())) + "\n\n"
	}

	s += normalStyle.Render("enter to connect  •  esc back  •  tab next field") + "\n"
	return s
}
