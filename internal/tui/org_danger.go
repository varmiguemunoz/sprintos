package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/domain"
)

type OrgDangerModel struct {
	org         domain.Organization
	currentUser *domain.User
	isOwner     bool
	confirming  bool
	loading     bool
	err         error
	orgSvc      *app.OrganizationService
	teamSvc     *app.TeamService
}

type OrgDeletedMsg struct {
	Err error
}

type OrgLeftMsg struct {
	Err error
}

func NewOrgDangerModel(
	org domain.Organization,
	currentUser *domain.User,
	isOwner bool,
	orgSvc *app.OrganizationService,
	teamSvc *app.TeamService,
) OrgDangerModel {
	return OrgDangerModel{
		org:         org,
		currentUser: currentUser,
		isOwner:     isOwner,
		orgSvc:      orgSvc,
		teamSvc:     teamSvc,
	}
}

func (m OrgDangerModel) deleteCmd() tea.Cmd {
	return func() tea.Msg {
		if err := m.orgSvc.Delete(m.org.ID); err != nil {
			return OrgDeletedMsg{Err: err}
		}
		return OrgDeletedMsg{}
	}
}

func (m OrgDangerModel) leaveCmd() tea.Cmd {
	return func() tea.Msg {
		if err := m.teamSvc.RemoveMember(m.currentUser.ID, m.org.ID); err != nil {
			return OrgLeftMsg{Err: err}
		}
		return OrgLeftMsg{}
	}
}

func (m OrgDangerModel) Init() tea.Cmd {
	return nil
}

func (m OrgDangerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case OrgDeletedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			m.confirming = false
			return m, nil
		}
		return m, func() tea.Msg {
			return OrgContextClearedMsg{}
		}

	case OrgLeftMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			m.confirming = false
			return m, nil
		}
		return m, func() tea.Msg {
			return OrgContextClearedMsg{}
		}

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		if m.confirming {
			switch msg.String() {
			case "y", "Y":
				m.loading = true
				m.err = nil
				if m.isOwner {
					return m, m.deleteCmd()
				}
				return m, m.leaveCmd()
			case "n", "N", "esc":
				m.confirming = false
			}
			return m, nil
		}

		switch msg.String() {
		case "enter":
			m.confirming = true
			m.err = nil
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

func (m OrgDangerModel) View() string {
	if m.loading {
		if m.isOwner {
			return titleStyle.Render("SprintOS — Danger Zone") +
				"\n\n" + normalStyle.Render("Deleting organization...") + "\n"
		}
		return titleStyle.Render("SprintOS — Danger Zone") +
			"\n\n" + normalStyle.Render("Leaving organization...") + "\n"
	}

	s := titleStyle.Render("SprintOS — Danger Zone") + "\n\n"

	if m.isOwner {
		s += errorStyle.Render("⚠  Delete Organization") + "\n\n"
		s += normalStyle.Render(fmt.Sprintf("You are about to permanently delete \"%s\".", m.org.Name)) + "\n"
		s += normalStyle.Render("This will remove all projects, sprints, tasks, and team members.") + "\n"
		s += errorStyle.Render("This action cannot be undone.") + "\n\n"
	} else {
		s += errorStyle.Render("⚠  Leave Organization") + "\n\n"
		s += normalStyle.Render(fmt.Sprintf("You are about to leave \"%s\".", m.org.Name)) + "\n"
		s += normalStyle.Render("You will lose access to all projects and tasks in this organization.") + "\n\n"
	}

	if m.err != nil {
		s += errorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())) + "\n\n"
	}

	if m.confirming {
		if m.isOwner {
			s += errorStyle.Render("Are you sure you want to delete this organization?") + "\n\n"
		} else {
			s += errorStyle.Render("Are you sure you want to leave this organization?") + "\n\n"
		}
		s += renderHintBar("y", "confirm", "n", "cancel", "esc", "cancel") + "\n"
		return s
	}

	if m.isOwner {
		s += renderHintBar("enter", "delete org", "esc", "back") + "\n"
	} else {
		s += renderHintBar("enter", "leave org", "esc", "back") + "\n"
	}

	return s
}
