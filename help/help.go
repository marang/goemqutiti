package help

import (
	_ "embed"
	"github.com/charmbracelet/lipgloss"
)

//go:generate go run github.com/charmbracelet/glow@latest -s dark help.md > help.txt

//go:embed help.txt
var rawHelp string

// helpText is the formatted help page displayed in the viewport.
var helpText = lipgloss.NewStyle().Margin(1, 2).Render(rawHelp)
