package emqutiti

import tea "github.com/charmbracelet/bubbletea"

// handleSpaceKey toggles selection in history.
func (m *model) handleSpaceKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] == idHistory && !m.history.showArchived {
		idx := m.history.list.Index()
		if idx >= 0 && idx < len(m.history.items) {
			if m.history.items[idx].isSelected != nil && *m.history.items[idx].isSelected {
				m.history.items[idx].isSelected = nil
			} else {
				v := true
				m.history.items[idx].isSelected = &v
			}
			m.history.selectionAnchor = idx
		}
	}
	return nil
}

// handleShiftUpKey extends selection upward in history.
func (m *model) handleShiftUpKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] == idHistory && !m.history.showArchived {
		if m.history.selectionAnchor == -1 {
			m.history.selectionAnchor = m.history.list.Index()
			if m.history.selectionAnchor >= 0 && m.history.selectionAnchor < len(m.history.items) {
				v := true
				m.history.items[m.history.selectionAnchor].isSelected = &v
			}
		}
		if m.history.list.Index() > 0 {
			m.history.list.CursorUp()
			idx := m.history.list.Index()
			m.updateSelectionRange(idx)
		}
	}
	return nil
}

// handleShiftDownKey extends selection downward in history.
func (m *model) handleShiftDownKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] == idHistory && !m.history.showArchived {
		if m.history.selectionAnchor == -1 {
			m.history.selectionAnchor = m.history.list.Index()
			if m.history.selectionAnchor >= 0 && m.history.selectionAnchor < len(m.history.items) {
				v := true
				m.history.items[m.history.selectionAnchor].isSelected = &v
			}
		}
		if m.history.list.Index() < len(m.history.list.Items())-1 {
			m.history.list.CursorDown()
			idx := m.history.list.Index()
			m.updateSelectionRange(idx)
		}
	}
	return nil
}

// handleSelectAllKey selects or clears all history items.
func (m *model) handleSelectAllKey() tea.Cmd {
	if m.ui.focusOrder[m.ui.focusIndex] == idHistory && !m.history.showArchived {
		allSelected := true
		for i := range m.history.items {
			if m.history.items[i].isSelected == nil || !*m.history.items[i].isSelected {
				allSelected = false
				break
			}
		}
		if allSelected {
			for i := range m.history.items {
				m.history.items[i].isSelected = nil
			}
			m.history.selectionAnchor = -1
		} else {
			for i := range m.history.items {
				v := true
				m.history.items[i].isSelected = &v
			}
			if len(m.history.items) > 0 {
				m.history.selectionAnchor = 0
			}
		}
	}
	return nil
}

// handleHistoryScroll handles history scroll keys.
func (m *model) handleHistoryScroll(_ string) tea.Cmd {
	// keep current selection and anchor
	return nil
}
