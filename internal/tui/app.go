package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/markbates/goth"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/domain"
	"github.com/varmiguemunoz/sprintos/internal/infrastructure/auth"
	"gorm.io/gorm"
)

type screen int

const (
	screenLogin screen = iota
	screenDashboard
	screenKanban
	screenCreateOrg
	screenCreateProject
	screenOrgSettings
	screenEditProject
	screenTaskDetail
	screenEditTask
	screenCreateTask
	screenInviteUser
	screenMCPSetup
	screenCreateComment
	screenAssignUser
	screenBoardSetup
	screenSearch
	screenSprintView
	screenNotificationSetup
	screenConnections
	screenCreateSprintTUI
	screenCreateSubtask
	screenSubtaskDetail
	screenEditSubtask
	screenCreateSubtaskComment
	screenEditSubtaskComment
	screenCEODashboard
	screenLogTime
	screenExportReport
	screenEditSprint
	screenGuide
	screenInvitationPrompt
	screenOrgSelector
	screenOrgMembers
	screenOrgDanger
)

type NavigateMsg struct {
	To             screen
	GothUser       *goth.User
	Project        domain.Project
	Org            domain.Organization
	Task           domain.Task
	StateID        uint
	Editing        bool
	Onboarding     bool
	Sprint         domain.Sprint
	Subtask        domain.Subtask
	SubtaskComment domain.SubtaskComment
}

type UserResolvedMsg struct {
	User        *domain.User
	OrgID       uint
	Org         *domain.Organization
	Invitations []domain.Invitation
	MemberOrgs  []domain.Organization
	Err         error
}

type GoBackMsg struct{}

type OrgContextClearedMsg struct{}

type AppModel struct {
	activeScreen      screen
	currentModel      tea.Model
	prevModel         tea.Model
	prevScreen        screen
	userSvc           *app.UserService
	orgSvc            *app.OrganizationService
	projectSvc        *app.ProjectService
	stateSvc          *app.StateService
	taskSvc           *app.TaskService
	teamSvc           *app.TeamService
	sprintSvc         *app.SprintService
	notifSvc          *app.NotificationService
	commentSvc        *app.CommentService
	invitationSvc     *app.InvitationService
	subtaskSvc        *app.SubtaskService
	subtaskCommentSvc *app.SubtaskCommentService
	timeSvc           *app.TimeEntryService
	dashboardSvc      *app.DashboardService
	reportSvc         *app.ReportService
	currentUser       *domain.User
	currentOrgID      uint
	currentOrg        *domain.Organization
	isOnboarding      bool
}

func (m AppModel) Init() tea.Cmd {
	return m.currentModel.Init()
}

