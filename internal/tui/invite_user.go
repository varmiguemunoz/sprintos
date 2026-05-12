package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/command_pm_app/internal/app"
	"github.com/varmiguemunoz/command_pm_app/internal/domain"
	"github.com/varmiguemunoz/command_pm_app/internal/infrastructure/email"
)

type InviteUserModel struct {
	input         textinput.Model
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

		inv, err := m.invitationSvc.Create(emailAddr, m.org.ID)
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
		return m, nil
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
		case "enter":
			m.loading = true
			m.sent = false
			m.err = nil
			return m, m.sendInviteCmd()
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m InviteUserModel) View() string {
	if m.loading {
		return titleStyle.Render("SprintOS — Invite Member") +
			"\n\n" + normalStyle.Render("Sending invitation...") + "\n"
	}

	s := titleStyle.Render("SprintOS — Invite Member") + "\n\n"
	s += normalStyle.Render(fmt.Sprintf("Invite someone to join \"%s\"", m.org.Name)) + "\n\n"
	s += selectedStyle.Render("Email address") + "\n"
	s += m.input.View() + "\n\n"

	if m.err != nil {
		s += errorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())) + "\n\n"
	}

	if m.sent {
		s += selectedStyle.Render("✓ Invitation sent! They'll receive a command to run in their terminal.") + "\n\n"
	}

	s += normalStyle.Render("enter to send  •  esc back to settings") + "\n"
	return s
}
