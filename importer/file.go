package importer

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/marang/emqutiti/ui"
)

// updateFile handles the file selection step.
func (m *Model) updateFile(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.file, cmd = m.file.Update(msg)
	if km, ok := msg.(tea.KeyMsg); ok && (km.Type == tea.KeyEnter || km.Type == tea.KeyCtrlN) {
		path := strings.TrimSpace(m.file.Value())
		if path == "" {
			return cmd
		}
		rows, err := ReadFile(path)
		if err != nil {
			m.file.SetValue(path + " (" + err.Error() + ")")
			return cmd
		}
		if len(rows) == 0 {
			m.file.SetValue(path + " (no data)")
			return cmd
		}
		m.rows = rows
		var fields []ui.Field
		for k := range rows[0] {
			m.headers = append(m.headers, k)
			fi := ui.NewTextField(k, "")
			fields = append(fields, fi)
		}
		if len(fields) > 0 {
			m.form = ui.Form{Fields: fields, Focus: 0}
			m.form.ApplyFocus()
		}
		m.step = stepMap
	}
	return cmd
}

// viewFile renders the file selection step.
func (m *Model) viewFile(bw, _ int) string {
	content := m.file.View() + "\n[enter] load file  [ctrl+n] next"
	return ui.LegendBox(content, "Import", bw, 0, ui.ColBlue, true, -1)
}
