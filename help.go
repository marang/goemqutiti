package emqutiti

import (
	_ "embed"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/marang/emqutiti/ui"
)

//go:embed docs/help.md
var helpMarkdown string

// helpText is the formatted help page displayed in the viewport.
var helpText = renderHelp()

// renderHelp parses the embedded Markdown and applies lipgloss styles.
func renderHelp() string {
	lines := strings.Split(helpMarkdown, "\n")
	var tableRows []string
	var bullets []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if strings.HasPrefix(l, "|") {
			parts := strings.Split(l, "|")
			if len(parts) >= 3 && strings.TrimSpace(parts[1]) != "Key" {
				key := strings.TrimSpace(parts[1])
				action := strings.TrimSpace(parts[2])
				row := lipgloss.JoinHorizontal(lipgloss.Left,
					ui.HelpKey.Render(key), action)
				tableRows = append(tableRows, row)
			}
		} else if strings.HasPrefix(l, "-") {
			bullets = append(bullets, "• "+strings.TrimSpace(strings.TrimPrefix(l, "-")))
		}
	}
	table := lipgloss.JoinVertical(lipgloss.Left, tableRows...)
	list := lipgloss.JoinVertical(lipgloss.Left, bullets...)
	content := lipgloss.JoinVertical(lipgloss.Left,
		ui.HelpHeader.Render("Shortcuts"),
		table,
		"",
		ui.HelpHeader.Render("Other Keys"),
		list,
	)
	return lipgloss.NewStyle().Margin(1, 2).Render(content)
}
