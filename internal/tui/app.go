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
)

type NavigateMsg struct {
	To         screen
	GothUser   *goth.User
	Project    domain.Project
	Org        domain.Organization
	Task       domain.Task
	StateID    uint
	Editing    bool
	Onboarding bool
	Sprint     domain.Sprint
}

type UserResolvedMsg struct {
	User  *domain.User
	OrgID uint
	Err   error
}

type AppModel struct {
	activeScreen screen
	currentModel tea.Model
	userSvc      *app.UserService
	orgSvc       *app.OrganizationService
	projectSvc   *app.ProjectService
	stateSvc     *app.StateService
	taskSvc      *app.TaskService
	teamSvc      *app.TeamService
	sprintSvc    *app.SprintService
	notifSvc     *app.NotificationService
	commentSvc    *app.CommentService
	invitationSvc *app.InvitationService
	currentUser   *domain.User
	currentOrgID uint
	currentOrg   *domain.Organization
	isOnboarding bool
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

		org, err := m.orgSvc.GetByOwnerID(user.ID)
		if err != nil {
			return UserResolvedMsg{User: user, OrgID: 0}
		}

		return UserResolvedMsg{User: user, OrgID: org.ID}
	}
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			detail := NewTaskDetailModel(msg.Task, msg.Project, m.taskSvc, m.commentSvc)
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

		if msg.OrgID == 0 {
			m.isOnboarding = true
			createOrg := NewCreateOrgModel(m.currentUser.ID, m.orgSvc, m.teamSvc)
			m.currentModel = createOrg
			m.activeScreen = screenCreateOrg
			return m, createOrg.Init()
		}

		org, err := m.orgSvc.GetByID(msg.OrgID)
		if err == nil {
			m.currentOrg = org
		}

		dashboard := NewDashboardModel(m.currentOrgID, m.projectSvc)
		m.currentModel = dashboard
		m.activeScreen = screenDashboard
		return m, dashboard.Init()
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

	startModel := tea.Model(NewLoginModel())
	startOrgID := uint(0)
	var startUser *domain.User
	var startOrg *domain.Organization

	session, err := auth.LoadSession()
	if err == nil {
		user, dbErr := userSvc.GetByEmail(session.Email)
		if dbErr == nil {
			startUser = user
			org, orgErr := orgSvc.GetByOwnerID(user.ID)
			if orgErr == nil {
				startOrgID = org.ID
				startOrg = org
				startModel = NewDashboardModel(startOrgID, projectSvc)
			} else {
				startModel = NewCreateOrgModel(user.ID, orgSvc, teamSvc)
			}
		}
	}

	model := AppModel{
		activeScreen: screenLogin,
		currentModel: startModel,
		userSvc:      userSvc,
		orgSvc:       orgSvc,
		projectSvc:   projectSvc,
		stateSvc:     stateSvc,
		taskSvc:      taskSvc,
		teamSvc:      teamSvc,
		commentSvc:    commentSvc,
		invitationSvc: invitationSvc,
		sprintSvc:     sprintSvc,
		notifSvc:      notifSvc,
		currentUser:   startUser,
		currentOrgID: startOrgID,
		currentOrg:   startOrg,
	}

	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running app: %w", err)
	}

	return nil
}
