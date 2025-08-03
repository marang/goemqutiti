package emqutiti

import (
	"testing"
	"time"
)

// Test that HistoryStore.Search respects topic, time, and payload filters for
// active and archived messages.
func TestHistoryStoreSearch(t *testing.T) {
	now := time.Now()

	t.Run("active", func(t *testing.T) {
		hs := &historyStore{}
		hs.Append(Message{Timestamp: now.Add(-30 * time.Minute), Topic: "a", Payload: "foo", Kind: "pub"})
		hs.Append(Message{Timestamp: now.Add(-2 * time.Hour), Topic: "b", Payload: "bar", Kind: "pub"})

		res := hs.Search(false, []string{"a"}, now.Add(-1*time.Hour), now, "")
		if len(res) != 1 || res[0].Topic != "a" {
			t.Fatalf("topic filter failed: %#v", res)
		}

		res = hs.Search(false, nil, now.Add(-1*time.Hour), now, "foo")
		if len(res) != 1 || res[0].Payload != "foo" {
			t.Fatalf("payload filter failed: %#v", res)
		}

		res = hs.Search(false, []string{"b"}, now.Add(-1*time.Hour), now, "")
		if len(res) != 0 {
			t.Fatalf("time filter failed: %#v", res)
		}
	})

	t.Run("archived", func(t *testing.T) {
		hs := &historyStore{}
		hs.Append(Message{Timestamp: now.Add(-30 * time.Minute), Topic: "a", Payload: "foo", Kind: "pub", Archived: true})
		hs.Append(Message{Timestamp: now.Add(-2 * time.Hour), Topic: "b", Payload: "bar", Kind: "pub", Archived: true})

		res := hs.Search(true, []string{"a"}, now.Add(-1*time.Hour), now, "")
		if len(res) != 1 || res[0].Topic != "a" {
			t.Fatalf("topic filter failed: %#v", res)
		}

		res = hs.Search(true, nil, now.Add(-1*time.Hour), now, "foo")
		if len(res) != 1 || res[0].Payload != "foo" {
			t.Fatalf("payload filter failed: %#v", res)
		}

		res = hs.Search(true, []string{"b"}, now.Add(-1*time.Hour), now, "")
		if len(res) != 0 {
			t.Fatalf("time filter failed: %#v", res)
		}
	})
}
