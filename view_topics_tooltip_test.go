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
	m.viewClient()
	tip := m.topicTooltip()
	plain := strings.ReplaceAll(ansi.Strip(tip), "\n", "")
	if !strings.Contains(plain, longName[:10]) {
		t.Fatalf("expected tooltip to include full topic name")
	}
}

func TestTopicTooltipShort(t *testing.T) {
	m, _ := initialModel(nil)
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 10})
	m.topics.Items = []topics.Item{{Name: "short", Subscribed: true}}
	m.topics.SetSelected(0)
	m.viewClient()
	tip := m.topicTooltip()
	if tip != "" {
		t.Fatalf("expected no tooltip for short topic")
	}
}
