package traces

import (
	"fmt"
	"strings"

	"github.com/marang/emqutiti/ui"
)

// ViewMessages shows captured messages for a trace.
func (t *Component) ViewMessages() string {
	t.api.ResetElemPos()
	title := fmt.Sprintf("Trace %s", t.viewKey)
	listLines := strings.Split(t.view.View(), "\n")
	help := ui.InfoStyle.Render("[esc] back")
	listLines = append(listLines, help)
	target := len(listLines)
	minHeight := t.api.TraceHeight() + 1
	if target < minHeight {
		for len(listLines) < minHeight {
			listLines = append(listLines, "")
		}
		target = minHeight
	}
	content := strings.Join(listLines, "\n")
	view := ui.LegendBox(content, title, t.api.Width()-2, target, ui.ColBlue, true, -1)
	return t.api.OverlayHelp(view)
}
