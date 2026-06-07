package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/varmiguemunoz/sprintos/internal/app"
)

var taskIDPattern = regexp.MustCompile(`\b([A-Z]+-\d+)\b`)

type GitHubHandler struct {
	githubSvc *app.GitHubService
	taskSvc   *app.TaskService
	orgSvc    *app.OrganizationService
}

func NewGitHubHandler(githubSvc *app.GitHubService, taskSvc *app.TaskService, orgSvc *app.OrganizationService) *GitHubHandler {
	return &GitHubHandler{
		githubSvc: githubSvc,
		taskSvc:   taskSvc,
		orgSvc:    orgSvc,
	}
}

type prEvent struct {
	Action      string `json:"action"`
	PullRequest struct {
		Title string `json:"title"`
		Head  struct {
			Ref string `json:"ref"`
		} `json:"head"`
		Merged bool `json:"merged"`
	} `json:"pull_request"`
	Repository struct {
		Name  string `json:"name"`
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
	} `json:"repository"`
}

func (h *GitHubHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "could not read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	eventType := r.Header.Get("X-GitHub-Event")
	if eventType != "pull_request" {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ignored")
		return
	}

	var event prEvent
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	repoOwner := event.Repository.Owner.Login
	repoName := event.Repository.Name

	integration, err := h.githubSvc.GetIntegrationByRepo(repoOwner, repoName)
	if err != nil {
		http.Error(w, "no integration configured for this repo", http.StatusNotFound)
		return
	}

	sig := r.Header.Get("X-Hub-Signature-256")
	if !verifySignature(body, integration.WebhookSecret, sig) {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	searchText := event.PullRequest.Title + " " + event.PullRequest.Head.Ref
	taskIDs := taskIDPattern.FindAllString(strings.ToUpper(searchText), -1)

	if len(taskIDs) == 0 {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "no task IDs found")
		return
	}

	org, err := h.orgSvc.GetByID(integration.OrganizationID)
	if err != nil {
		http.Error(w, "org not found", http.StatusInternalServerError)
		return
	}

	for _, taskRef := range taskIDs {
		parts := strings.SplitN(taskRef, "-", 2)
		if len(parts) != 2 {
			continue
		}
		prefix := parts[0]
		if prefix != strings.ToUpper(org.Prefix) {
			continue
		}

		var taskNumber int
		fmt.Sscanf(parts[1], "%d", &taskNumber)

		task, err := h.taskSvc.GetByTaskNumber(taskNumber, integration.ProjectID)
		if err != nil {
			continue
		}

		var targetStateID uint
		switch {
		case event.Action == "opened" || event.Action == "reopened":
			targetStateID = integration.InReviewStateID
		case event.Action == "closed" && event.PullRequest.Merged:
			targetStateID = integration.DoneStateID
		default:
			continue
		}

		_, _ = h.taskSvc.MoveState(task.ID, targetStateID)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "ok")
}

func verifySignature(body []byte, secret, signature string) bool {
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}
