package view

import (
	"github.com/marang/emqutiti/ui"
)

// viewTraceForm renders the form for new traces.
func (m model) viewTraceForm() string {
	m.ui.elemPos = map[string]int{}
	content := m.traces.form.View()
	view := ui.LegendBox(content, "New Trace", m.ui.width-2, 0, ui.ColBlue, true, -1)
	return m.overlayHelp(view)
}
