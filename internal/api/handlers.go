package api

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func (s *Server) listProjects(w http.ResponseWriter, r *http.Request) {
	key := currentKey(r)
	projects, err := s.projectSvc.ListByOrganization(key.OrgID)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, projects)
}

func (s *Server) createProject(w http.ResponseWriter, r *http.Request) {
	key := currentKey(r)
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		Error(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if body.Name == "" {
		Error(w, http.StatusBadRequest, "name is required")
		return
	}
	project, err := s.projectSvc.Create(body.Name, body.Description, key.OrgID, key.UserID)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusCreated, project)
}

func (s *Server) listTasks(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	projectID, _ := strconv.ParseUint(q.Get("project_id"), 10, 64)
	if projectID == 0 {
		Error(w, http.StatusBadRequest, "project_id is required")
		return
	}

	stateID, _ := strconv.ParseUint(q.Get("state_id"), 10, 64)
	assigneeID, _ := strconv.ParseUint(q.Get("assignee_id"), 10, 64)

	tasks, err := s.taskSvc.ListByProject(uint(projectID))
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	filtered := tasks[:0]
	for _, t := range tasks {
		if stateID > 0 && t.StateID != uint(stateID) {
			continue
		}
		if assigneeID > 0 && (t.AssignedTo == nil || *t.AssignedTo != uint(assigneeID)) {
			continue
		}
		filtered = append(filtered, t)
	}

	JSON(w, http.StatusOK, filtered)
}

func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
	key := currentKey(r)
	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		ProjectID   uint   `json:"project_id"`
		StateID     uint   `json:"state_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		Error(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if body.Title == "" || body.ProjectID == 0 || body.StateID == 0 {
		Error(w, http.StatusBadRequest, "title, project_id and state_id are required")
		return
	}
	task, err := s.taskSvc.Create(body.Title, body.Description, body.StateID, body.ProjectID, key.UserID, nil, nil, nil)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.outboundSvc.Fire(key.OrgID, "task.created", task)
	JSON(w, http.StatusCreated, task)
}

func (s *Server) getTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid task ID")
		return
	}
	task, err := s.taskSvc.GetByID(uint(id))
	if err != nil {
		Error(w, http.StatusNotFound, err.Error())
		return
	}
	JSON(w, http.StatusOK, task)
}

func (s *Server) updateTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid task ID")
		return
	}
	var body struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		Error(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	task, err := s.taskSvc.Update(uint(id), body.Title, body.Description, nil, nil, nil)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, task)
}

func (s *Server) deleteTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid task ID")
		return
	}
	if err := s.taskSvc.Delete(uint(id)); err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) moveTask(w http.ResponseWriter, r *http.Request) {
	key := currentKey(r)
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid task ID")
		return
	}
	var body struct {
		StateID uint `json:"state_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		Error(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if body.StateID == 0 {
		Error(w, http.StatusBadRequest, "state_id is required")
		return
	}
	task, err := s.taskSvc.MoveState(uint(id), body.StateID)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.outboundSvc.Fire(key.OrgID, "task.moved", task)
	JSON(w, http.StatusOK, task)
}

func (s *Server) listStates(w http.ResponseWriter, r *http.Request) {
	projectID, _ := strconv.ParseUint(r.URL.Query().Get("project_id"), 10, 64)
	if projectID == 0 {
		Error(w, http.StatusBadRequest, "project_id is required")
		return
	}
	states, err := s.stateSvc.ListByProject(uint(projectID))
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, states)
}

func (s *Server) listMembers(w http.ResponseWriter, r *http.Request) {
	key := currentKey(r)
	members, err := s.teamSvc.ListMembers(key.OrgID)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, members)
}

func (s *Server) listWebhooks(w http.ResponseWriter, r *http.Request) {
	key := currentKey(r)
	hooks, err := s.outboundSvc.ListByOrg(key.OrgID)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusOK, hooks)
}

func (s *Server) createWebhook(w http.ResponseWriter, r *http.Request) {
	key := currentKey(r)
	var body struct {
		URL    string   `json:"url"`
		Events []string `json:"events"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		Error(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if body.URL == "" {
		Error(w, http.StatusBadRequest, "url is required")
		return
	}
	if len(body.Events) == 0 {
		body.Events = []string{"task.created", "task.moved"}
	}
	hook, err := s.outboundSvc.Create(key.OrgID, body.URL, body.Events)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSON(w, http.StatusCreated, hook)
}

func (s *Server) deleteWebhook(w http.ResponseWriter, r *http.Request) {
	key := currentKey(r)
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid webhook ID")
		return
	}
	if err := s.outboundSvc.Delete(uint(id), key.OrgID); err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
