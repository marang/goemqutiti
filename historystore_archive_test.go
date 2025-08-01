package main

import (
	"fmt"
	"testing"
	"time"
)

func TestArchiveAndSearchArchived(t *testing.T) {
	hs := &HistoryStore{}
	ts := time.Now()
	msg := Message{Timestamp: ts, Topic: "t1", Payload: "p1", Kind: "pub"}
	hs.Add(msg)
	key := fmt.Sprintf("%s/%020d", msg.Topic, msg.Timestamp.UnixNano())
	if err := hs.Archive(key); err != nil {
		t.Fatalf("Archive failed: %v", err)
	}
	if res := hs.Search([]string{"t1"}, time.Time{}, time.Time{}, ""); len(res) != 0 {
		t.Fatalf("expected no active messages, got %d", len(res))
	}
	res := hs.SearchArchived([]string{"t1"}, time.Time{}, time.Time{}, "")
	if len(res) != 1 {
		t.Fatalf("expected 1 archived message, got %d", len(res))
	}
	if !res[0].Archived {
		t.Fatalf("archived flag not set")
	}
}
