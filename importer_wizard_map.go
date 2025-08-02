package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/marang/goemqutiti/ui"
)

// updateMap handles mapping columns to fields.
func (w *ImportWizard) updateMap(msg tea.Msg) (tea.Model, tea.Cmd) {
	if len(w.form.Fields) == 0 {
		return w, nil
	}
	switch m := msg.(type) {
	case tea.KeyMsg:
		switch m.String() {
		case "tab", "shift+tab", "up", "down", "k", "j":
			w.form.CycleFocus(m)
		case "ctrl+n":
			w.step = stepTemplate
			w.tmpl.Focus()
			return w, nil
		case "ctrl+p":
			w.step = stepFile
			return w, nil
		}
	case tea.MouseMsg:
		if m.Action == tea.MouseActionPress && m.Button == tea.MouseButtonLeft {
			if m.Y >= 1 && m.Y-1 < len(w.form.Fields) {
				w.form.Focus = m.Y - 1
			}
		}
	}
	w.form.ApplyFocus()
	cmd := w.form.Fields[w.form.Focus].Update(msg)
	if km, ok := msg.(tea.KeyMsg); ok && km.Type == tea.KeyEnter {
		w.step = stepTemplate
		w.tmpl.Focus()
	}
	return w, cmd
}

// viewMap renders the column mapping step.
func (w *ImportWizard) viewMap(bw, _ int) string {
	colw := 0
	for _, h := range w.headers {
		if w := lipgloss.Width(h); w > colw {
			colw = w
		}
	}
	var b strings.Builder
	for i, h := range w.headers {
		label := h
		if i == w.form.Focus {
			label = ui.FocusedStyle.Render(h)
		}
		fmt.Fprintf(&b, "%*s : %s\n", colw, label, w.form.Fields[i].View())
	}
	b.WriteString("\nUse a.b to nest fields\n[enter] continue  [ctrl+n] next  [ctrl+p] back")
	return ui.LegendBox(b.String(), "Map Columns", bw, 0, ui.ColBlue, true, -1)
}
