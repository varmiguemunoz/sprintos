package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/domain"
)

type CreateOrgModel struct {
	inputs  []textinput.Model
	focused int
	loading bool
	err     error
	ownerID uint
	orgSvc  *app.OrganizationService
	teamSvc *app.TeamService
}

type OrgCreatedMsg struct {
	Org *domain.Organization
	Err error
}

func NewCreateOrgModel(ownerID uint, orgSvc *app.OrganizationService, teamSvc *app.TeamService) CreateOrgModel {
	inputs := make([]textinput.Model, 3)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Acme Corp"
	inputs[0].CharLimit = 100
	inputs[0].Focus()

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "A short description (optional)"
	inputs[1].CharLimit = 200

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "+1234567890"
	inputs[2].CharLimit = 20

	inputs = append(inputs, textinput.New())
	inputs[3].Placeholder = "SPR (2-5 uppercase letters used as task prefix)"
	inputs[3].CharLimit = 5

	return CreateOrgModel{
		inputs:  inputs,
		focused: 0,
		ownerID: ownerID,
		orgSvc:  orgSvc,
		teamSvc: teamSvc,
	}
}

func (m CreateOrgModel) submitCmd() tea.Cmd {
	return func() tea.Msg {
		name := m.inputs[0].Value()
		description := m.inputs[1].Value()
		whatsapp := m.inputs[2].Value()

		if name == "" {
			return OrgCreatedMsg{Err: fmt.Errorf("organization name is required")}
		}
		if whatsapp == "" {
			return OrgCreatedMsg{Err: fmt.Errorf("whatsapp number is required")}
		}

		prefix := strings.ToUpper(strings.TrimSpace(m.inputs[3].Value()))
		if prefix == "" {
			prefix = "TSK"
		}

		org, err := m.orgSvc.CreateWithPrefix(name, description, whatsapp, prefix, m.ownerID)
		if err != nil {
			return OrgCreatedMsg{Err: err}
		}

		_, err = m.teamSvc.AddMember(m.ownerID, org.ID, "owner")
		if err != nil {
			return OrgCreatedMsg{Err: err}
		}

		return OrgCreatedMsg{Org: org}
	}
}

func (m CreateOrgModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m CreateOrgModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	if msg, ok := msg.(OrgCreatedMsg); ok {
		if msg.Err != nil {
			m.err = msg.Err
			m.loading = false
			return m, nil
		}
		org := *msg.Org
		return m, func() tea.Msg {
			return NavigateMsg{To: screenCreateProject, Org: org}
		}
	}

	if m.loading {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "tab", "down":
			m.inputs[m.focused].Blur()
			m.focused = (m.focused + 1) % len(m.inputs)
			m.inputs[m.focused].Focus()
		case "shift+tab", "up":
			m.inputs[m.focused].Blur()
			m.focused--
			if m.focused < 0 {
				m.focused = len(m.inputs) - 1
			}
			m.inputs[m.focused].Focus()
		case "enter":
			if m.focused == len(m.inputs)-1 {
				m.loading = true
				return m, m.submitCmd()
			}
			m.inputs[m.focused].Blur()
			m.focused++
			m.inputs[m.focused].Focus()
		}

	case OrgCreatedMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.loading = false
			return m, nil
		}
		org := *msg.Org
		return m, func() tea.Msg {
			return NavigateMsg{To: screenCreateProject, Org: org}
		}
	}

	var cmd tea.Cmd
	m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
	return m, cmd
}

func (m CreateOrgModel) View() string {
	if m.loading {
		return titleStyle.Render("SprintOS — Create Organization") +
			"\n\n" + normalStyle.Render("Creating your organization...") + "\n"
	}

	labels := []string{"Organization name *", "Description", "WhatsApp number *", "Task ID prefix * (e.g. SPR)"}

	s := titleStyle.Render("SprintOS — Set up your organization") + "\n\n"
	s += normalStyle.Render("This is required to get started. You can edit these details later.") + "\n\n"

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

	s += normalStyle.Render("tab/↓ next field  •  shift+tab/↑ previous  •  enter to confirm  •  esc to quit") + "\n"
	return s
}
