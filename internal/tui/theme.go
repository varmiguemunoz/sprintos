package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7C3AED"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED")).Bold(true)
	normalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444"))
	successStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981"))
	warningStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#4B5563"))
	labelStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	valueStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#F9FAFB")).Bold(true)

	highlightStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1E1540")).
			Foreground(lipgloss.Color("#A78BFA"))

	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#374151")).
			Padding(0, 1)

	activeCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C3AED")).
			Padding(0, 1)

	hintKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED")).
			Bold(true)
)

func sectionHeader(title string) string {
	return selectedStyle.Render("╴ "+title)
}

func renderHintBar(pairs ...string) string {
	if len(pairs)%2 != 0 {
		return ""
	}
	parts := make([]string, 0, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		parts = append(parts, hintKeyStyle.Render(pairs[i])+" "+normalStyle.Render(pairs[i+1]))
	}
	return strings.Join(parts, dimStyle.Render("  •  "))
}

func renderBar(value, maxValue, barWidth int, style lipgloss.Style) string {
	if maxValue == 0 || barWidth <= 0 {
		return ""
	}
	n := value * barWidth / maxValue
	if n == 0 && value > 0 {
		n = 1
	}
	if n > barWidth {
		n = barWidth
	}
	return style.Render(strings.Repeat("█", n))
}
