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
	chipStyle    = lipgloss.NewStyle().Padding(0, 1).MarginRight(1).Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("63"))
	chipInactive = chipStyle.Copy().Foreground(lipgloss.Color("240"))
)

func legendBox(content, label string, width int) string {
	b := lipgloss.RoundedBorder()
	top := b.TopLeft + " " + label + " " + strings.Repeat(b.Top, width-lipgloss.Width(label)-3) + b.TopRight
	bottom := b.BottomLeft + strings.Repeat(b.Bottom, width-2) + b.BottomRight
	lines := strings.Split(content, "\n")
	for i, l := range lines {
		lines[i] = b.Left + lipgloss.PlaceHorizontal(width-2, lipgloss.Left, l) + b.Right
	}
	middle := strings.Join(lines, "\n")
	return top + "\n" + middle + "\n" + bottom
}
