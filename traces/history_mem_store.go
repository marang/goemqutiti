package traces

import (
	"strings"
	"time"

	"github.com/marang/emqutiti/history"
)

type memStore struct {
	msgs []history.Message
}

func newMemStore(msgs []history.Message) *memStore {
	return &memStore{msgs: msgs}
}

func (m *memStore) Append(history.Message) error { return nil }

func (m *memStore) Search(archived bool, topics []string, start, end time.Time, payload string) []history.Message {
	var out []history.Message
	topicSet := map[string]struct{}{}
	for _, t := range topics {
		if t != "" {
			topicSet[t] = struct{}{}
		}
	}
	for _, msg := range m.msgs {
		if msg.Archived != archived {
			continue
		}
		if len(topicSet) > 0 {
			if _, ok := topicSet[msg.Topic]; !ok {
				continue
			}
		}
		if !start.IsZero() && msg.Timestamp.Before(start) {
			continue
		}
		if !end.IsZero() && msg.Timestamp.After(end) {
			continue
		}
		if payload != "" && !strings.Contains(msg.Payload, payload) {
			continue
		}
		out = append(out, msg)
	}
	return out
}

func (m *memStore) Delete(string) error { return nil }

func (m *memStore) Archive(string) error { return nil }

func (m *memStore) Count(archived bool) int {
	c := 0
	for _, msg := range m.msgs {
		if msg.Archived == archived {
			c++
		}
	}
	return c
}

func (m *memStore) Close() error { return nil }

var _ history.Store = (*memStore)(nil)