func (m AppModel) resolveUserCmd(gothUser goth.User) tea.Cmd {
	return func() tea.Msg {
		user, _, err := m.userSvc.FindOrCreateByOAuth(
			gothUser.Provider,
			gothUser.UserID,
			gothUser.Email,
			gothUser.Name,
			&gothUser.AvatarURL,
		)
		if err != nil {
			return UserResolvedMsg{Err: fmt.Errorf("could not resolve user: %w", err)}
		}

		invitations, invErr := m.invitationSvc.GetPendingByEmail(user.Email)
		if invErr != nil {
			invitations = nil
		}

		memberOrgs, memberErr := m.teamSvc.GetOrganizationsByMemberUserID(user.ID)
		if memberErr != nil {
			return UserResolvedMsg{Err: fmt.Errorf("could not load your organizations: %w", memberErr)}
		}

		ownedOrg, ownedErr := m.orgSvc.GetByOwnerID(user.ID)

		if ownedErr == nil {
			allOrgs := append([]domain.Organization{*ownedOrg}, memberOrgs...)
			if len(memberOrgs) == 0 && len(invitations) == 0 {
				return UserResolvedMsg{User: user, OrgID: ownedOrg.ID, Org: ownedOrg}
			}
			return UserResolvedMsg{
				User:        user,
				OrgID:       0,
				Invitations: invitations,
				MemberOrgs:  allOrgs,
			}
		}

		return UserResolvedMsg{
			User:        user,
			OrgID:       0,
			Invitations: invitations,
			MemberOrgs:  memberOrgs,
		}
	}
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(GoBackMsg); ok {
		if m.prevModel != nil {
			m.currentModel = m.prevModel
			m.activeScreen = m.prevScreen
			m.prevModel = nil
		}
		return m, nil
	}

	switch msg := msg.(type) {

	case NavigateMsg:
		switch msg.To {
		case screenDashboard:
			if msg.GothUser != nil {
				return m, m.resolveUserCmd(*msg.GothUser)
			}
			dashboard := NewDashboardModel(m.currentOrgID, m.projectSvc)
			m.currentModel = dashboard
			m.activeScreen = screenDashboard
			return m, dashboard.Init()

		case screenKanban:
			kanban := NewKanbanModel(msg.Project, m.stateSvc, m.taskSvc)
			m.currentModel = kanban
			m.activeScreen = screenKanban
			return m, kanban.Init()

		case screenCreateOrg:
			createOrg := NewCreateOrgModel(m.currentUser.ID, m.orgSvc, m.teamSvc)
			m.currentModel = createOrg
			m.activeScreen = screenCreateOrg
			return m, createOrg.Init()

		case screenCreateProject:
			orgID := msg.Org.ID
			if orgID == 0 {
				orgID = m.currentOrgID
			}
			createProject := NewCreateProjectModel(
				orgID,
				m.currentUser.ID,
				m.projectSvc,
				m.stateSvc,
			)
			m.currentModel = createProject
			m.activeScreen = screenCreateProject
			return m, createProject.Init()

		case screenOrgSettings:
			if m.currentOrg != nil {
				settings := NewOrgSettingsModel(*m.currentOrg, m.orgSvc)
				m.currentModel = settings
				m.activeScreen = screenOrgSettings
				return m, settings.Init()
			}

		case screenEditProject:
			edit := NewEditProjectModel(msg.Project, m.projectSvc)
			m.currentModel = edit
			m.activeScreen = screenEditProject
			return m, edit.Init()

		case screenTaskDetail:
			userID := uint(0)
			if m.currentUser != nil {
				userID = m.currentUser.ID
			}
			detail := NewTaskDetailModel(msg.Task, msg.Project, m.taskSvc, m.commentSvc, m.subtaskSvc, m.timeSvc, m.stateSvc, userID)
			m.currentModel = detail
			m.activeScreen = screenTaskDetail
			return m, detail.Init()

		case screenEditTask:
			editTask := NewEditTaskModel(msg.Task, msg.Project, m.taskSvc)
			m.currentModel = editTask
			m.activeScreen = screenEditTask
			return m, editTask.Init()

		case screenBoardSetup:
			setup := NewBoardSetupModel(msg.Project, msg.Editing, m.isOnboarding, m.stateSvc, m.taskSvc)
			m.currentModel = setup
			m.activeScreen = screenBoardSetup
			return m, setup.Init()

		case screenAssignUser:
			if m.currentUser != nil {
				assignUser := NewAssignUserModel(
					msg.Task,
					msg.Project,
					m.currentOrgID,
					m.teamSvc,
					m.taskSvc,
				)
				m.currentModel = assignUser
				m.activeScreen = screenAssignUser
				return m, assignUser.Init()
			}

		case screenCreateComment:
			if m.currentUser != nil {
				createComment := NewCreateCommentModel(
					msg.Task,
					msg.Project,
					m.currentUser.ID,
					m.commentSvc,
				)
				m.currentModel = createComment
				m.activeScreen = screenCreateComment
				return m, createComment.Init()
			}

		case screenSprintView:
			sv := NewSprintViewModel(msg.Project, m.sprintSvc, m.taskSvc, m.stateSvc)
			m.currentModel = sv
			m.activeScreen = screenSprintView
			return m, sv.Init()

		case screenNotificationSetup:
			ns := NewNotificationSetupModel(m.currentOrgID, m.notifSvc)
			m.currentModel = ns
			m.activeScreen = screenNotificationSetup
			return m, ns.Init()

		case screenConnections:
			conn := NewConnectionsModel(m.currentOrgID, m.notifSvc)
			m.currentModel = conn
			m.activeScreen = screenConnections
			return m, conn.Init()

		case screenCreateSprintTUI:
			cs := NewCreateSprintTUIModel(msg.Project, m.sprintSvc)
			m.currentModel = cs
			m.activeScreen = screenCreateSprintTUI
			return m, cs.Init()

		case screenCreateSubtask:
			if m.currentUser != nil {
				createSubtask := NewCreateSubtaskModel(
					msg.Task,
					msg.Project,
					m.currentUser.ID,
					m.subtaskSvc,
				)
				m.currentModel = createSubtask
				m.activeScreen = screenCreateSubtask
				return m, createSubtask.Init()
			}

		case screenSubtaskDetail:
			userID := uint(0)
			if m.currentUser != nil {
				userID = m.currentUser.ID
			}
			subtaskDetail := NewSubtaskDetailModel(
				msg.Subtask,
				msg.Task,
				msg.Project,
				m.subtaskSvc,
				m.subtaskCommentSvc,
				m.timeSvc,
				userID,
			)
			m.currentModel = subtaskDetail
			m.activeScreen = screenSubtaskDetail
			return m, subtaskDetail.Init()

		case screenEditSubtask:
			editSubtask := NewEditSubtaskModel(
				msg.Subtask,
				msg.Task,
				msg.Project,
				m.subtaskSvc,
			)
			m.currentModel = editSubtask
			m.activeScreen = screenEditSubtask
			return m, editSubtask.Init()

		case screenCreateSubtaskComment:
			if m.currentUser != nil {
				createSubtaskComment := NewCreateSubtaskCommentModel(
					msg.Subtask,
					msg.Task,
					msg.Project,
					m.currentUser.ID,
					m.subtaskCommentSvc,
				)
				m.currentModel = createSubtaskComment
				m.activeScreen = screenCreateSubtaskComment
				return m, createSubtaskComment.Init()
			}

		case screenEditSubtaskComment:
			editSubtaskComment := NewEditSubtaskCommentModel(
				msg.SubtaskComment,
				msg.Subtask,
				msg.Task,
				msg.Project,
				m.subtaskCommentSvc,
			)
			m.currentModel = editSubtaskComment
			m.activeScreen = screenEditSubtaskComment
			return m, editSubtaskComment.Init()

		case screenCEODashboard:
			if msg.Org.ID != 0 {
				m.currentOrgID = msg.Org.ID
				org := msg.Org
				m.currentOrg = &org
			}
			ceo := NewCEODashboardModel(m.currentOrgID, m.dashboardSvc, m.projectSvc)
			m.currentModel = ceo
			m.activeScreen = screenCEODashboard
			return m, ceo.Init()

		case screenLogTime:
			if m.currentUser != nil {
				var taskPtr *domain.Task
				var subtaskPtr *domain.Subtask
				if msg.Task.ID != 0 {
					t := msg.Task
					taskPtr = &t
				}
				if msg.Subtask.ID != 0 {
					st := msg.Subtask
					subtaskPtr = &st
				}
				logTime := NewLogTimeModel(taskPtr, subtaskPtr, msg.Project, m.currentUser.ID, m.timeSvc)
				m.currentModel = logTime
				m.activeScreen = screenLogTime
				return m, logTime.Init()
			}

		case screenGuide:
			m.prevModel = m.currentModel
			m.prevScreen = m.activeScreen
			guide := NewGuideModel()
			m.currentModel = guide
			m.activeScreen = screenGuide
			return m, guide.Init()

		case screenEditSprint:
			editSprint := NewEditSprintModel(msg.Sprint, msg.Project, m.sprintSvc)
			m.currentModel = editSprint
			m.activeScreen = screenEditSprint
			return m, editSprint.Init()

		case screenExportReport:
			export := NewExportReportModel(m.currentOrgID, m.reportSvc, m.projectSvc)
			m.currentModel = export
			m.activeScreen = screenExportReport
			return m, export.Init()

		case screenSearch:
			search := NewSearchModel(m.currentOrgID, m.projectSvc, m.taskSvc)
			m.currentModel = search
			m.activeScreen = screenSearch
			return m, search.Init()

		case screenMCPSetup:
			m.isOnboarding = false
			setup := NewMCPSetupModel()
			m.currentModel = setup
			m.activeScreen = screenMCPSetup
			return m, setup.Init()

		case screenInviteUser:
			if m.currentOrg != nil {
				invite := NewInviteUserModel(*m.currentOrg, m.invitationSvc)
				m.currentModel = invite
				m.activeScreen = screenInviteUser
				return m, invite.Init()
			}

		case screenCreateTask:
			createTask := NewCreateTaskModel(
				msg.StateID,
				msg.Project.ID,
				m.currentUser.ID,
				msg.Project,
				m.taskSvc,
			)
			m.currentModel = createTask
			m.activeScreen = screenCreateTask
			return m, createTask.Init()

		case screenInvitationPrompt:
			if m.currentUser != nil {
				prompt := NewInvitationPromptModel(nil, nil, m.currentUser, m.invitationSvc, m.teamSvc)
				m.currentModel = prompt
				m.activeScreen = screenInvitationPrompt
				return m, prompt.Init()
			}

		case screenOrgSelector:
			if m.currentUser != nil {
				memberOrgs, _ := m.teamSvc.GetOrganizationsByMemberUserID(m.currentUser.ID)
				ownedOrg, ownedErr := m.orgSvc.GetByOwnerID(m.currentUser.ID)
				allOrgs := memberOrgs
				if ownedErr == nil {
					allOrgs = append([]domain.Organization{*ownedOrg}, memberOrgs...)
				}
				selector := NewOrgSelectorModel(allOrgs)
				m.currentModel = selector
				m.activeScreen = screenOrgSelector
				return m, selector.Init()
			}

		case screenOrgMembers:
			if m.currentOrg != nil && m.currentUser != nil {
				members := NewOrgMembersModel(*m.currentOrg, m.currentUser.ID, m.teamSvc)
				m.currentModel = members
				m.activeScreen = screenOrgMembers
				return m, members.Init()
			}

		case screenOrgDanger:
			if m.currentOrg != nil && m.currentUser != nil {
				isOwner := m.currentOrg.OwnerID == m.currentUser.ID
				danger := NewOrgDangerModel(*m.currentOrg, m.currentUser, isOwner, m.orgSvc, m.teamSvc)
				m.currentModel = danger
				m.activeScreen = screenOrgDanger
				return m, danger.Init()
			}

		case screenLogin:
			_ = auth.ClearSession()
			m.currentUser = nil
			m.currentOrgID = 0
			m.currentOrg = nil
			login := NewLoginModel()
			m.currentModel = login
			m.activeScreen = screenLogin
			return m, login.Init()
		}

	case UserResolvedMsg:
		if msg.Err != nil {
			return m, tea.Quit
		}
		m.currentUser = msg.User
		m.currentOrgID = msg.OrgID

		if len(msg.Invitations) > 0 {
			prompt := NewInvitationPromptModel(msg.Invitations, msg.MemberOrgs, m.currentUser, m.invitationSvc, m.teamSvc)
			m.currentModel = prompt
			m.activeScreen = screenInvitationPrompt
			return m, prompt.Init()
		}

		if len(msg.MemberOrgs) == 1 {
			org := msg.MemberOrgs[0]
			m.currentOrgID = org.ID
			m.currentOrg = &org
			if msg.OrgID != 0 && msg.Org != nil {
				m.currentOrg = msg.Org
			}
			ceo := NewCEODashboardModel(m.currentOrgID, m.dashboardSvc, m.projectSvc)
			m.currentModel = ceo
			m.activeScreen = screenCEODashboard
			return m, ceo.Init()
		}

		if len(msg.MemberOrgs) > 1 {
			selector := NewOrgSelectorModel(msg.MemberOrgs)
			m.currentModel = selector
			m.activeScreen = screenOrgSelector
			return m, selector.Init()
		}

		if msg.OrgID != 0 {
			if msg.Org != nil {
				m.currentOrg = msg.Org
			} else {
				org, err := m.orgSvc.GetByID(msg.OrgID)
				if err == nil {
					m.currentOrg = org
				}
			}
			ceo := NewCEODashboardModel(m.currentOrgID, m.dashboardSvc, m.projectSvc)
			m.currentModel = ceo
			m.activeScreen = screenCEODashboard
			return m, ceo.Init()
		}

		m.isOnboarding = true
		createOrg := NewCreateOrgModel(m.currentUser.ID, m.orgSvc, m.teamSvc)
		m.currentModel = createOrg
		m.activeScreen = screenCreateOrg
		return m, createOrg.Init()

	case OrgContextClearedMsg:
		m.currentOrgID = 0
		m.currentOrg = nil
		if m.currentUser != nil {
			memberOrgs, memberErr := m.teamSvc.GetOrganizationsByMemberUserID(m.currentUser.ID)
			if memberErr == nil && len(memberOrgs) == 1 {
				org := memberOrgs[0]
				m.currentOrgID = org.ID
				m.currentOrg = &org
				ceo := NewCEODashboardModel(m.currentOrgID, m.dashboardSvc, m.projectSvc)
				m.currentModel = ceo
				m.activeScreen = screenCEODashboard
				return m, ceo.Init()
			}
			if memberErr == nil && len(memberOrgs) > 1 {
				selector := NewOrgSelectorModel(memberOrgs)
				m.currentModel = selector
				m.activeScreen = screenOrgSelector
				return m, selector.Init()
			}
		}
		m.isOnboarding = true
		createOrg := NewCreateOrgModel(m.currentUser.ID, m.orgSvc, m.teamSvc)
		m.currentModel = createOrg
		m.activeScreen = screenCreateOrg
		return m, createOrg.Init()

	}

	updated, cmd := m.currentModel.Update(msg)
	m.currentModel = updated
	return m, cmd
}

