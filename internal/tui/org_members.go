package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/domain"
)

type OrgMembersModel struct {
	members       []domain.TeamMember
	cursor        int
	loading       bool
	confirming    bool
	err           error
	org           domain.Organization
	currentUserID uint
	isOwner       bool
	teamSvc       *app.TeamService
}

type OrgMembersLoadedMsg struct {
	Members []domain.TeamMember
	Err     error
}

type MemberRoleUpdatedMsg struct {
	Err error
}

type MemberRemovedMsg struct {
	Err error
}

func NewOrgMembersModel(org domain.Organization, currentUserID uint, teamSvc *app.TeamService) OrgMembersModel {
	return OrgMembersModel{
		loading:       true,
		org:           org,
		currentUserID: currentUserID,
		isOwner:       org.OwnerID == currentUserID,
		teamSvc:       teamSvc,
	}
}

func (m OrgMembersModel) loadCmd() tea.Cmd {
	return func() tea.Msg {
		members, err := m.teamSvc.ListMembers(m.org.ID)
		return OrgMembersLoadedMsg{Members: members, Err: err}
	}
}

func (m OrgMembersModel) cycleRoleCmd(member domain.TeamMember) tea.Cmd {
	return func() tea.Msg {
		roles := []string{domain.RoleMember, domain.RoleManager, domain.RoleOwner}
		nextRole := domain.RoleMember
		for i, r := range roles {
			if r == member.Role {
				nextRole = roles[(i+1)%len(roles)]
				break
			}
		}
		err := m.teamSvc.UpdateMemberRole(member.UserID, m.org.ID, nextRole)
		return MemberRoleUpdatedMsg{Err: err}
	}
}

func (m OrgMembersModel) removeCmd(member domain.TeamMember) tea.Cmd {
	return func() tea.Msg {
		err := m.teamSvc.RemoveMemberByID(member.ID)
		return MemberRemovedMsg{Err: err}
	}
}

func (m OrgMembersModel) Init() tea.Cmd {
	return m.loadCmd()
}

func (m OrgMembersModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case OrgMembersLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.members = msg.Members
		return m, nil

	case MemberRoleUpdatedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.err = nil
		return m, m.loadCmd()

	case MemberRemovedMsg:
		m.loading = false
		m.confirming = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.err = nil
		if m.cursor >= len(m.members)-1 && m.cursor > 0 {
			m.cursor--
		}
		return m, m.loadCmd()

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		if m.confirming {
			switch msg.String() {
			case "y", "Y":
				if len(m.members) > 0 {
					selected := m.members[m.cursor]
					m.loading = true
					m.confirming = false
					return m, m.removeCmd(selected)
				}
			case "n", "N", "esc":
				m.confirming = false
			}
			return m, nil
		}

		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.members)-1 {
				m.cursor++
			}
		case "r":
			if m.isOwner && len(m.members) > 0 {
				selected := m.members[m.cursor]
				if selected.UserID == m.org.OwnerID {
					return m, nil
				}
				m.loading = true
				return m, m.cycleRoleCmd(selected)
			}
		case "D":
			if m.isOwner && len(m.members) > 0 {
				selected := m.members[m.cursor]
				if selected.UserID == m.org.OwnerID {
					return m, nil
				}
				m.confirming = true
			}
		case "esc":
			return m, func() tea.Msg {
				return NavigateMsg{To: screenOrgSettings}
			}
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m OrgMembersModel) View() string {
	if m.loading {
		return titleStyle.Render("SprintOS — Team Members") +
			"\n\n" + normalStyle.Render("Loading members...") + "\n"
	}

	s := titleStyle.Render(fmt.Sprintf("SprintOS — %s · Team Members", m.org.Name)) + "\n\n"

	if m.err != nil {
		s += errorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())) + "\n\n"
	}

	if len(m.members) == 0 {
		s += dimStyle.Render("No members yet.") + "\n"
	} else {
		header := fmt.Sprintf("  %-24s %-32s %-10s", "Name", "Email", "Role")
		s += dimStyle.Render(header) + "\n"
		s += dimStyle.Render("  " + strings.Repeat("─", 66)) + "\n"

		for i, member := range m.members {
			name := truncate(member.User.Name, 22)
			email := truncate(member.User.Email, 30)
			role := capitalizeFirst(member.Role)

			isCurrentUser := member.UserID == m.currentUserID
			isOrgOwner := member.UserID == m.org.OwnerID

			if isOrgOwner {
				role += " ★"
			}

			line := fmt.Sprintf("  %-24s %-32s %-10s", name, email, role)

			if i == m.cursor {
				if isCurrentUser {
					s += selectedStyle.Render("> " + line[2:]) + "\n"
				} else {
					s += selectedStyle.Render("> " + line[2:]) + "\n"
				}
			} else {
				if isCurrentUser {
					s += valueStyle.Render(line) + "\n"
				} else {
					s += normalStyle.Render(line) + "\n"
				}
			}
		}
	}

	if m.confirming && len(m.members) > 0 {
		selected := m.members[m.cursor]
		s += "\n" + errorStyle.Render(fmt.Sprintf("⚠  Remove %s from the organization?", selected.User.Name)) + "\n"
		s += dimStyle.Render("   This cannot be undone.") + "\n\n"
		s += renderHintBar("y", "confirm", "n", "cancel", "esc", "cancel") + "\n"
		return s
	}

	if m.isOwner {
		s += "\n" + renderHintBar("↑/↓", "move", "r", "cycle role", "D", "remove", "esc", "back") + "\n"
	} else {
		s += "\n" + renderHintBar("↑/↓", "move", "esc", "back") + "\n"
	}

	return s
}
