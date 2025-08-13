package steps

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/marang/emqutiti/constants"
	"github.com/marang/emqutiti/ui"
)

// MapStep handles mapping columns to fields.
type MapStep struct{ *Base }

func NewMapStep(b *Base) *MapStep {
	b.Current = Map
	return &MapStep{Base: b}
}

func (s *MapStep) Update(msg tea.Msg) (Step, tea.Cmd) {
	if len(s.Form.Fields) == 0 {
		return s, nil
	}
	switch ev := msg.(type) {
	case tea.KeyMsg:
		switch ev.String() {
		case constants.KeyTab, constants.KeyShiftTab, constants.KeyUp, constants.KeyDown, constants.KeyK, constants.KeyJ:
			s.Form.CycleFocus(ev)
		case constants.KeyCtrlN:
			return NewTemplateStep(s.Base), nil
		case constants.KeyCtrlP:
			return NewFileStep(s.Base), nil
		}
	case tea.MouseMsg:
		if ev.Action == tea.MouseActionPress && ev.Button == tea.MouseButtonLeft {
			if ev.Y >= 1 && ev.Y-1 < len(s.Form.Fields) {
				s.Form.Focus = ev.Y - 1
			}
		}
	}
	s.Form.ApplyFocus()
	cmd := s.Form.Fields[s.Form.Focus].Update(msg)
	if km, ok := msg.(tea.KeyMsg); ok && km.Type == tea.KeyEnter {
		return NewTemplateStep(s.Base), cmd
	}
	return s, cmd
}

func (s *MapStep) View(bw, _ int) string {
	colw := 0
	for _, h := range s.Headers {
		if w := lipgloss.Width(h); w > colw {
			colw = w
		}
	}
	var b strings.Builder
	for i, h := range s.Headers {
		label := h
		if i == s.Form.Focus {
			label = ui.FocusedStyle.Render(h)
		}
		padding := strings.Repeat(" ", colw-lipgloss.Width(h))
		fmt.Fprintf(&b, "%s%s : %s\n", padding, label, s.Form.Fields[i].View())
	}
	b.WriteString("\nUse a.b to nest fields\n[enter] continue  [ctrl+n] next  [ctrl+p] back")
	return ui.LegendBox(b.String(), "Map Columns", bw, 0, ui.ColBlue, true, -1)
}
