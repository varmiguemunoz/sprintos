package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/markbates/goth"
	"github.com/varmiguemunoz/sprintos/internal/domain"
	"github.com/varmiguemunoz/sprintos/internal/infrastructure/auth"
)

type LoginModel struct {
	cursor  int
	choices []string
	loading bool
	err     error
}

type LoginResultMsg struct {
	User goth.User
	Err  error
}

func NewLoginModel() LoginModel {
	return LoginModel{
		choices: []string{"Login with GitHub", "Exit"},
	}
}

func startLoginCmd(provider string) tea.Cmd {
	return func() tea.Msg {
		user, err := auth.StartLogin(provider)
		return LoginResultMsg{User: user, Err: err}
	}
}

func (m LoginModel) Init() tea.Cmd {
	return nil
}

func (m LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter":
			if m.loading {
				return m, nil
			}
			switch m.cursor {
			case 0:
				m.loading = true
				return m, startLoginCmd("github")
			case 1:
				return m, tea.Quit
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		}

	case LoginResultMsg:
		if msg.Err != nil {
			m.err = msg.Err
			m.loading = false
			return m, nil
		}
		_ = auth.SaveSession(&domain.User{
			Name:       msg.User.Name,
			Email:      msg.User.Email,
			Provider:   msg.User.Provider,
			ProviderID: msg.User.UserID,
		})
		gothUser := msg.User
		return m, func() tea.Msg {
			return NavigateMsg{To: screenDashboard, GothUser: &gothUser}
		}
	}

	return m, nil
}

func (m LoginModel) View() string {
	if m.loading {
		return titleStyle.Render("SprintOS") +
			"\n\n" +
			normalStyle.Render("Opening browser... complete the login and come back.") +
			"\n"
	}

	s := titleStyle.Render("SprintOS") + "\n\n"
	s += normalStyle.Render("Welcome! Please choose how to log in:") + "\n\n"

	for i, choice := range m.choices {
		if i == m.cursor {
			s += selectedStyle.Render("> "+choice) + "\n"
		} else {
			s += normalStyle.Render("  "+choice) + "\n"
		}
	}

	if m.err != nil {
		s += "\n" + errorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error())) + "\n"
	}

	s += "\n" + normalStyle.Render("↑/↓ to move  •  enter to select  •  q to quit") + "\n"

	return s
}
