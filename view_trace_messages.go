package emqutiti

import (
	"fmt"
	"strings"

	"github.com/marang/emqutiti/ui"
)

// ViewMessages shows captured messages for a trace.
func (t *tracesComponent) ViewMessages() string {
	t.m.ui.elemPos = map[string]int{}
	title := fmt.Sprintf("Trace %s", t.viewKey)
	listLines := strings.Split(t.view.View(), "\n")
	help := ui.InfoStyle.Render("[esc] back")
	listLines = append(listLines, help)
	target := len(listLines)
	minHeight := t.m.layout.trace.height + 1
	if target < minHeight {
		for len(listLines) < minHeight {
			listLines = append(listLines, "")
		}
		target = minHeight
	}
	content := strings.Join(listLines, "\n")
	view := ui.LegendBox(content, title, t.m.ui.width-2, target, ui.ColBlue, true, -1)
	return t.m.overlayHelp(view)
}
