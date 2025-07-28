package main

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"goemqutiti/ui"
)

func chipCoords(m *model, idx int) (int, int) {
	if idx < 0 || idx >= len(m.chipBounds) {
		return -1, -1
	}
	b := m.chipBounds[idx]
	return b.x, b.y - m.viewport.YOffset
}

func setupTopics(m *model) {
	names := []string{"testtopic", "asdfsedf", "asdasd", "sdfdfasssssd", "asdasdasss", "asasasdfffa", "asasdfa", "aasdf", "asdfa", "asdasasdfasdf"}
	for _, n := range names {
		m.topics = append(m.topics, topicItem{title: n, active: true})
	}
}

func TestMouseToggleFirstTopic(t *testing.T) {
	m := initialModel(nil)
	m.Update(tea.WindowSizeMsg{Width: 40, Height: 20})
	setupTopics(m)
	m.viewClient()
	x, y := chipCoords(m, 0)
	for offset := 0; offset < m.chipBounds[0].h; offset++ {
		activeBefore := m.topics[0].active
		m.Update(tea.MouseMsg{Type: tea.MouseLeft, X: x, Y: y + offset})
		if m.selectedTopic != 0 {
			t.Fatalf("expected selected topic 0, got %d", m.selectedTopic)
		}
		if m.topics[0].active == activeBefore {
			t.Fatalf("click offset %d did not toggle topic", offset)
		}
	}
}

func TestMouseToggleThirdRowTopic(t *testing.T) {
	m := initialModel(nil)
	m.Update(tea.WindowSizeMsg{Width: 40, Height: 20})
	setupTopics(m)
	m.viewClient()
	// topic index 6 resides on third row
	x, y := chipCoords(m, 6)
	for offset := 0; offset < m.chipBounds[6].h; offset++ {
		before := m.topics[6].active
		m.Update(tea.MouseMsg{Type: tea.MouseLeft, X: x, Y: y + offset})
		if m.selectedTopic != 6 {
			t.Fatalf("expected selected topic 6, got %d", m.selectedTopic)
		}
		if m.topics[6].active == before {
			t.Fatalf("offset %d did not toggle topic 6", offset)
		}
	}
}

func TestMouseToggleFourthRowTopic(t *testing.T) {
	m := initialModel(nil)
	m.Update(tea.WindowSizeMsg{Width: 40, Height: 20})
	setupTopics(m)
	m.viewClient()
	// topic index 8 resides on the fourth row
	x, y := chipCoords(m, 8)
	for offset := 0; offset < m.chipBounds[8].h; offset++ {
		before := m.topics[8].active
		m.Update(tea.MouseMsg{Type: tea.MouseLeft, X: x, Y: y + offset})
		if m.selectedTopic != 8 {
			t.Fatalf("expected selected topic 8, got %d", m.selectedTopic)
		}
		if m.topics[8].active == before {
			t.Fatalf("offset %d did not toggle topic 8", offset)
		}
	}
}

func setupManyTopics(m *model, n int) {
	for i := 0; i < n; i++ {
		title := fmt.Sprintf("topic-%d", i)
		m.topics = append(m.topics, topicItem{title: title, active: true})
	}
}

func TestMouseToggleFifteenthRowTopic(t *testing.T) {
	m := initialModel(nil)
	// Enough height for many rows
	m.Update(tea.WindowSizeMsg{Width: 40, Height: 80})
	setupManyTopics(m, 50)
	m.viewClient()
	// Index of the first chip on the 15th row (0-based rows, 3 chips per row)
	idx := 14 * 3
	x, y := chipCoords(m, idx)
	for offset := 0; offset < m.chipBounds[idx].h; offset++ {
		before := m.topics[idx].active
		m.Update(tea.MouseMsg{Type: tea.MouseLeft, X: x, Y: y + offset})
		if m.selectedTopic != idx {
			t.Fatalf("expected selected topic %d, got %d", idx, m.selectedTopic)
		}
		if m.topics[idx].active == before {
			t.Fatalf("offset %d did not toggle topic %d", offset, idx)
		}
	}
}

func TestMouseToggleWithScroll(t *testing.T) {
	m := initialModel(nil)
	// Small height so we need to scroll to reach later rows
	m.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	setupManyTopics(m, 30)
	m.viewClient()
	// Scroll viewport to show around row 6
	scroll := m.elemPos["topics"] + 2 + 6*lipgloss.Height(ui.ChipStyle.Render("test"))
	m.viewport.SetYOffset(scroll)
	if m.viewport.YOffset != scroll {
		t.Fatalf("expected YOffset %d got %d", scroll, m.viewport.YOffset)
	}

	// Choose a chip on row 7 (0-based index -> row 7 => start index 6*3)
	idx := 6 * 3
	x, y := chipCoords(m, idx)
	for offset := 0; offset < m.chipBounds[idx].h; offset++ {
		before := m.topics[idx].active
		m.Update(tea.MouseMsg{Type: tea.MouseLeft, X: x, Y: y + offset})
		if m.selectedTopic != idx {
			t.Fatalf("expected selected %d got %d", idx, m.selectedTopic)
		}
		if m.topics[idx].active == before {
			t.Fatalf("offset %d did not toggle topic %d", offset, idx)
		}
	}
}
