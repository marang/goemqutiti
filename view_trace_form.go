package emqutiti

import (
	"github.com/marang/emqutiti/ui"
)

// ViewForm renders the form for new traces.
func (t *tracesComponent) ViewForm() string {
	t.api.ResetElemPos()
	content := t.form.View()
	view := ui.LegendBox(content, "New Trace", t.api.Width()-2, 0, ui.ColBlue, true, -1)
	return t.api.OverlayHelp(view)
}
