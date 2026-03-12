package tui

import "github.com/charmbracelet/lipgloss"

var (
	AppBg       = lipgloss.NewStyle().Background(lipgloss.Color("235")).Foreground(lipgloss.Color("252"))
	Header      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("25")).Padding(0, 1)
	Subtle      = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	Title       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229"))
	Section     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("117"))
	Box         = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("61")).Padding(0, 1)
	Highlight   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("221"))
	Accent      = lipgloss.NewStyle().Foreground(lipgloss.Color("81"))
	OK          = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	Warn        = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	Err         = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	Dim         = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	Filter      = lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("166")).Padding(0, 1)
	LogLine     = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	Help        = lipgloss.NewStyle().Foreground(lipgloss.Color("251")).Background(lipgloss.Color("237")).Padding(0, 1)
	MetricLabel = lipgloss.NewStyle().Foreground(lipgloss.Color("110"))
)
