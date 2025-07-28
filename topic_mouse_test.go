package main

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"goemqutiti/ui"
)

func chipCoords(m *model, idx int) (int, int) {
	if idx < 0 || idx >= len(m.topics.chipBounds) {
		return -1, -1
	}
	b := m.topics.chipBounds[idx]
	return b.x, b.y - m.viewport.YOffset
}

func setupTopics(m *model) {
	names := []string{"testtopic", "asdfsedf", "asdasd", "sdfdfasssssd", "asdasdasss", "asasasdfffa", "asasdfa", "aasdf", "asdfa", "asdasasdfasdf"}
	for _, n := range names {
		m.topics.items = append(m.topics.items, topicItem{title: n, active: true})
	}
	m.layout.topics.height = len(names)
}

func TestMouseToggleFirstTopic(t *testing.T) {
	m := initialModel(nil)
	m.Update(tea.WindowSizeMsg{Width: 40, Height: 20})
	setupTopics(m)
	m.viewClient()
	x, y := chipCoords(m, 0)
	name := m.topics.items[0].title
	for offset := 0; offset < m.topics.chipBounds[0].h; offset++ {
		before := m.topics.items[0].active
		m.Update(tea.MouseMsg{Type: tea.MouseLeft, X: x, Y: y + offset})
		idx := -1
		for i, tpc := range m.topics.items {
			if tpc.title == name {
				idx = i
				break
			}
		}
		if idx >= 0 && m.topics.items[idx].active == before {
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
	name := m.topics.items[6].title
	for offset := 0; offset < m.topics.chipBounds[6].h; offset++ {
		before := m.topics.items[6].active
		m.Update(tea.MouseMsg{Type: tea.MouseLeft, X: x, Y: y + offset})
		idx := -1
		for i, tpc := range m.topics.items {
			if tpc.title == name {
				idx = i
				break
			}
		}
		if idx >= 0 && m.topics.items[idx].active == before {
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
	name := m.topics.items[8].title
	for offset := 0; offset < m.topics.chipBounds[8].h; offset++ {
		before := m.topics.items[8].active
		m.Update(tea.MouseMsg{Type: tea.MouseLeft, X: x, Y: y + offset})
		idx := -1
		for i, tpc := range m.topics.items {
			if tpc.title == name {
				idx = i
				break
			}
		}
		if idx >= 0 && m.topics.items[idx].active == before {
			t.Fatalf("offset %d did not toggle topic 8", offset)
		}
	}
}

func setupManyTopics(m *model, n int) {
	for i := 0; i < n; i++ {
		title := fmt.Sprintf("topic-%d", i)
		m.topics.items = append(m.topics.items, topicItem{title: title, active: true})
	}
	m.layout.topics.height = n
}

func TestMouseToggleFifteenthRowTopic(t *testing.T) {
	t.Skip("layout changed")
	m := initialModel(nil)
	// Enough height for many rows
	m.Update(tea.WindowSizeMsg{Width: 40, Height: 80})
	setupManyTopics(m, 50)
	m.viewClient()
	// Index of the first chip on the 15th row (0-based rows, 3 chips per row)
	idx := 14 * 3
	x, y := chipCoords(m, idx)
	name := m.topics.items[idx].title
	for offset := 0; offset < m.topics.chipBounds[idx].h; offset++ {
		before := false
		for _, tpc := range m.topics.items {
			if tpc.title == name {
				before = tpc.active
				break
			}
		}
		m.Update(tea.MouseMsg{Type: tea.MouseLeft, X: x, Y: y + offset})
		after := false
		for _, tpc := range m.topics.items {
			if tpc.title == name {
				after = tpc.active
				break
			}
		}
		if before == after {
			t.Fatalf("offset %d did not toggle topic %d", offset, idx)
		}
	}
}

func TestMouseToggleWithScroll(t *testing.T) {
	t.Skip("layout changed")
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
	name := m.topics.items[idx].title
	for offset := 0; offset < m.topics.chipBounds[idx].h; offset++ {
		before := false
		for _, tpc := range m.topics.items {
			if tpc.title == name {
				before = tpc.active
				break
			}
		}
		m.Update(tea.MouseMsg{Type: tea.MouseLeft, X: x, Y: y + offset})
		after := false
		for _, tpc := range m.topics.items {
			if tpc.title == name {
				after = tpc.active
				break
			}
		}
		if before == after {
			t.Fatalf("offset %d did not toggle topic %d", offset, idx)
		}
	}
}
