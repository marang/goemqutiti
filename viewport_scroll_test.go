package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestViewportScrollCtrlJ(t *testing.T) {
	m := initialModel(nil)
	m.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	for i := 0; i < 50; i++ {
		m.appendHistory("t", "msg", "pub", "")
	}
	m.viewClient()
	if m.ui.viewport.YOffset != 0 {
		t.Fatalf("expected initial offset 0")
	}
	m.Update(tea.KeyMsg{Type: tea.KeyCtrlJ})
	if m.ui.viewport.YOffset == 0 {
		t.Fatalf("expected viewport to scroll down")
	}
}
