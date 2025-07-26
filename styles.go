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
	greenBorder  = borderStyle.BorderForeground(lipgloss.Color("34"))
	chipStyle    = lipgloss.NewStyle().Padding(0, 1).MarginRight(1).Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("63")).Faint(true)
	chipInactive = chipStyle.Foreground(lipgloss.Color("240"))
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("63")).PaddingLeft(1)
)

func legendBox(content, label string, width int, focused bool) string {
	color := lipgloss.Color("63")
	if focused {
		color = lipgloss.Color("205")
	}
	return legendStyledBox(content, label, width, color)
}

func legendGreenBox(content, label string, width int, focused bool) string {
	color := lipgloss.Color("34")
	if focused {
		color = lipgloss.Color("205")
	}
	return legendStyledBox(content, label, width, color)
}

func legendStyledBox(content, label string, width int, color lipgloss.Color) string {
	if width < lipgloss.Width(label)+4 {
		width = lipgloss.Width(label) + 4
	}
	b := lipgloss.RoundedBorder()
	cy := lipgloss.Color("51")
	top := lipgloss.NewStyle().Foreground(color).Render(b.TopLeft+" "+label+" "+strings.Repeat(b.Top, width-lipgloss.Width(label)-4)) +
		lipgloss.NewStyle().Foreground(cy).Render(b.TopRight)
	bottom := lipgloss.NewStyle().Foreground(cy).Render(b.BottomLeft + strings.Repeat(b.Bottom, width-2) + b.BottomRight)
	lines := strings.Split(content, "\n")
	for i, l := range lines {
		side := color
		if i == len(lines)-1 {
			side = cy
		}
		left := lipgloss.NewStyle().Foreground(color).Render(b.Left)
		right := lipgloss.NewStyle().Foreground(side).Render(b.Right)
		lines[i] = left + lipgloss.PlaceHorizontal(width-2, lipgloss.Left, l) + right
	}
	middle := strings.Join(lines, "\n")
	return top + "\n" + middle + "\n" + bottom
}
