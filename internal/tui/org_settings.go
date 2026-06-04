package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/domain"
)

type OrgSettingsModel struct {
	inputs  []textinput.Model
	focused int
	loading bool
	saved   bool
	err     error
	org     domain.Organization
	orgSvc  *app.OrganizationService
}

type OrgSettingsSavedMsg struct {
	Err error
}

func NewOrgSettingsModel(org domain.Organization, orgSvc *app.OrganizationService) OrgSettingsModel {
	inputs := make([]textinput.Model, 3)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Organization name"
	inputs[0].CharLimit = 100
	inputs[0].SetValue(org.Name)
	inputs[0].Focus()

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Description (optional)"
	inputs[1].CharLimit = 200
	if org.Description != nil {
		inputs[1].SetValue(*org.Description)
	}

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "+1234567890"
	inputs[2].CharLimit = 20
	inputs[2].SetValue(org.WhatsappNumber)

	return OrgSettingsModel{
		inputs: inputs,
		org:    org,
		orgSvc: orgSvc,
	}
}

func (m OrgSettingsModel) saveCmd() tea.Cmd {
	return func() tea.Msg {
		name := m.inputs[0].Value()
		description := m.inputs[1].Value()
		whatsapp := m.inputs[2].Value()

		if name == "" {
			return OrgSettingsSavedMsg{Err: fmt.Errorf("organization name is required")}
		}

		_, err := m.orgSvc.Update(m.org.ID, name, description, whatsapp)
		if err != nil {
			return OrgSettingsSavedMsg{Err: err}
		}

		return OrgSettingsSavedMsg{}
	}
}

func (m OrgSettingsModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m OrgSettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(OrgSettingsSavedMsg); ok {
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
		case "c":
			return m, func() tea.Msg {
				return NavigateMsg{To: screenConnections}
			}
		case "i":
			return m, func() tea.Msg {
				return NavigateMsg{To: screenInviteUser}
			}
		case "m":
			return m, func() tea.Msg {
				return NavigateMsg{To: screenMCPSetup}
			}
		case "esc":
			return m, func() tea.Msg {
				return NavigateMsg{To: screenDashboard}
			}
		case "ctrl+c":
			return m, tea.Quit
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
		case "L":
			return m, func() tea.Msg {
				return NavigateMsg{To: screenLogin}
			}
		}
	}

	var cmd tea.Cmd
	m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
	return m, cmd
}

func (m OrgSettingsModel) View() string {
	if m.loading {
		return titleStyle.Render("SprintOS — Organization Settings") +
			"\n\n" + normalStyle.Render("Saving...") + "\n"
	}

	labels := []string{"Organization name *", "Description", "WhatsApp number"}

	s := titleStyle.Render("SprintOS — Organization Settings") + "\n\n"

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

	if m.saved {
		s += selectedStyle.Render("✓ Changes saved successfully") + "\n\n"
	}

	s += normalStyle.Render("tab/↓ next • enter save • c notifications • i invite member  •  m mcp setup  •  L logout  •  esc back") + "\n"
	return s
}
