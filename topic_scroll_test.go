package emqutiti

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/marang/emqutiti/ui"
)

func TestTopicsScrollDown(t *testing.T) {
	m, _ := initialModel(nil)
	m.Update(tea.WindowSizeMsg{Width: 40, Height: 20})
	setupManyTopics(m, 10)
	m.layout.topics.height = 2
	m.viewClient()
	if m.topics.vp.YOffset != 0 {
		t.Fatalf("expected initial scroll 0")
	}
	m.setFocus(idTopics)
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	rowH := lipgloss.Height(ui.ChipStyle.Render("t"))
	if m.topics.vp.YOffset != rowH {
		t.Fatalf("expected scroll %d got %d", rowH, m.topics.vp.YOffset)
	}
}

func TestTopicsScrollDownJ(t *testing.T) {
	m, _ := initialModel(nil)
	m.Update(tea.WindowSizeMsg{Width: 40, Height: 20})
	setupManyTopics(m, 10)
	m.layout.topics.height = 2
	m.viewClient()
	if m.topics.vp.YOffset != 0 {
		t.Fatalf("expected initial scroll 0")
	}
	m.setFocus(idTopics)
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	rowH := lipgloss.Height(ui.ChipStyle.Render("t"))
	if m.topics.vp.YOffset != rowH {
		t.Fatalf("expected scroll %d got %d", rowH, m.topics.vp.YOffset)
	}
}

func TestTopicSelectionScroll(t *testing.T) {
	m, _ := initialModel(nil)
	m.Update(tea.WindowSizeMsg{Width: 40, Height: 20})
	setupManyTopics(m, 10)
	m.layout.topics.height = 2
	m.viewClient()
	m.setFocus(idTopics)
	m.topics.selected = 0
	// Move selection to the 9th item which resides on the third row
	for i := 0; i < 8; i++ {
		_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	}
	rowH := lipgloss.Height(ui.ChipStyle.Render("t"))
	if m.topics.selected != 8 {
		t.Fatalf("expected selected index 8 got %d", m.topics.selected)
	}
	if m.topics.vp.YOffset != rowH*1 {
		t.Fatalf("expected scroll %d got %d", rowH*1, m.topics.vp.YOffset)
	}
}
