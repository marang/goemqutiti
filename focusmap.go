package main

import tea "github.com/charmbracelet/bubbletea"

// Focusable represents a UI element that can gain or lose focus.
type Focusable interface {
	Focus()
	Blur()
	IsFocused() bool
	View() string
}

type teaFocusable interface {
	Focus() tea.Cmd
	Blur()
	Focused() bool
	View() string
}

// adapt converts a Bubble Tea focusable model into the Focusable interface.
func adapt(f teaFocusable) Focusable { return focusAdapter{f} }

type focusAdapter struct{ f teaFocusable }

func (a focusAdapter) Focus()          { _ = a.f.Focus() }
func (a focusAdapter) Blur()           { a.f.Blur() }
func (a focusAdapter) IsFocused() bool { return a.f.Focused() }
func (a focusAdapter) View() string    { return a.f.View() }

// nullFocusable is a no-op focusable used for non-interactive areas.
type nullFocusable struct{ focused bool }

func (n *nullFocusable) Focus()          { n.focused = true }
func (n *nullFocusable) Blur()           { n.focused = false }
func (n *nullFocusable) IsFocused() bool { return n.focused }
func (n *nullFocusable) View() string    { return "" }

// FocusMap manages focus among a set of focusable elements.
type FocusMap struct {
	items      []Focusable
	focusIndex int
}

// NewFocusMap creates a FocusMap and focuses the first element if present.
func NewFocusMap(items []Focusable) *FocusMap {
	fm := &FocusMap{items: items}
	if len(items) > 0 {
		items[0].Focus()
	}
	return fm
}

// Index returns the currently focused index.
func (fm *FocusMap) Index() int { return fm.focusIndex }

// Set focuses the element at the given index.
func (fm *FocusMap) Set(i int) {
	if len(fm.items) == 0 || i < 0 || i >= len(fm.items) {
		return
	}
	for idx, it := range fm.items {
		if it == nil {
			continue
		}
		if idx == i {
			it.Focus()
		} else {
			it.Blur()
		}
	}
	fm.focusIndex = i
}

// Next moves focus to the next element.
func (fm *FocusMap) Next() { fm.Set((fm.focusIndex + 1) % len(fm.items)) }

// Prev moves focus to the previous element.
func (fm *FocusMap) Prev() { fm.Set((fm.focusIndex - 1 + len(fm.items)) % len(fm.items)) }
