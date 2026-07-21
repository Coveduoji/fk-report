package main

import "github.com/charmbracelet/lipgloss"

var (
	colorAccent = lipgloss.Color("#E63946")
	colorMuted  = lipgloss.Color("240")
	colorText   = lipgloss.Color("252")
	colorError  = lipgloss.Color("208")
	colorOk     = lipgloss.Color("78")
	colorDone   = lipgloss.Color("240")

	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(colorAccent)

	doneItemStyle = lipgloss.NewStyle().Foreground(colorDone).Strikethrough(true)
	buttonStyle   = lipgloss.NewStyle().Padding(0, 3).Foreground(colorText)
	buttonActiveStyle = lipgloss.NewStyle().Padding(0, 3).Bold(true).
				Foreground(lipgloss.Color("230")).Background(colorAccent)
	previewHeaderStyle = lipgloss.NewStyle().Foreground(colorMuted).Italic(true)

	bannerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorAccent).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).
			Padding(1, 4)

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
