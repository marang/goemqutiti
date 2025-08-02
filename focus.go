package main

import tea "github.com/charmbracelet/bubbletea"

// setFocus moves focus to the given element id.
func (m *model) setFocus(id string) tea.Cmd {
	for i, name := range m.ui.focusOrder {
		if name == id {
			m.focus.Set(i)
			m.ui.focusIndex = m.focus.Index()
			break
		}
	}
	m.scrollToFocused()
	return nil
}

// focusFromMouse determines which element was clicked and focuses it.
func (m *model) focusFromMouse(y int) tea.Cmd {
	cy := y + m.ui.viewport.YOffset - 1
	chosen := ""
	maxPos := -1
	for _, id := range m.ui.focusOrder {
		if pos, ok := m.ui.elemPos[id]; ok && cy >= pos && pos > maxPos {
			chosen = id
			maxPos = pos
		}
	}
	if chosen != "" {
		if chosen != m.ui.focusOrder[m.ui.focusIndex] {
			return m.setFocus(chosen)
		}
		return nil
	}
	if len(m.ui.focusOrder) > 0 && m.ui.focusOrder[m.ui.focusIndex] != m.ui.focusOrder[0] {
		return m.setFocus(m.ui.focusOrder[0])
	}
	return nil
}

// scrollToFocused ensures the focused element is visible in the viewport.
func (m *model) scrollToFocused() {
	if len(m.ui.focusOrder) == 0 {
		return
	}
	id := m.ui.focusOrder[m.ui.focusIndex]
	pos, ok := m.ui.elemPos[id]
	if !ok {
		return
	}
	offset := pos - 1
	if offset < 0 {
		offset = 0
	}
	if offset < m.ui.viewport.YOffset {
		m.ui.viewport.SetYOffset(offset)
	} else if offset >= m.ui.viewport.YOffset+m.ui.viewport.Height {
		m.ui.viewport.SetYOffset(offset - m.ui.viewport.Height + 1)
	}
}
