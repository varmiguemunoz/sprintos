package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

func (s *Server) trayListAllTasks(w http.ResponseWriter, r *http.Request) {
	key := currentKey(r)

	projects, err := s.projectSvc.ListByOrganization(key.OrgID)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	type taskOut struct {
		ID    uint   `json:"id"`
		Title string `json:"title"`
	}

	all := []taskOut{}
	for _, p := range projects {
		tasks, err := s.taskSvc.ListByProject(p.ID)
		if err != nil {
			continue
		}
		for _, t := range tasks {
			if !t.State.IsDone {
				all = append(all, taskOut{ID: t.ID, Title: t.Title})
			}
		}
	}

	JSON(w, http.StatusOK, all)
}

func (s *Server) trayListProjects(w http.ResponseWriter, r *http.Request) {
	key := currentKey(r)
	projects, err := s.projectSvc.ListByOrganization(key.OrgID)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	type projectOut struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}
	out := make([]projectOut, len(projects))
	for i, p := range projects {
		out[i] = projectOut{ID: p.ID, Name: p.Name}
	}
	JSON(w, http.StatusOK, out)
}

func (s *Server) trayListTasks(w http.ResponseWriter, r *http.Request) {
	projectID, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		Error(w, http.StatusBadRequest, "invalid project id")
		return
	}

	tasks, err := s.taskSvc.ListByProject(uint(projectID))
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	type taskOut struct {
		ID        uint   `json:"id"`
		Title     string `json:"title"`
		ProjectID uint   `json:"project_id"`
	}
	out := make([]taskOut, len(tasks))
	for i, t := range tasks {
		out[i] = taskOut{ID: t.ID, Title: t.Title, ProjectID: t.ProjectID}
	}
	JSON(w, http.StatusOK, out)
}

func (s *Server) trayStartTimer(w http.ResponseWriter, r *http.Request) {
	key := currentKey(r)

	var body struct {
		TaskID uint `json:"task_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.TaskID == 0 {
		Error(w, http.StatusBadRequest, "task_id is required")
		return
	}

	_, err := s.timeSvc.StartTimer(&body.TaskID, nil, key.UserID)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	JSON(w, http.StatusOK, map[string]string{"status": "started"})
}

func (s *Server) trayStopTimer(w http.ResponseWriter, r *http.Request) {
	key := currentKey(r)

	entry, err := s.timeSvc.StopTimer(key.UserID)
	if err != nil {
		Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if entry == nil {
		JSON(w, http.StatusOK, map[string]string{"status": "no active timer"})
		return
	}
	JSON(w, http.StatusOK, map[string]interface{}{"status": "stopped", "minutes": entry.Minutes})
}

func (s *Server) trayGetActiveTimer(w http.ResponseWriter, r *http.Request) {
	key := currentKey(r)

	timer, err := s.timeSvc.GetActiveTimer(key.UserID)
	if err != nil || timer == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	taskTitle := ""
	if timer.TaskID != nil {
		if task, err := s.taskSvc.GetByID(*timer.TaskID); err == nil {
			taskTitle = task.Title
		}
	}

	elapsed := int(time.Since(timer.StartedAt).Minutes())

	type out struct {
		Running        bool      `json:"running"`
		TaskID         uint      `json:"task_id"`
		TaskTitle      string    `json:"task_title"`
		StartedAt      time.Time `json:"started_at"`
		ElapsedMinutes int       `json:"elapsed_minutes"`
	}

	taskID := uint(0)
	if timer.TaskID != nil {
		taskID = *timer.TaskID
	}

	JSON(w, http.StatusOK, out{
		Running:        true,
		TaskID:         taskID,
		TaskTitle:      taskTitle,
		StartedAt:      timer.StartedAt,
		ElapsedMinutes: elapsed,
	})
}
