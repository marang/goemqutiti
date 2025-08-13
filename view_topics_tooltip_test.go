package emqutiti

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/marang/emqutiti/topics"
)

func TestTopicTooltipLong(t *testing.T) {
	m, _ := initialModel(nil)
	m.Update(tea.WindowSizeMsg{Width: 20, Height: 10})
	longName := strings.Repeat("x", 50)
	m.topics.Items = []topics.Item{{Name: longName, Subscribed: true}}
	m.topics.SetSelected(0)
	view := m.viewClient()
	if !strings.Contains(m.topics.VP.View(), "…") {
		t.Fatalf("expected truncated chip")
	}
	tip, _, _ := m.topicTooltip()
	plainTip := ansi.Strip(tip)
	if strings.Count(plainTip, "x") != len(longName) {
		t.Fatalf("expected tooltip to include full topic name")
	}
	plain := ansi.Strip(view)
	idx := strings.Index(plain, longName[:10])
	if idx < 0 {
		t.Fatalf("expected view to include tooltip text")
	}
	line := strings.Count(plain[:idx], "\n")
	if line >= m.ui.height {
		t.Fatalf("tooltip rendered outside viewport")
	}
}

func TestTopicTooltipLongWide(t *testing.T) {
	m, _ := initialModel(nil)
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 10})
	longName := strings.Repeat("x", maxTopicChipWidth+10)
	m.topics.Items = []topics.Item{{Name: longName, Subscribed: true}}
	m.topics.SetSelected(0)
	view := m.viewClient()
	if !strings.Contains(m.topics.VP.View(), "…") {
		t.Fatalf("expected truncated chip")
	}
	tip, _, _ := m.topicTooltip()
	plainTip := ansi.Strip(tip)
	if strings.Count(plainTip, "x") != len(longName) {
		t.Fatalf("expected tooltip to include full topic name")
	}
	plain := ansi.Strip(view)
	idx := strings.Index(plain, longName[:10])
	if idx < 0 {
		t.Fatalf("expected view to include tooltip text")
	}
	line := strings.Count(plain[:idx], "\n")
	if line >= m.ui.height {
		t.Fatalf("tooltip rendered outside viewport")
	}
}

func TestTopicTooltipShort(t *testing.T) {
	m, _ := initialModel(nil)
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 10})
	m.topics.Items = []topics.Item{{Name: "short", Subscribed: true}}
	m.topics.SetSelected(0)
	m.viewClient()
	if strings.Contains(m.topics.VP.View(), "…") {
		t.Fatalf("expected no truncation for short topic")
	}
	tip, _, _ := m.topicTooltip()
	if tip != "" {
		t.Fatalf("expected no tooltip for short topic")
	}
}
