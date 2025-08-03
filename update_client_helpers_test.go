package emqutiti

import (
	"fmt"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/marang/emqutiti/ui"
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
	m.history.items = []historyItem{
		{timestamp: time.Now(), topic: "t1", payload: "p1", kind: "pub"},
		{timestamp: time.Now(), topic: "t2", payload: "p2", kind: "pub"},
		{timestamp: time.Now(), topic: "t3", payload: "p3", kind: "pub"},
	}
	items := make([]list.Item, len(m.history.items))
	for i, it := range m.history.items {
		items[i] = it
	}
	m.history.list.SetItems(items)
	m.setFocus(idHistory)

	m.handleHistorySelection(0, true)
	if m.history.selectionAnchor != 0 {
		t.Fatalf("anchor = %d, want 0", m.history.selectionAnchor)
	}
	m.handleHistorySelection(2, true)
	for i := 0; i <= 2; i++ {
		if m.history.items[i].isSelected == nil || !*m.history.items[i].isSelected {
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
	hs.Append(Message{Timestamp: ts, Topic: "foo", Payload: "hello", Kind: "pub"})
	hs.Append(Message{Timestamp: ts, Topic: "bar", Payload: "bye", Kind: "pub"})

	m.history.list.SetFilteringEnabled(true)
	m.history.list.SetFilterText("topic=foo")
	m.history.list.SetFilterState(list.Filtering)
	m.filterHistoryList()

	items := m.history.list.Items()
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	hi := items[0].(historyItem)
	if hi.topic != "foo" {
		t.Fatalf("unexpected topic %q", hi.topic)
	}
}

func TestHandleHistoryClick(t *testing.T) {
	m, _ := initialModel(nil)
	m.Update(tea.WindowSizeMsg{Width: 40, Height: 20})
	m.history.items = []historyItem{{timestamp: time.Now(), topic: "t1", payload: "p1", kind: "pub"}}
	items := []list.Item{m.history.items[0]}
	m.history.list.SetItems(items)
	m.viewClient()
	m.setFocus(idHistory)
	y := m.ui.elemPos[idHistory] + 1
	m.handleHistoryClick(tea.MouseMsg{Y: y})
	if m.history.list.Index() != 0 {
		t.Fatalf("expected index 0 got %d", m.history.list.Index())
	}
}

func TestHistoryScroll(t *testing.T) {
	m, _ := initialModel(nil)
	for i := 0; i < 30; i++ {
		hi := historyItem{timestamp: time.Now(), topic: fmt.Sprintf("t%d", i), payload: "p", kind: "pub"}
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
