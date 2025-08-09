package history

import (
	"testing"
	"time"
)

// TestMessagesToItems verifies conversion from Message slices to history items.
func TestMessagesToItems(t *testing.T) {
	msgs := []Message{
		{Timestamp: time.Unix(0, 1), Topic: "t1", Payload: "p1", Kind: "pub", Archived: false, Retained: true},
		{Timestamp: time.Unix(0, 2), Topic: "t2", Payload: "p2", Kind: "sub", Archived: true, Retained: false},
	}
	hitems, litems := MessagesToItems(msgs)
	if len(hitems) != len(msgs) {
		t.Fatalf("history items len %d want %d", len(hitems), len(msgs))
	}
	if len(litems) != len(msgs) {
		t.Fatalf("list items len %d want %d", len(litems), len(msgs))
	}
	for i, hi := range hitems {
		m := msgs[i]
		if hi.Timestamp != m.Timestamp || hi.Topic != m.Topic || hi.Payload != m.Payload || hi.Kind != m.Kind || hi.Archived != m.Archived || hi.Retained != m.Retained {
			t.Fatalf("item %d mismatch: %#v vs %#v", i, hi, m)
		}
		if li, ok := litems[i].(Item); ok {
			if li != hi {
				t.Fatalf("list item %d mismatch: %#v vs %#v", i, li, hi)
			}
		} else {
			t.Fatalf("list item %d has unexpected type %T", i, litems[i])
		}
	}
}
