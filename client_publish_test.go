package emqutiti

import (
	"testing"

	"github.com/marang/emqutiti/topics"
)

func TestHandlePublishKeyFlags(t *testing.T) {
	m, _ := initialModel(nil)
	m.topics.Items = []topics.Item{
		{Name: "a", Publish: true},
		{Name: "b"},
		{Name: "c", Publish: true},
	}
	m.message.SetPayload("hello")
	m.SetFocus(idMessage)
	m.handlePublishKey()
	items := m.payloads.Items()
	if len(items) != 2 {
		t.Fatalf("expected 2 payloads, got %d", len(items))
	}
	if items[0].Topic != "a" || items[1].Topic != "c" {
		t.Fatalf("unexpected topics: %+v", items)
	}
}

func TestHandlePublishKeyFallback(t *testing.T) {
	m, _ := initialModel(nil)
	m.topics.Items = []topics.Item{
		{Name: "a"},
		{Name: "b"},
	}
	m.topics.SetSelected(1)
	m.message.SetPayload("hi")
	m.SetFocus(idMessage)
	m.handlePublishKey()
	items := m.payloads.Items()
	if len(items) != 1 {
		t.Fatalf("expected 1 payload, got %d", len(items))
	}
	if items[0].Topic != "b" {
		t.Fatalf("expected topic 'b', got %q", items[0].Topic)
	}
}
