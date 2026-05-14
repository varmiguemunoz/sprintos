package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/domain"
)

type searchResult struct {
	task    domain.Task
	project domain.Project
}

type SearchModel struct {
	input      textinput.Model
	results    []searchResult
	allTasks   []searchResult
	cursor     int
	err        error
	projectSvc *app.ProjectService
	taskSvc    *app.TaskService
	orgID      uint
	loaded     bool
}

type SearchDataLoadedMsg struct {
	Results []searchResult
	Err     error
}

func NewSearchModel(orgID uint, projectSvc *app.ProjectService, taskSvc *app.TaskService) SearchModel {
	input := textinput.New()
	input.Placeholder = "Type to search tasks..."
	input.CharLimit = 100
	input.Focus()
	return SearchModel{
		input:      input,
		orgID:      orgID,
		projectSvc: projectSvc,
		taskSvc:    taskSvc,
	}
}

func (m SearchModel) loadAllCmd() tea.Cmd {
	return func() tea.Msg {
		projects, err := m.projectSvc.ListByOrganization(m.orgID)
		if err != nil {
			return SearchDataLoadedMsg{Err: err}
		}
		var all []searchResult
		for _, p := range projects {
			tasks, _ := m.taskSvc.ListByProject(p.ID)
			for _, t := range tasks {
				all = append(all, searchResult{task: t, project: p})
			}
		}
		return SearchDataLoadedMsg{Results: all}
	}
}

func (m SearchModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.loadAllCmd())
}

func (m SearchModel) filter(query string) []searchResult {
	if query == "" {
		return m.allTasks
	}
	q := strings.ToLower(query)
	var out []searchResult
	for _, r := range m.allTasks {
		if strings.Contains(strings.ToLower(r.task.Title), q) ||
			strings.Contains(strings.ToLower(r.project.Name), q) {
			out = append(out, r)
		}
	}
	return out
}

func (m SearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(SearchDataLoadedMsg); ok {
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.allTasks = msg.Results
		m.results = msg.Results
		m.loaded = true
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, func() tea.Msg {
				return NavigateMsg{To: screenDashboard}
			}
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.results)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.results) > 0 {
				r := m.results[m.cursor]
				task := r.task
				project := r.project
				return m, func() tea.Msg {
					return NavigateMsg{To: screenTaskDetail, Task: task, Project: project}
				}
			}
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.results = m.filter(m.input.Value())
	m.cursor = 0
	return m, cmd
}

func (m SearchModel) View() string {
	s := titleStyle.Render("SprintOS — Search") + "\n\n"
	s += m.input.View() + "\n\n"

	if !m.loaded {
		s += normalStyle.Render("Loading tasks...") + "\n"
		return s
	}

	if len(m.results) == 0 {
		s += normalStyle.Render("No results.") + "\n"
	} else {
		shown := m.results
		if len(shown) > 12 {
			shown = shown[:12]
		}
		for i, r := range shown {
			ref := fmt.Sprintf("#%d", r.task.TaskNumber)
			line := fmt.Sprintf("[%s] %s  %s", r.project.Name, ref, r.task.Title)
			if i == m.cursor {
				s += selectedStyle.Render("> "+line) + "\n"
			} else {
				s += normalStyle.Render("  "+line) + "\n"
			}
		}
		if len(m.results) > 12 {
			s += normalStyle.Render(fmt.Sprintf("  ... and %d more", len(m.results)-12)) + "\n"
		}
	}

	s += "\n" + normalStyle.Render("↑/↓ move  •  enter open task  •  esc back") + "\n"
	return s
}
