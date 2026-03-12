package tui

import "github.com/charmbracelet/lipgloss"

var (
	Title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	Box   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
	OK    = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	Warn  = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	Err   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	Dim   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)
