package emqutiti

import (
	"fmt"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/marang/emqutiti/ui"

	"github.com/marang/emqutiti/history"
)

func TestHandleMouseScrollTopics(t *testing.T) {
	m, _ := initialModel(nil)
	m.Update(tea.WindowSizeMsg{Width: 40, Height: 20})
	setupManyTopics(m, 10)
	m.layout.topics.height = 2
	m.viewClient()
	m.setFocus(idTopics)
	rowH := lipgloss.Height(ui.ChipStyle.Render("t"))
	_, handled := m.handleMouseScroll(tea.MouseMsg{Action: tea.MouseActionPress, Button: tea.MouseButtonWheelDown})
	if !handled {
		t.Fatalf("expected scroll event handled")
	}
	if m.topics.vp.YOffset != rowH {
		t.Fatalf("expected scroll %d got %d", rowH, m.topics.vp.YOffset)
	}
}

func TestHandleHistorySelectionShift(t *testing.T) {
	m, _ := initialModel(nil)
	m.history.items = []history.Item{
		{Timestamp: time.Now(), Topic: "t1", Payload: "p1", Kind: "pub"},
		{Timestamp: time.Now(), Topic: "t2", Payload: "p2", Kind: "pub"},
		{Timestamp: time.Now(), Topic: "t3", Payload: "p3", Kind: "pub"},
	}
	items := make([]list.Item, len(m.history.items))
	for i, it := range m.history.items {
		items[i] = it
	}
	m.history.list.SetItems(items)
	m.setFocus(idHistory)

	m.history.HandleSelection(0, true)
	if m.history.selectionAnchor != 0 {
		t.Fatalf("anchor = %d, want 0", m.history.selectionAnchor)
	}
	m.history.HandleSelection(2, true)
	for i := 0; i <= 2; i++ {
		if m.history.items[i].IsSelected == nil || !*m.history.items[i].IsSelected {
			t.Fatalf("item %d not selected", i)
		}
	}
	if m.history.selectionAnchor != 0 {
		t.Fatalf("anchor = %d, want 0", m.history.selectionAnchor)
	}
}

func TestFilterHistoryList(t *testing.T) {
	m, _ := initialModel(nil)
	hs := &historyStore{}
	m.history.store = hs
	ts := time.Now()
	hs.Append(history.Message{Timestamp: ts, Topic: "foo", Payload: "hello", Kind: "pub"})
	hs.Append(history.Message{Timestamp: ts, Topic: "bar", Payload: "bye", Kind: "pub"})

	m.history.list.SetFilteringEnabled(true)
	m.history.list.SetFilterText("topic=foo")
	m.history.list.SetFilterState(list.Filtering)
	m.filterHistoryList()

	items := m.history.list.Items()
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	hi := items[0].(history.Item)
	if hi.Topic != "foo" {
		t.Fatalf("unexpected topic %q", hi.Topic)
	}
}

func TestHandleHistoryClick(t *testing.T) {
	m, _ := initialModel(nil)
	m.Update(tea.WindowSizeMsg{Width: 40, Height: 20})
	m.history.items = []history.Item{{Timestamp: time.Now(), Topic: "t1", Payload: "p1", Kind: "pub"}}
	items := []list.Item{m.history.items[0]}
	m.history.list.SetItems(items)
	m.viewClient()
	m.setFocus(idHistory)
	y := m.ui.elemPos[idHistory] + 1
	m.history.HandleClick(tea.MouseMsg{Y: y}, m.ui.elemPos[idHistory], m.ui.viewport.YOffset)
	if m.history.list.Index() != 0 {
		t.Fatalf("expected index 0 got %d", m.history.list.Index())
	}
}

func TestHistoryScroll(t *testing.T) {
	m, _ := initialModel(nil)
	for i := 0; i < 30; i++ {
		hi := history.Item{Timestamp: time.Now(), Topic: fmt.Sprintf("t%d", i), Payload: "p", Kind: "pub"}
		m.history.items = append(m.history.items, hi)
	}
	items := make([]list.Item, len(m.history.items))
	for i, it := range m.history.items {
		items[i] = it
	}
	m.history.list.SetItems(items)
	m.setFocus(idHistory)
	_, handled := m.handleMouseScroll(tea.MouseMsg{Action: tea.MouseActionPress, Button: tea.MouseButtonWheelDown})
	if !handled {
		t.Fatalf("expected scroll event handled")
	}
}

func TestUpdateClientStatus(t *testing.T) {
	m, _ := initialModel(nil)
	cmds := m.updateClientStatus()
	if len(cmds) != 1 {
		t.Fatalf("expected 1 cmd got %d", len(cmds))
	}
	m.mqttClient = &MQTTClient{MessageChan: make(chan MQTTMessage)}
	cmds = m.updateClientStatus()
	if len(cmds) != 2 {
		t.Fatalf("expected 2 cmds got %d", len(cmds))
	}
}
