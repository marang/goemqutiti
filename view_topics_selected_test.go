package emqutiti

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/marang/emqutiti/topics"
)

func TestSelectedTopicInfoLong(t *testing.T) {
	m, _ := initialModel(nil)
	m.Update(tea.WindowSizeMsg{Width: 20, Height: 10})
	longName := strings.Repeat("x", 50)
	m.topics.Items = []topics.Item{{Name: longName, Subscribed: true}}
	m.topics.SetSelected(0)
	m.viewClient()
	if !strings.Contains(m.topics.VP.View(), "…") {
		t.Fatalf("expected truncated chip")
	}
	info := ansi.Strip(m.selectedTopicInfo())
	if strings.Count(info, "x") != len(longName) {
		t.Fatalf("expected info line to include full topic name")
	}
}

func TestSelectedTopicInfoLongWide(t *testing.T) {
	m, _ := initialModel(nil)
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 10})
	longName := strings.Repeat("x", maxTopicChipWidth+10)
	m.topics.Items = []topics.Item{{Name: longName, Subscribed: true}}
	m.topics.SetSelected(0)
	m.viewClient()
	if !strings.Contains(m.topics.VP.View(), "…") {
		t.Fatalf("expected truncated chip")
	}
	info := ansi.Strip(m.selectedTopicInfo())
	if strings.Count(info, "x") != len(longName) {
		t.Fatalf("expected info line to include full topic name")
	}
}

func TestSelectedTopicInfoShort(t *testing.T) {
	m, _ := initialModel(nil)
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 10})
	m.topics.Items = []topics.Item{{Name: "short", Subscribed: true}}
	m.topics.SetSelected(0)
	m.viewClient()
	if strings.Contains(m.topics.VP.View(), "…") {
		t.Fatalf("expected no truncation for short topic")
	}
	if m.selectedTopicInfo() != "" {
		t.Fatalf("expected no info line for short topic")
	}
}
