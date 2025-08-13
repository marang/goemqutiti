// Package ui provides shared TUI components and styles.
package ui

import "github.com/charmbracelet/lipgloss"

// Tooltip displays text inside a bordered box.
type Tooltip struct {
	Text  string
	Width int
	Style *lipgloss.Style
}

// View renders the tooltip at the given coordinates using lipgloss.Place.
func (t Tooltip) View(x, y int) string {
	style := TooltipStyle
	if t.Style != nil {
		style = *t.Style
	}
	box := style.Width(t.Width).Render(t.Text)
	w := lipgloss.Width(box)
	h := lipgloss.Height(box)
	return lipgloss.Place(x+w, y+h, lipgloss.Right, lipgloss.Bottom, box)
}

// RenderTooltip renders content in a tooltip at the specified coordinates.
// When focused is true, the tooltip uses the focused style.
func RenderTooltip(content string, x, y int, focused bool) string {
	style := TooltipStyle
	if focused {
		style = TooltipFocused
	}
	t := Tooltip{
		Text:  content,
		Width: lipgloss.Width(content),
		Style: &style,
	}
	return t.View(x, y)
}
