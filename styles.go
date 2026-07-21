package main

import "github.com/charmbracelet/lipgloss"

var (
	colorAccent = lipgloss.Color("#7D56F4")
	colorMuted  = lipgloss.Color("240")
	colorText   = lipgloss.Color("252")
	colorError  = lipgloss.Color("203")
	colorOk     = lipgloss.Color("78")

	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(colorAccent)

	sidebarStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorMuted).
			Padding(0, 1)

	contentStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorMuted).
			Padding(0, 1)

	selectedItemStyle = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	itemStyle         = lipgloss.NewStyle().Foreground(colorText)
	helpStyle         = lipgloss.NewStyle().Foreground(colorMuted).Padding(0, 1)
	statusStyle       = lipgloss.NewStyle().Foreground(colorOk)
	errorStyle        = lipgloss.NewStyle().Foreground(colorError)
)
