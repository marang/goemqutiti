package traces

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/marang/emqutiti/ui"
)

// ViewMessages shows captured messages for a trace.
func (t *Component) ViewMessages() string {
	t.api.ResetElemPos()
	title := fmt.Sprintf("Trace %s", t.viewKey)
	if t.FilterQuery() != "" {
		t.Component.List().SetSize(t.api.Width()-4, t.api.TraceHeight()-1)
	} else {
		t.Component.List().SetSize(t.api.Width()-4, t.api.TraceHeight())
	}
	listLines := strings.Split(t.Component.List().View(), "\n")
	if t.FilterQuery() != "" {
		inner := t.api.Width() - 4
		filterLine := fmt.Sprintf("Filters: %s", t.FilterQuery())
		filterLine = ansi.Truncate(filterLine, inner, "")
		listLines = append([]string{filterLine}, listLines...)
	}
	help := ui.InfoStyle.Render("[esc] back")
	listLines = append(listLines, help)
	content := strings.Join(listLines, "\n")
	view := ui.LegendBox(content, title, t.api.Width()-2, t.api.TraceHeight(), ui.ColBlue, true, -1)
	return t.api.OverlayHelp(view)
}
