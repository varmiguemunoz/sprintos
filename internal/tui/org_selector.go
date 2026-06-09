package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/sprintos/internal/domain"
)

type OrgSelectorModel struct {
	orgs   []domain.Organization
	cursor int
}

func NewOrgSelectorModel(orgs []domain.Organization) OrgSelectorModel {
	return OrgSelectorModel{orgs: orgs}
}

func (m OrgSelectorModel) Init() tea.Cmd {
	return nil
}

func (m OrgSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		totalItems := len(m.orgs) + 1
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < totalItems-1 {
				m.cursor++
			}
		case "enter":
			if m.cursor < len(m.orgs) {
				org := m.orgs[m.cursor]
				return m, func() tea.Msg {
					return NavigateMsg{To: screenCEODashboard, Org: org}
				}
			}
			return m, func() tea.Msg {
				return NavigateMsg{To: screenCreateOrg}
			}
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m OrgSelectorModel) View() string {
	s := titleStyle.Render("SprintOS — Select Organization") + "\n\n"
	s += normalStyle.Render("You belong to multiple organizations. Choose one to continue:") + "\n\n"

	for i, org := range m.orgs {
		if i == m.cursor {
			s += selectedStyle.Render(fmt.Sprintf("> %s", org.Name)) + "\n"
		} else {
			s += normalStyle.Render(fmt.Sprintf("  %s", org.Name)) + "\n"
		}
	}

	createIdx := len(m.orgs)
	if m.cursor == createIdx {
		s += selectedStyle.Render("> + Create a new organization") + "\n"
	} else {
		s += normalStyle.Render("  + Create a new organization") + "\n"
	}

	s += "\n" + renderHintBar("↑/↓", "move", "enter", "select", "q", "quit") + "\n"
	return s
}
