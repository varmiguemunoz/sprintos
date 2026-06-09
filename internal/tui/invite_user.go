package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/domain"
	"github.com/varmiguemunoz/sprintos/internal/infrastructure/email"
)

type InviteUserModel struct {
	input         textinput.Model
	focusField    int
	roles         []string
	roleIdx       int
	loading       bool
	sent          bool
	err           error
	org           domain.Organization
	invitationSvc *app.InvitationService
}

type InviteSentMsg struct {
	Err error
}

func NewInviteUserModel(org domain.Organization, invitationSvc *app.InvitationService) InviteUserModel {
	input := textinput.New()
	input.Placeholder = "colleague@email.com"
	input.CharLimit = 200
	input.Focus()

	return InviteUserModel{
		input:         input,
		focusField:    0,
		roles:         []string{"Owner", "Manager", "Member"},
		roleIdx:       1,
		org:           org,
		invitationSvc: invitationSvc,
	}
}

func (m InviteUserModel) sendInviteCmd() tea.Cmd {
	return func() tea.Msg {
		emailAddr := m.input.Value()
		if emailAddr == "" {
			return InviteSentMsg{Err: fmt.Errorf("email address is required")}
		}

		role := strings.ToLower(m.roles[m.roleIdx])

		inv, err := m.invitationSvc.Create(emailAddr, m.org.ID, role)
		if err != nil {
			return InviteSentMsg{Err: err}
		}

		if err := email.SendInvitation(emailAddr, m.org.Name, inv.Token); err != nil {
			return InviteSentMsg{Err: fmt.Errorf("invitation created but email failed: %w", err)}
		}

		return InviteSentMsg{}
	}
}

func (m InviteUserModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m InviteUserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(InviteSentMsg); ok {
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.sent = true
		m.input.SetValue("")
		m.roleIdx = 1
		m.focusField = 0
		m.input.Focus()
		return m, textinput.Blink
	}

	if m.loading {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg {
				return NavigateMsg{To: screenOrgSettings}
			}
		case "ctrl+c":
			return m, tea.Quit
		case "tab", "down":
			m.sent = false
			m.err = nil
			if m.focusField == 0 {
				m.input.Blur()
				m.focusField = 1
			} else {
				m.focusField = 0
				m.input.Focus()
				return m, textinput.Blink
			}
		case "shift+tab", "up":
			m.sent = false
			m.err = nil
			if m.focusField == 1 {
				m.focusField = 0
				m.input.Focus()
				return m, textinput.Blink
			}
		case "left":
			if m.focusField == 1 {
				m.roleIdx--
				if m.roleIdx < 0 {
					m.roleIdx = len(m.roles) - 1
				}
			}
		case "right":
			if m.focusField == 1 {
				m.roleIdx = (m.roleIdx + 1) % len(m.roles)
			}
		case "enter":
			m.loading = true
			m.sent = false
			m.err = nil
			return m, m.sendInviteCmd()
		}
	}

	if m.focusField == 0 {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m InviteUserModel) View() string {
	if m.loading {
		return titleStyle.Render("SprintOS — Invite Member") +
			"\n\n" + normalStyle.Render("Sending invitation...") + "\n"
	}

	s := titleStyle.Render("SprintOS — Invite Member") + "\n\n"
	s += normalStyle.Render(fmt.Sprintf("Invite someone to join \"%s\"", m.org.Name)) + "\n\n"

	if m.focusField == 0 {
		s += selectedStyle.Render("Email address") + "\n"
	} else {
		s += normalStyle.Render("Email address") + "\n"
	}
	s += m.input.View() + "\n\n"

	if m.focusField == 1 {
		s += selectedStyle.Render("Role") + "\n"
	} else {
		s += normalStyle.Render("Role") + "\n"
	}

	roleRow := ""
	for i, r := range m.roles {
		if i == m.roleIdx {
			roleRow += selectedStyle.Render("[ " + r + " ]")
		} else {
			roleRow += dimStyle.Render("  " + r + "  ")
		}
		if i < len(m.roles)-1 {
			roleRow += "  "
		}
	}
	s += roleRow + "\n"

	if m.focusField == 1 {
		s += dimStyle.Render("← → to change role") + "\n"
	}

	s += "\n"

	if m.err != nil {
		s += errorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())) + "\n\n"
	}

	if m.sent {
		s += selectedStyle.Render("✓ Invitation sent! They'll receive a command to run in their terminal.") + "\n\n"
	}

	s += normalStyle.Render("tab to switch field  •  enter to send  •  esc back to settings") + "\n"
	return s
}
