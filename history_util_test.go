package main

import (
	"testing"
	"time"
)

// TestMessagesToHistoryItems verifies conversion from Message slices to history items.
func TestMessagesToHistoryItems(t *testing.T) {
	msgs := []Message{
		{Timestamp: time.Unix(0, 1), Topic: "t1", Payload: "p1", Kind: "pub", Archived: false},
		{Timestamp: time.Unix(0, 2), Topic: "t2", Payload: "p2", Kind: "sub", Archived: true},
	}
	hitems, litems := messagesToHistoryItems(msgs)
	if len(hitems) != len(msgs) {
		t.Fatalf("history items len %d want %d", len(hitems), len(msgs))
	}
	if len(litems) != len(msgs) {
		t.Fatalf("list items len %d want %d", len(litems), len(msgs))
	}
	for i, hi := range hitems {
		m := msgs[i]
		if hi.timestamp != m.Timestamp || hi.topic != m.Topic || hi.payload != m.Payload || hi.kind != m.Kind || hi.archived != m.Archived {
			t.Fatalf("item %d mismatch: %#v vs %#v", i, hi, m)
		}
		if li, ok := litems[i].(historyItem); ok {
			if li != hi {
				t.Fatalf("list item %d mismatch: %#v vs %#v", i, li, hi)
			}
		} else {
			t.Fatalf("list item %d has unexpected type %T", i, litems[i])
		}
	}
}
