package help

import (
	_ "embed"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

//go:embed help.md
var helpMarkdown string

// helpText is the formatted help page displayed in the viewport.
var helpText = renderHelp()

// renderHelp converts the embedded Markdown into a styled string using
// Glow's glamour renderer.
func renderHelp() string {
	out, err := glamour.Render(helpMarkdown, "dark")
	if err != nil {
		return helpMarkdown
	}
	return lipgloss.NewStyle().Margin(1, 2).Render(out)
}
