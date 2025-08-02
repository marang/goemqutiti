package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// updateFile handles the file selection step.
func (w *ImportWizard) updateFile(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	w.file, cmd = w.file.Update(msg)
	if km, ok := msg.(tea.KeyMsg); ok && (km.Type == tea.KeyEnter || km.Type == tea.KeyCtrlN) {
		path := strings.TrimSpace(w.file.Value())
		if path == "" {
			return w, nil
		}
		rows, err := ReadFile(path)
		if err != nil {
			w.file.SetValue(path + " (" + err.Error() + ")")
			return w, nil
		}
		if len(rows) == 0 {
			w.file.SetValue(path + " (no data)")
			return w, nil
		}
		w.rows = rows
		var fields []formField
		for k := range rows[0] {
			w.headers = append(w.headers, k)
			fi := newTextField(k, "")
			fields = append(fields, fi)
		}
		if len(fields) > 0 {
			w.form = Form{fields: fields, focus: 0}
			w.form.ApplyFocus()
		}
		w.step = stepMap
	}
	return w, cmd
}
