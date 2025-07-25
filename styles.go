package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle  = focusedStyle
	noCursor     = lipgloss.NewStyle()
	borderStyle  = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63")).Padding(0, 1)
	greenBorder  = borderStyle.Copy().BorderForeground(lipgloss.Color("34"))
	chipStyle    = lipgloss.NewStyle().Padding(0, 1).MarginRight(1).Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("63")).Faint(true)
	chipInactive = chipStyle.Copy().Foreground(lipgloss.Color("240"))
)

func legendBox(content, label string, width int) string {
	return legendStyledBox(content, label, width, borderStyle)
}

func legendGreenBox(content, label string, width int) string {
	return legendStyledBox(content, label, width, greenBorder)
}

func legendStyledBox(content, label string, width int, style lipgloss.Style) string {
	b := lipgloss.RoundedBorder()
	top := b.TopLeft + " " + label + " " + strings.Repeat(b.Top, width-lipgloss.Width(label)-3) + b.TopRight
	bottom := b.BottomLeft + strings.Repeat(b.Bottom, width-2) + b.BottomRight
	lines := strings.Split(content, "\n")
	for i, l := range lines {
		lines[i] = b.Left + lipgloss.PlaceHorizontal(width-2, lipgloss.Left, l) + b.Right
	}
	middle := strings.Join(lines, "\n")
	return style.Render(top + "\n" + middle + "\n" + bottom)
}
