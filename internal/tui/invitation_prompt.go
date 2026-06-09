package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/domain"
)

type InvitationPromptModel struct {
	invitations   []domain.Invitation
	memberOrgs    []domain.Organization
	cursor        int
	loading       bool
	err           error
	currentUser   *domain.User
	invitationSvc *app.InvitationService
	teamSvc       *app.TeamService
}

type InvitationAcceptedMsg struct {
	OrgID uint
	Org   domain.Organization
	Err   error
}

type InvitationDeclinedMsg struct {
	Index int
	Err   error
}

func NewInvitationPromptModel(
	invitations []domain.Invitation,
	memberOrgs []domain.Organization,
	currentUser *domain.User,
	invitationSvc *app.InvitationService,
	teamSvc *app.TeamService,
) InvitationPromptModel {
	return InvitationPromptModel{
		invitations:   invitations,
		memberOrgs:    memberOrgs,
		currentUser:   currentUser,
		invitationSvc: invitationSvc,
		teamSvc:       teamSvc,
	}
}

func (m InvitationPromptModel) acceptCmd(inv domain.Invitation) tea.Cmd {
	return func() tea.Msg {
		accepted, err := m.invitationSvc.Accept(inv.Token)
		if err != nil {
			return InvitationAcceptedMsg{Err: err}
		}

		_, err = m.teamSvc.AddMember(m.currentUser.ID, inv.OrganizationID, inv.Role)
		if err != nil {
			return InvitationAcceptedMsg{Err: fmt.Errorf("could not add you to the organization: %w", err)}
		}

		return InvitationAcceptedMsg{OrgID: accepted.OrganizationID, Org: accepted.Organization}
	}
}

func (m InvitationPromptModel) declineCmd(index int, token string) tea.Cmd {
	return func() tea.Msg {
		if err := m.invitationSvc.Decline(token); err != nil {
			return InvitationDeclinedMsg{Index: index, Err: err}
		}
		return InvitationDeclinedMsg{Index: index}
	}
}

func (m InvitationPromptModel) Init() tea.Cmd {
	return nil
}

func (m InvitationPromptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case InvitationAcceptedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		return m, func() tea.Msg {
			return NavigateMsg{To: screenCEODashboard, Org: msg.Org}
		}

	case InvitationDeclinedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}

		remaining := make([]domain.Invitation, 0, len(m.invitations)-1)
		for i, inv := range m.invitations {
			if i != msg.Index {
				remaining = append(remaining, inv)
			}
		}
		m.invitations = remaining
		m.err = nil

		if m.cursor >= len(m.invitations) && m.cursor > 0 {
			m.cursor--
		}

		if len(m.invitations) > 0 {
			return m, nil
		}

		if len(m.memberOrgs) == 0 {
			return m, func() tea.Msg {
				return NavigateMsg{To: screenCreateOrg}
			}
		}

		if len(m.memberOrgs) == 1 {
			org := m.memberOrgs[0]
			return m, func() tea.Msg {
				return NavigateMsg{To: screenCEODashboard, Org: org}
			}
		}

		return m, func() tea.Msg {
			return NavigateMsg{To: screenOrgSelector}
		}

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.invitations)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.invitations) == 0 {
				return m, nil
			}
			inv := m.invitations[m.cursor]
			m.loading = true
			m.err = nil
			return m, m.acceptCmd(inv)
		case "d":
			if len(m.invitations) == 0 {
				return m, nil
			}
			idx := m.cursor
			token := m.invitations[idx].Token
			m.loading = true
			m.err = nil
			return m, m.declineCmd(idx, token)
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m InvitationPromptModel) View() string {
	if m.loading {
		return titleStyle.Render("SprintOS — Pending Invitations") +
			"\n\n" + normalStyle.Render("Processing...") + "\n"
	}

	s := titleStyle.Render("SprintOS — Pending Invitations") + "\n\n"

	if len(m.invitations) == 0 {
		s += normalStyle.Render("No pending invitations.") + "\n"
		return s
	}

	s += normalStyle.Render("You have been invited to join the following organization(s):") + "\n\n"

	for i, inv := range m.invitations {
		role := capitalizeFirst(inv.Role)
		line := fmt.Sprintf("Join \"%s\" as %s", inv.Organization.Name, role)
		if i == m.cursor {
			s += selectedStyle.Render("> "+line) + "\n"
		} else {
			s += normalStyle.Render("  "+line) + "\n"
		}
	}

	if m.err != nil {
		s += "\n" + errorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())) + "\n"
	}

	s += "\n" + renderHintBar("↑/↓", "move", "enter", "accept", "d", "decline", "q", "quit") + "\n"
	return s
}
