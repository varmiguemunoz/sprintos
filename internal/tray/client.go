package tray

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const trayTokenHeader = "X-Tray-Token"

type ProjectSummary struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type TaskSummary struct {
	ID        uint   `json:"id"`
	Title     string `json:"title"`
	ProjectID uint   `json:"project_id"`
}

type TimerStatus struct {
	Running        bool      `json:"running"`
	TaskID         uint      `json:"task_id"`
	TaskTitle      string    `json:"task_title"`
	StartedAt      time.Time `json:"started_at"`
	ElapsedMinutes int       `json:"elapsed_minutes"`
}

type Client struct {
	baseURL       string
	internalToken string
	http          *http.Client
}

func NewClient(baseURL, internalToken string) *Client {
	return &Client{
		baseURL:       baseURL,
		internalToken: internalToken,
		http:          &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *Client) IsReady() bool {
	resp, err := c.http.Get(c.baseURL + "/api/health")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (c *Client) do(method, path string, body interface{}) (*http.Response, error) {
	var reqBody *strings.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = strings.NewReader(string(data))
	} else {
		reqBody = strings.NewReader("")
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set(trayTokenHeader, c.internalToken)
	req.Header.Set("Content-Type", "application/json")

	return c.http.Do(req)
}

func (c *Client) ListAllTasks() ([]TaskSummary, error) {
	resp, err := c.do("GET", "/api/tray/tasks", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned %d", resp.StatusCode)
	}

	var tasks []TaskSummary
	if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (c *Client) ListProjects() ([]ProjectSummary, error) {
	resp, err := c.do("GET", "/api/tray/projects", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned %d", resp.StatusCode)
	}

	var projects []ProjectSummary
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, err
	}
	return projects, nil
}

func (c *Client) ListTasks(projectID uint) ([]TaskSummary, error) {
	resp, err := c.do("GET", fmt.Sprintf("/api/tray/projects/%d/tasks", projectID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned %d", resp.StatusCode)
	}

	var tasks []TaskSummary
	if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (c *Client) StartTimer(taskID uint) error {
	resp, err := c.do("POST", "/api/tray/timer/start", map[string]uint{"task_id": taskID})
	if err != nil {
		return err
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) StopTimer() error {
	resp, err := c.do("POST", "/api/tray/timer/stop", nil)
	if err != nil {
		return err
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) GetActiveTimer() (*TimerStatus, error) {
	resp, err := c.do("GET", "/api/tray/timer/active", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return &TimerStatus{Running: false}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned %d", resp.StatusCode)
	}

	var status TimerStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, err
	}
	return &status, nil
}

func clampStr(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}
