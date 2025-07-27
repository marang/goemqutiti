package main

import (
	"github.com/charmbracelet/lipgloss"

	"goemqutiti/ui"
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(ui.ColPink)
	blurredStyle = lipgloss.NewStyle().Foreground(ui.ColGray)
	cursorStyle  = focusedStyle
	noCursor     = lipgloss.NewStyle()
	borderStyle  = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(ui.ColBlue).Padding(0, 1)
	greenBorder  = borderStyle.BorderForeground(ui.ColGreen)
	chipStyle    = lipgloss.NewStyle().Padding(0, 1).MarginRight(1).Border(lipgloss.NormalBorder()).BorderForeground(ui.ColBlue).Faint(true)
	chipInactive = chipStyle.Foreground(ui.ColGray)
	infoStyle    = lipgloss.NewStyle().Foreground(ui.ColBlue).PaddingLeft(1)
	connStyle    = lipgloss.NewStyle().Foreground(ui.ColGray).PaddingLeft(1)
)
