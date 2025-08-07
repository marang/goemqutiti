package ui

import (
	"math"
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"
)

// LegendBox renders a bordered box with a label and optional height.
// The border color can be customized, and when focused the border
// is highlighted in pink.
func LegendBox(content, label string, width, height int, border lipgloss.Color, focused bool, scroll float64) string {
	col := border
	if focused {
		col = ColPink
	}
	return legendStyledBox(content, label, width, height, col, scroll)
}

func legendStyledBox(content, label string, width, height int, color lipgloss.Color, scroll float64) string {
	content = strings.TrimRight(content, "\n")
	if width < lipgloss.Width(label)+4 {
		width = lipgloss.Width(label) + 4
	}

	b := lipgloss.RoundedBorder()
	cy := ColCyan
	top := lipgloss.NewStyle().Foreground(color).Render(
		b.TopLeft+" "+label+" "+strings.Repeat(b.Top, width-lipgloss.Width(label)-4),
	) + lipgloss.NewStyle().Foreground(cy).Render(b.TopRight)
	bottom := lipgloss.NewStyle().Foreground(cy).Render(
		b.BottomLeft + strings.Repeat(b.Bottom, width-2) + b.BottomRight,
	)

	lines := strings.Split(content, "\n")
	if height > 0 {
		if len(lines) > height {
			lines = lines[:height]
		}
		for len(lines) < height {
			lines = append(lines, "")
		}
	}

	indicator := -1
	if scroll >= 0 && height > 0 {
		indicator = int(math.Round(scroll * float64(height-1)))
		if indicator < 0 {
			indicator = 0
		}
		if indicator >= height {
			indicator = height - 1
		}
	}

	for i, l := range lines {
		l = strings.TrimRightFunc(l, unicode.IsSpace)
		side := color
		if i == len(lines)-1 {
			side = cy
		}
		left := lipgloss.NewStyle().Foreground(color).Render(b.Left)
		rightChar := string(b.Right)
		if i == indicator {
			rightChar = "⧱"
			side = cy
		}
		right := lipgloss.NewStyle().Foreground(side).Render(rightChar)
		lines[i] = left + lipgloss.PlaceHorizontal(width-2, lipgloss.Left, l) + right
	}
	middle := strings.Join(lines, "\n")
	return top + "\n" + middle + "\n" + bottom
}
