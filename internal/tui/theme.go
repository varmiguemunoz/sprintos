package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7C3AED"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED")).Bold(true)
	normalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444"))
)
