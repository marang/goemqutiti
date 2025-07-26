package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func chipCoords(m *model, idx int) (int, int) {
	width := m.width - 4
	curX, curY := 0, 0
	for i, t := range m.topics {
		chip := chipStyle.Render(t.title)
		if !t.active {
			chip = chipInactive.Render(t.title)
		}
		w := lipgloss.Width(chip)
		if curX+w > width && curX > 0 {
			curY++
			curX = 0
		}
		if i == idx {
			return curX, curY
		}
		curX += w
	}
	return -1, -1
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
	start := m.elemPos["topics"] + 1
	_, _ = m.Update(tea.MouseMsg{Type: tea.MouseLeft, X: x + 2, Y: y + start})
	if m.selectedTopic != 0 {
		t.Fatalf("expected selected topic 0, got %d", m.selectedTopic)
	}
	if m.topics[0].active {
		t.Fatalf("topic 0 not toggled")
	}
}

func TestMouseToggleThirdRowTopic(t *testing.T) {
	m := initialModel(nil)
	m.Update(tea.WindowSizeMsg{Width: 40, Height: 20})
	setupTopics(m)
	m.viewClient()
	// topic index 6 resides on third row
	x, y := chipCoords(m, 6)
	start := m.elemPos["topics"] + 1
	_, _ = m.Update(tea.MouseMsg{Type: tea.MouseLeft, X: x + 2, Y: y + start})
	if m.selectedTopic != 6 {
		t.Fatalf("expected selected topic 6, got %d", m.selectedTopic)
	}
	if m.topics[6].active {
		t.Fatalf("topic 6 not toggled")
	}
}