func (m AppModel) View() string {
	return m.currentModel.View()
}

func Start(db *gorm.DB) error {
	auth.SetupProviders()

	userSvc := app.NewUserService(db)
	orgSvc := app.NewOrganizationService(db)
	projectSvc := app.NewProjectService(db)
	stateSvc := app.NewStateService(db)
	taskSvc := app.NewTaskService(db)
	teamSvc := app.NewTeamService(db)
	commentSvc := app.NewCommentService(db)
	sprintSvc := app.NewSprintService(db)
	notifSvc := app.NewNotificationService(db)
	invitationSvc := app.NewInvitationService(db)
	subtaskSvc := app.NewSubtaskService(db)
	subtaskCommentSvc := app.NewSubtaskCommentService(db)
	timeSvc := app.NewTimeEntryService(db)
	dashboardSvc := app.NewDashboardService(db)
	reportSvc := app.NewReportService(db)

	startModel := tea.Model(NewLoginModel())
	startOrgID := uint(0)
	var startUser *domain.User
	var startOrg *domain.Organization

	session, err := auth.LoadSession()
	if err == nil {
		user, dbErr := userSvc.GetByEmail(session.Email)
		if dbErr == nil {
			startUser = user
			invitations, _ := invitationSvc.GetPendingByEmail(user.Email)
			memberOrgs, _ := teamSvc.GetOrganizationsByMemberUserID(user.ID)
			ownedOrg, ownedErr := orgSvc.GetByOwnerID(user.ID)

			if ownedErr == nil {
				allOrgs := append([]domain.Organization{*ownedOrg}, memberOrgs...)
				if len(memberOrgs) == 0 && len(invitations) == 0 {
					startOrgID = ownedOrg.ID
					startOrg = ownedOrg
					startModel = NewCEODashboardModel(startOrgID, dashboardSvc, projectSvc)
				} else if len(invitations) > 0 {
					startModel = NewInvitationPromptModel(invitations, allOrgs, user, invitationSvc, teamSvc)
				} else {
					startModel = NewOrgSelectorModel(allOrgs)
				}
			} else if len(invitations) > 0 {
				startModel = NewInvitationPromptModel(invitations, memberOrgs, user, invitationSvc, teamSvc)
			} else if len(memberOrgs) == 1 {
				startOrgID = memberOrgs[0].ID
				startOrg = &memberOrgs[0]
				startModel = NewCEODashboardModel(startOrgID, dashboardSvc, projectSvc)
			} else if len(memberOrgs) > 1 {
				startModel = NewOrgSelectorModel(memberOrgs)
			} else {
				startModel = NewCreateOrgModel(user.ID, orgSvc, teamSvc)
			}
		}
	}

	model := AppModel{
		activeScreen:      screenLogin,
		currentModel:      startModel,
		userSvc:           userSvc,
		orgSvc:            orgSvc,
		projectSvc:        projectSvc,
		stateSvc:          stateSvc,
		taskSvc:           taskSvc,
		teamSvc:           teamSvc,
		commentSvc:        commentSvc,
		invitationSvc:     invitationSvc,
		sprintSvc:         sprintSvc,
		notifSvc:          notifSvc,
		subtaskSvc:        subtaskSvc,
		subtaskCommentSvc: subtaskCommentSvc,
		timeSvc:           timeSvc,
		dashboardSvc:      dashboardSvc,
		reportSvc:         reportSvc,
		currentUser:       startUser,
		currentOrgID:      startOrgID,
		currentOrg:        startOrg,
	}

	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running app: %w", err)
	}

	return nil
}
