package emqutiti

import (
	"fmt"
	"strings"

	"github.com/marang/emqutiti/ui"
)

// viewTraceMessages shows captured messages for a trace.
func (m model) viewTraceMessages() string {
	m.ui.elemPos = map[string]int{}
	title := fmt.Sprintf("Trace %s", m.traces.viewKey)
	listLines := strings.Split(m.traces.view.View(), "\n")
	help := ui.InfoStyle.Render("[esc] back")
	listLines = append(listLines, help)
	target := len(listLines)
	minHeight := m.layout.trace.height + 1
	if target < minHeight {
		for len(listLines) < minHeight {
			listLines = append(listLines, "")
		}
		target = minHeight
	}
	content := strings.Join(listLines, "\n")
	view := ui.LegendBox(content, title, m.ui.width-2, target, ui.ColBlue, true, -1)
	return m.overlayHelp(view)
}
