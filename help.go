package main

import (
	_ "embed"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"
	glowutils "github.com/charmbracelet/glow/v2/utils"
)

//go:embed docs/help.md
var helpMarkdown string

// helpText is the rendered markdown displayed in the help viewport.
var helpText = renderHelp()

// renderHelp converts the embedded markdown to ANSI formatted text using
// Glow's glamour helpers. On error it falls back to the raw markdown.
func renderHelp() string {
	r, err := glamour.NewTermRenderer(
		glowutils.GlamourStyle(styles.DarkStyle, false),
		glamour.WithWordWrap(0),
	)
	if err != nil {
		return helpMarkdown
	}
	out, err := r.Render(helpMarkdown)
	if err != nil {
		return helpMarkdown
	}
	return out
}
