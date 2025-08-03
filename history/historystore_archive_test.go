package history

import (
	"fmt"
	"testing"
	"time"
)

func TestArchiveAndSearch(t *testing.T) {
	hs := &store{}
	ts := time.Now()
	msg := Message{Timestamp: ts, Topic: "t1", Payload: "p1", Kind: "pub"}
	hs.Append(msg)
	key := fmt.Sprintf("%s/%020d", msg.Topic, msg.Timestamp.UnixNano())
	if err := hs.Archive(key); err != nil {
		t.Fatalf("Archive failed: %v", err)
	}
	if res := hs.Search(false, []string{"t1"}, time.Time{}, time.Time{}, ""); len(res) != 0 {
		t.Fatalf("expected no active messages, got %d", len(res))
	}
	res := hs.Search(true, []string{"t1"}, time.Time{}, time.Time{}, "")
	if len(res) != 1 {
		t.Fatalf("expected 1 archived message, got %d", len(res))
	}
	if !res[0].Archived {
		t.Fatalf("archived flag not set")
	}
}
