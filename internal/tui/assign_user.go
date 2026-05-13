package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/command_pm_app/internal/app"
	"github.com/varmiguemunoz/command_pm_app/internal/domain"
)

type AssignUserModel struct {
	task    domain.Task
	project domain.Project
	members []domain.TeamMember
	cursor  int
	loading bool
	err     error
	orgID   uint
	teamSvc *app.TeamService
	taskSvc *app.TaskService
}

type MembersLoadedMsg struct {
	Members []domain.TeamMember
	Err     error
}

type UserAssignedMsg struct {
	Err error
}

func NewAssignUserModel(
	task domain.Task,
	project domain.Project,
	orgID uint,
	teamSvc *app.TeamService,
	taskSvc *app.TaskService,
) AssignUserModel {
	return AssignUserModel{
		task:    task,
		project: project,
		orgID:   orgID,
		loading: true,
		teamSvc: teamSvc,
		taskSvc: taskSvc,
	}
}

func (m AssignUserModel) loadMembersCmd() tea.Cmd {
	return func() tea.Msg {
		members, err := m.teamSvc.ListMembers(m.orgID)
		return MembersLoadedMsg{Members: members, Err: err}
	}
}

func (m AssignUserModel) assignCmd(userID *uint) tea.Cmd {
	return func() tea.Msg {
		_, err := m.taskSvc.AssignUser(m.task.ID, userID)
		return UserAssignedMsg{Err: err}
	}
}

func (m AssignUserModel) Init() tea.Cmd {
	return m.loadMembersCmd()
}

func (m AssignUserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(MembersLoadedMsg); ok {
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.members = msg.Members
		return m, nil
	}

	if msg, ok := msg.(UserAssignedMsg); ok {
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		task := m.task
		project := m.project
		return m, func() tea.Msg {
			return NavigateMsg{To: screenTaskDetail, Task: task, Project: project}
		}
	}

	if m.loading {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			// +1 for the "Unassign" option at index 0
			if m.cursor < len(m.members) {
				m.cursor++
			}
		case "enter":
			if m.cursor == 0 {
				// Unassign
				return m, m.assignCmd(nil)
			}
			// Assign selected member
			memberIndex := m.cursor - 1
			userID := m.members[memberIndex].UserID
			return m, m.assignCmd(&userID)
		case "esc":
			task := m.task
			project := m.project
			return m, func() tea.Msg {
				return NavigateMsg{To: screenTaskDetail, Task: task, Project: project}
			}
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m AssignUserModel) View() string {
	if m.loading {
		return titleStyle.Render("SprintOS — Assign User") +
			"\n\n" + normalStyle.Render("Loading team members...") + "\n"
	}

	s := titleStyle.Render(fmt.Sprintf("SprintOS — Assign: %s", m.task.Title)) + "\n\n"

	if m.task.Assignee != nil {
		s += normalStyle.Render(fmt.Sprintf("Currently assigned to: %s", m.task.Assignee.Name)) + "\n\n"
	} else {
		s += normalStyle.Render("Currently unassigned") + "\n\n"
	}

	// Unassign option
	if m.cursor == 0 {
		s += errorStyle.Render("> Unassign") + "\n"
	} else {
		s += normalStyle.Render("  Unassign") + "\n"
	}

	// Team members
	for i, member := range m.members {
		name := member.User.Name
		if name == "" {
			name = member.User.Email
		}

		isCurrentAssignee := m.task.AssignedTo != nil && *m.task.AssignedTo == member.UserID
		label := name
		if isCurrentAssignee {
			label = name + " ✓"
		}

		if m.cursor == i+1 {
			s += selectedStyle.Render(fmt.Sprintf("> %s", label)) + "\n"
		} else {
			s += normalStyle.Render(fmt.Sprintf("  %s", label)) + "\n"
		}
	}

	if m.err != nil {
		s += "\n" + errorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())) + "\n"
	}

	s += "\n" + normalStyle.Render("↑/↓ move  •  enter to select  •  esc back") + "\n"
	return s
}
