package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/sprintos/internal/infrastructure/mcpdetector"
)

type MCPSetupModel struct {
	tools    []mcpdetector.AITool
	cursor   int
	selected map[int]bool
	loading  bool
	done     bool
	err      error
	results  []string
}

type MCPInstallDoneMsg struct {
	Results []string
	Err     error
}

func NewMCPSetupModel() MCPSetupModel {
	tools := mcpdetector.DetectTools()
	selected := make(map[int]bool)
	for i, t := range tools {
		if t.Detected && !t.Configured {
			selected[i] = true
		}
	}
	return MCPSetupModel{
		tools:    tools,
		selected: selected,
	}
}

func (m MCPSetupModel) installCmd() tea.Cmd {
	return func() tea.Msg {
		binaryPath, err := os.Executable()
		if err != nil {
			return MCPInstallDoneMsg{Err: fmt.Errorf("could not find binary path: %w", err)}
		}

		var results []string
		for i, tool := range m.tools {
			if !m.selected[i] {
				continue
			}
			t := tool
			if err := mcpdetector.InstallMCP(&t, binaryPath); err != nil {
				results = append(results, fmt.Sprintf("✗ %s: %s", t.Name, err.Error()))
			} else {
				results = append(results, fmt.Sprintf("✓ %s: configured", t.Name))
			}
		}

		return MCPInstallDoneMsg{Results: results}
	}
}

func (m MCPSetupModel) Init() tea.Cmd {
	return nil
}

func (m MCPSetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(MCPInstallDoneMsg); ok {
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.done = true
		m.results = msg.Results
		return m, nil
	}

	if m.loading {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg {
				return NavigateMsg{To: screenOrgSettings}
			}
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.tools)-1 {
				m.cursor++
			}
		case " ":
			if m.tools[m.cursor].Detected {
				m.selected[m.cursor] = !m.selected[m.cursor]
			}
		case "enter":
			if !m.done {
				m.loading = true
				return m, m.installCmd()
			}
		}
	}

	return m, nil
}

func (m MCPSetupModel) View() string {
	if m.loading {
		return titleStyle.Render("SprintOS — MCP Setup") +
			"\n\n" + normalStyle.Render("Installing MCP on selected tools...") + "\n"
	}

	s := titleStyle.Render("SprintOS — MCP Setup") + "\n\n"

	if m.done {
		s += selectedStyle.Render("Installation complete:") + "\n\n"
		for _, r := range m.results {
			s += normalStyle.Render("  "+r) + "\n"
		}
		s += "\n" + normalStyle.Render("Restart your AI tools to activate the MCP server.") + "\n\n"
		s += normalStyle.Render("esc back  •  q quit") + "\n"
		return s
	}

	s += normalStyle.Render("Detected AI tools on your system:") + "\n\n"

	for i, tool := range m.tools {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}

		var status string
		if !tool.Detected {
			status = normalStyle.Render("✗ " + tool.Name + " (not found)")
		} else if tool.Configured {
			status = selectedStyle.Render("✓ " + tool.Name + " (already configured)")
		} else if m.selected[i] {
			status = selectedStyle.Render("☑ " + tool.Name + " (will install)")
		} else {
			status = normalStyle.Render("☐ " + tool.Name + " (found, not selected)")
		}

		s += normalStyle.Render(cursor) + status + "\n"
	}

	if m.err != nil {
		s += "\n" + errorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())) + "\n"
	}

	s += "\n" + normalStyle.Render("↑/↓ move  •  space select  •  enter install  •  esc back") + "\n"
	return s
}
