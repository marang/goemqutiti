package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"goemqutiti/ui"
)

func TestTopicsScrollDown(t *testing.T) {
	m := initialModel(nil)
	m.Update(tea.WindowSizeMsg{Width: 40, Height: 20})
	setupManyTopics(m, 10)
	m.layout.topics.height = 2
	m.viewClient()
	if m.topics.vp.YOffset != 0 {
		t.Fatalf("expected initial scroll 0")
	}
	m.setFocus("topics")
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	rowH := lipgloss.Height(ui.ChipStyle.Render("t"))
	if m.topics.vp.YOffset != rowH {
		t.Fatalf("expected scroll %d got %d", rowH, m.topics.vp.YOffset)
	}
}
