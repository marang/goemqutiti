package ui

import (
	list "github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// MouseScroller handles mouse wheel scrolling for content displayed inside a
// LegendBox. Implementations may customize the scroll behaviour.
// CanScroll reports whether there is additional content beyond the visible
// area.
type MouseScroller interface {
	Scroll(tea.MouseMsg) tea.Cmd
	CanScroll() bool
}

// NewListMouseScroller returns a MouseScroller for a bubbles list. The list
// moves by delta rows per wheel tick.
func NewListMouseScroller(lst *list.Model, delta int) MouseScroller {
	if delta <= 0 {
		delta = 1
	}
	return &listMouseScroller{list: lst, delta: delta}
}

type listMouseScroller struct {
	list  *list.Model
	delta int
}

func (s *listMouseScroller) Scroll(msg tea.MouseMsg) tea.Cmd {
	switch msg.Button {
	case tea.MouseButtonWheelDown:
		for i := 0; i < s.delta; i++ {
			s.list.CursorDown()
		}
	case tea.MouseButtonWheelUp:
		for i := 0; i < s.delta; i++ {
			s.list.CursorUp()
		}
	}
	return nil
}

func (s *listMouseScroller) CanScroll() bool {
	return len(s.list.Items()) > s.list.Paginator.PerPage
}
