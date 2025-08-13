package history

import tea "github.com/charmbracelet/bubbletea"

// Scroll delegates mouse wheel handling to the configured scroller.
func (h *Component) Scroll(msg tea.MouseMsg) tea.Cmd { return h.sc.Scroll(msg) }

// CanScroll reports whether the configured scroller can scroll.
func (h *Component) CanScroll() bool { return h.sc.CanScroll() }

// HandleSelection updates history selection based on index and shift key.
func (h *Component) HandleSelection(idx int, shift bool) {
	h.list.Select(idx)
	if shift {
		if h.selectionAnchor == -1 {
			h.selectionAnchor = h.list.Index()
			if h.selectionAnchor >= 0 && h.selectionAnchor < len(h.items) {
				v := true
				h.items[h.selectionAnchor].IsSelected = &v
			}
		}
		h.updateSelectionRange(idx)
	} else {
		for i := range h.items {
			h.items[i].IsSelected = nil
		}
		h.selectionAnchor = -1
	}
}

// HandleClick selects a history item based on mouse position.
func (h *Component) HandleClick(msg tea.MouseMsg, top, vpYOffset int) {
	idx := h.indexAt(msg.Y, top, vpYOffset)
	if idx >= 0 {
		h.HandleSelection(idx, msg.Shift)
	}
}

// UpdateSelectionRange selects history entries from the anchor to idx.
func (h *Component) UpdateSelectionRange(idx int) { h.updateSelectionRange(idx) }

func (h *Component) indexAt(y, top, vpYOffset int) int {
	rel := y - (top + 1) + vpYOffset
	if rel < 0 {
		return -1
	}
	hgt := 2 // history delegate height
	idx := rel / hgt
	start := h.list.Paginator.Page * h.list.Paginator.PerPage
	i := start + idx
	if i >= len(h.list.Items()) || i < 0 {
		return -1
	}
	return i
}

func (h *Component) updateSelectionRange(idx int) {
	start := h.selectionAnchor
	end := idx
	if start > end {
		start, end = end, start
	}
	for i := range h.items {
		h.items[i].IsSelected = nil
	}
	for i := start; i <= end && i < len(h.items); i++ {
		v := true
		h.items[i].IsSelected = &v
	}
}
