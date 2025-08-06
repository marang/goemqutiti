package importer

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/marang/emqutiti/constants"
	"github.com/marang/emqutiti/ui"
)

// updateMap handles mapping columns to fields.
func (m *Model) updateMap(msg tea.Msg) tea.Cmd {
	if len(m.form.Fields) == 0 {
		return nil
	}
	switch ev := msg.(type) {
	case tea.KeyMsg:
		switch ev.String() {
		case constants.KeyTab, constants.KeyShiftTab, constants.KeyUp, constants.KeyDown, constants.KeyK, constants.KeyJ:
			m.form.CycleFocus(ev)
		case constants.KeyCtrlN:
			m.step = stepTemplate
			m.tmpl.Focus()
			return nil
		case constants.KeyCtrlP:
			m.step = stepFile
			return nil
		}
	case tea.MouseMsg:
		if ev.Action == tea.MouseActionPress && ev.Button == tea.MouseButtonLeft {
			if ev.Y >= 1 && ev.Y-1 < len(m.form.Fields) {
				m.form.Focus = ev.Y - 1
			}
		}
	}
	m.form.ApplyFocus()
	cmd := m.form.Fields[m.form.Focus].Update(msg)
	if km, ok := msg.(tea.KeyMsg); ok && km.Type == tea.KeyEnter {
		m.step = stepTemplate
		m.tmpl.Focus()
	}
	return cmd
}

// viewMap renders the column mapping step.
func (m *Model) viewMap(bw, _ int) string {
	colw := 0
	for _, h := range m.headers {
		if w := lipgloss.Width(h); w > colw {
			colw = w
		}
	}
	var b strings.Builder
	for i, h := range m.headers {
		label := h
		if i == m.form.Focus {
			label = ui.FocusedStyle.Render(h)
		}
		fmt.Fprintf(&b, "%*s : %s\n", colw, label, m.form.Fields[i].View())
	}
	b.WriteString("\nUse a.b to nest fields\n[enter] continue  [ctrl+n] next  [ctrl+p] back")
	return ui.LegendBox(b.String(), "Map Columns", bw, 0, ui.ColBlue, true, -1)
}
