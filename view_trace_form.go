package emqutiti

import (
	"github.com/marang/emqutiti/ui"
)

// ViewForm renders the form for new traces.
func (t *tracesComponent) ViewForm() string {
	t.m.ui.elemPos = map[string]int{}
	content := t.form.View()
	view := ui.LegendBox(content, "New Trace", t.m.ui.width-2, 0, ui.ColBlue, true, -1)
	return t.m.overlayHelp(view)
}
