package api

import (
	"encoding/json"
	"net/http"

	"github.com/varmiguemunoz/sprintos/internal/app"
	"gorm.io/gorm"
)

type Server struct {
	db            *gorm.DB
	internalToken string
	taskSvc       *app.TaskService
	projectSvc    *app.ProjectService
	stateSvc      *app.StateService
	teamSvc       *app.TeamService
	orgSvc        *app.OrganizationService
	userSvc       *app.UserService
	apiKeySvc     *app.APIKeyService
	outboundSvc   *app.OutboundWebhookService
	timeSvc       *app.TimeEntryService
}

func NewServer(db *gorm.DB, internalToken string) *Server {
	return &Server{
		db:            db,
		internalToken: internalToken,
		taskSvc:       app.NewTaskService(db),
		projectSvc:    app.NewProjectService(db),
		stateSvc:      app.NewStateService(db),
		teamSvc:       app.NewTeamService(db),
		orgSvc:        app.NewOrganizationService(db),
		userSvc:       app.NewUserService(db),
		apiKeySvc:     app.NewAPIKeyService(db),
		outboundSvc:   app.NewOutboundWebhookService(db),
		timeSvc:       app.NewTimeEntryService(db),
	}
}

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/docs", s.handleDocs)
	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("GET /api/projects", s.auth(s.rateLimit(s.listProjects)))
	mux.HandleFunc("POST /api/projects", s.auth(s.rateLimit(s.createProject)))
	mux.HandleFunc("GET /api/tasks", s.auth(s.rateLimit(s.listTasks)))
	mux.HandleFunc("POST /api/tasks", s.auth(s.rateLimit(s.createTask)))
	mux.HandleFunc("GET /api/tasks/{id}", s.auth(s.rateLimit(s.getTask)))
	mux.HandleFunc("PATCH /api/tasks/{id}", s.auth(s.rateLimit(s.updateTask)))
	mux.HandleFunc("DELETE /api/tasks/{id}", s.auth(s.rateLimit(s.deleteTask)))
	mux.HandleFunc("POST /api/tasks/{id}/move", s.auth(s.rateLimit(s.moveTask)))
	mux.HandleFunc("GET /api/states", s.auth(s.rateLimit(s.listStates)))
	mux.HandleFunc("GET /api/members", s.auth(s.rateLimit(s.listMembers)))
	mux.HandleFunc("GET /api/webhooks", s.auth(s.rateLimit(s.listWebhooks)))
	mux.HandleFunc("POST /api/webhooks", s.auth(s.rateLimit(s.createWebhook)))
	mux.HandleFunc("DELETE /api/webhooks/{id}", s.auth(s.rateLimit(s.deleteWebhook)))

	mux.HandleFunc("GET /api/tray/tasks", s.trayAuth(s.trayListAllTasks))
	mux.HandleFunc("GET /api/tray/projects", s.trayAuth(s.trayListProjects))
	mux.HandleFunc("GET /api/tray/projects/{id}/tasks", s.trayAuth(s.trayListTasks))
	mux.HandleFunc("POST /api/tray/timer/start", s.trayAuth(s.trayStartTimer))
	mux.HandleFunc("POST /api/tray/timer/stop", s.trayAuth(s.trayStopTimer))
	mux.HandleFunc("GET /api/tray/timer/active", s.trayAuth(s.trayGetActiveTimer))
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleDocs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	http.ServeFile(w, r, "docs/openapi.yaml")
}

func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func Error(w http.ResponseWriter, status int, msg string) {
	JSON(w, status, map[string]string{"error": msg})
}
