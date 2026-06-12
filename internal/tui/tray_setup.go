package tui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/varmiguemunoz/sprintos/internal/tray"
)

type TraySetupModel struct {
	loading bool
	status  string
	err     error
}

type trayActionDoneMsg struct {
	err error
}

func NewTraySetupModel() TraySetupModel {
	return TraySetupModel{}
}

func (m TraySetupModel) Init() tea.Cmd {
	return nil
}

func (m TraySetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case trayActionDoneMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			m.status = ""
		}
		return m, nil

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}
		switch msg.String() {
		case "i":
			if runtime.GOOS != "darwin" {
				m.err = fmt.Errorf("menu bar app is only supported on macOS")
				return m, nil
			}
			m.loading = true
			m.err = nil
			m.status = ""
			return m, func() tea.Msg {
				return trayActionDoneMsg{err: tray.Install()}
			}

		case "u":
			if runtime.GOOS != "darwin" {
				m.err = fmt.Errorf("menu bar app is only supported on macOS")
				return m, nil
			}
			m.loading = true
			m.err = nil
			m.status = ""
			return m, func() tea.Msg {
				err := tray.Uninstall()
				if err != nil {
					return trayActionDoneMsg{err: err}
				}
				return trayActionDoneMsg{}
			}

		case "r":
			if runtime.GOOS != "darwin" {
				m.err = fmt.Errorf("menu bar app is only supported on macOS")
				return m, nil
			}
			m.err = nil
			binary, err := os.Executable()
			if err != nil {
				m.err = fmt.Errorf("could not find binary: %w", err)
				return m, nil
			}
			cmd := exec.Command(binary, "tray")
			if err := cmd.Start(); err != nil {
				m.err = fmt.Errorf("could not launch tray: %w", err)
				return m, nil
			}
			m.status = "Tray app launched in background"
			return m, nil

		case "esc":
			return m, func() tea.Msg {
				return NavigateMsg{To: screenOrgSettings}
			}
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m TraySetupModel) View() string {
	s := titleStyle.Render("SprintOS — Menu Bar App") + "\n\n"

	if runtime.GOOS != "darwin" {
		s += errorStyle.Render("⚠  Menu bar app is only available on macOS.") + "\n\n"
		s += normalStyle.Render("Your current OS: "+runtime.GOOS) + "\n\n"
		s += renderHintBar("esc", "back") + "\n"
		return s
	}

	installed := tray.IsInstalled()

	if installed {
		s += selectedStyle.Render("Status: Installed ✓") + "\n"
		s += normalStyle.Render("Launches automatically when you log in.") + "\n\n"
	} else {
		s += normalStyle.Render("Status: Not installed") + "\n"
		s += normalStyle.Render("Install it to launch automatically on login.") + "\n\n"
	}

	if m.loading {
		s += normalStyle.Render("Working...") + "\n"
		return s
	}

	if m.err != nil {
		s += errorStyle.Render("Error: "+m.err.Error()) + "\n\n"
	}

	if m.status != "" {
		s += selectedStyle.Render("✓ "+m.status) + "\n\n"
	}

	if installed {
		s += renderHintBar("r", "run now", "u", "uninstall", "esc", "back") + "\n"
	} else {
		s += renderHintBar("i", "install", "r", "run now", "esc", "back") + "\n"
	}

	return s
}
