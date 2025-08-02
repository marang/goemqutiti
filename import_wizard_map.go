package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

// updateMap handles mapping columns to fields.
func (w *ImportWizard) updateMap(msg tea.Msg) (tea.Model, tea.Cmd) {
	if len(w.form.fields) == 0 {
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
			if m.Y >= 1 && m.Y-1 < len(w.form.fields) {
				w.form.focus = m.Y - 1
			}
		}
	}
	w.form.ApplyFocus()
	cmd := w.form.fields[w.form.focus].Update(msg)
	if km, ok := msg.(tea.KeyMsg); ok && km.Type == tea.KeyEnter {
		w.step = stepTemplate
		w.tmpl.Focus()
	}
	return w, cmd
}
