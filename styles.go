package main

import (
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(colPink)
	blurredStyle = lipgloss.NewStyle().Foreground(colGray)
	cursorStyle  = focusedStyle
	noCursor     = lipgloss.NewStyle()
	borderStyle  = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(colBlue).Padding(0, 1)
	greenBorder  = borderStyle.BorderForeground(colGreen)
	chipStyle    = lipgloss.NewStyle().Padding(0, 1).MarginRight(1).Border(lipgloss.NormalBorder()).BorderForeground(colBlue).Faint(true)
	chipInactive = chipStyle.Foreground(colGray)
	infoStyle    = lipgloss.NewStyle().Foreground(colBlue).PaddingLeft(1)
	connStyle    = lipgloss.NewStyle().Foreground(colGray).PaddingLeft(1)
)

func legendBox(content, label string, width int, focused bool) string {
	color := colBlue
	if focused {
		color = colPink
	}
	return legendStyledBox(content, label, width, color)
}

func legendGreenBox(content, label string, width int, focused bool) string {
	color := colGreen
	if focused {
		color = colPink
	}
	return legendStyledBox(content, label, width, color)
}

func legendStyledBox(content, label string, width int, color lipgloss.Color) string {
	content = strings.TrimRight(content, "\n")
	if width < lipgloss.Width(label)+4 {
		width = lipgloss.Width(label) + 4
	}

	b := lipgloss.RoundedBorder()
	cy := colCyan
	top := lipgloss.NewStyle().Foreground(color).Render(
		b.TopLeft+" "+label+" "+strings.Repeat(b.Top, width-lipgloss.Width(label)-4),
	) + lipgloss.NewStyle().Foreground(cy).Render(b.TopRight)
	bottom := lipgloss.NewStyle().Foreground(cy).Render(
		b.BottomLeft + strings.Repeat(b.Bottom, width-2) + b.BottomRight,
	)

	lines := strings.Split(content, "\n")
	for i, l := range lines {
		l = strings.TrimRightFunc(l, unicode.IsSpace)
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
