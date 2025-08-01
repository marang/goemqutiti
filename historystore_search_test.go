package main

import (
	"testing"
	"time"
)

// Test that HistoryStore.Search respects topic, time, and payload filters.
func TestHistoryStoreSearch(t *testing.T) {
	hs := &HistoryStore{}
	now := time.Now()
	hs.Add(Message{Timestamp: now.Add(-30 * time.Minute), Topic: "a", Payload: "foo", Kind: "pub"})
	hs.Add(Message{Timestamp: now.Add(-2 * time.Hour), Topic: "b", Payload: "bar", Kind: "pub"})

	res := hs.Search([]string{"a"}, now.Add(-1*time.Hour), now, "")
	if len(res) != 1 || res[0].Topic != "a" {
		t.Fatalf("topic filter failed: %#v", res)
	}

	res = hs.Search(nil, now.Add(-1*time.Hour), now, "foo")
	if len(res) != 1 || res[0].Payload != "foo" {
		t.Fatalf("payload filter failed: %#v", res)
	}

	res = hs.Search([]string{"b"}, now.Add(-1*time.Hour), now, "")
	if len(res) != 0 {
		t.Fatalf("time filter failed: %#v", res)
	}
}
