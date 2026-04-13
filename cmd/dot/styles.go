package main

import "github.com/charmbracelet/lipgloss"

var (
	colorPrimary = lipgloss.Color("12")  // bright blue
	colorSuccess = lipgloss.Color("10")  // bright green
	colorError   = lipgloss.Color("9")   // bright red
	colorMuted   = lipgloss.Color("8")   // dark gray
	colorAccent  = lipgloss.Color("105") // purple

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			MarginTop(1).
			MarginBottom(1)

	subheaderStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			MarginBottom(1)

	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorSuccess)

	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorError)

	mutedStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Italic(true)

	commandNameStyle = lipgloss.NewStyle().
				Foreground(colorAccent).
				Bold(true)

	commandDescStyle = lipgloss.NewStyle().
				Foreground(colorMuted)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(0, 1)
)

const dotBanner = `
 ██████╗  ██████╗ ████████╗
 ██╔══██╗██╔═══██╗╚══██╔══╝
 ██║  ██║██║   ██║   ██║
 ██║  ██║██║   ██║   ██║
 ██████╔╝╚██████╔╝   ██║
 ╚═════╝  ╚═════╝    ╚═╝   `
