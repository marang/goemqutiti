package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
)

// Message holds a timestamped MQTT message with optional payload text.
type Message struct {
	Timestamp time.Time
	Topic     string
	Payload   string
	Kind      string
}

// Index stores messages in memory and supports filtered searches.
type Index struct {
	mu   sync.RWMutex
	msgs []Message
	db   *badger.DB
}

// DefaultDir returns the default directory for the history database.
func DefaultDir(profile string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("history", profile)
	}
	return filepath.Join(home, ".emqutiti", "history", profile)
}

// Open opens (or creates) a persistent message index for the given profile.
// If profile is empty, "default" is used.
func Open(profile string) (*Index, error) {
	if profile == "" {
		profile = "default"
	}
	path := DefaultDir(profile)
	os.MkdirAll(path, 0755)
	opts := badger.DefaultOptions(path).WithLogger(nil)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	idx := &Index{db: db}
	// Load existing messages
	db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			item.Value(func(val []byte) error {
				var m Message
				json.Unmarshal(val, &m)
				idx.msgs = append(idx.msgs, m)
				return nil
			})
		}
		return nil
	})
	return idx, nil
}

// Close closes the underlying database.
func (i *Index) Close() error {
	if i.db != nil {
		return i.db.Close()
	}
	return nil
}

// Add appends a message to the index.
func (i *Index) Add(msg Message) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.msgs = append(i.msgs, msg)
	if i.db != nil {
		// Keys use the format <topic>/<timestamp>. Slashes in the topic are
		// safe because BadgerDB treats keys as byte strings and doesn't
		// interpret '/' as a directory separator. Keeping slashes allows
		// prefix queries on hierarchical topics.
		key := []byte(fmt.Sprintf("%s/%020d", msg.Topic, msg.Timestamp.UnixNano()))
		val, _ := json.Marshal(msg)
		i.db.Update(func(txn *badger.Txn) error {
			return txn.Set(key, val)
		})
	}
}

// Search returns all messages matching the provided filters. Zero timestamps
// disable the corresponding time constraints.
func (i *Index) Search(topics []string, start, end time.Time, payload string) []Message {
	i.mu.RLock()
	defer i.mu.RUnlock()

	var out []Message
	topicSet := map[string]struct{}{}
	for _, t := range topics {
		if t == "" {
			continue
		}
		topicSet[t] = struct{}{}
	}

	for _, m := range i.msgs {
		if len(topicSet) > 0 {
			if _, ok := topicSet[m.Topic]; !ok {
				continue
			}
		}
		if !start.IsZero() && m.Timestamp.Before(start) {
			continue
		}
		if !end.IsZero() && m.Timestamp.After(end) {
			continue
		}
		if payload != "" && !strings.Contains(m.Payload, payload) {
			continue
		}
		out = append(out, m)
	}
	return out
}

// ParseQuery interprets a filter string in the form:
//
//	"topic=a,b start=2023-01-02T15:04:05Z end=2023-01-02T16:00 payload=foo".
//
// Fields may appear in any order and are optional. Unrecognised tokens are
// treated as payload search text.
func ParseQuery(q string) (topics []string, start, end time.Time, payload string) {
	var payloadParts []string
	for _, f := range strings.Fields(q) {
		switch {
		case strings.HasPrefix(f, "topic="):
			ts := strings.TrimPrefix(f, "topic=")
			if ts != "" {
				topics = strings.Split(ts, ",")
			}
		case strings.HasPrefix(f, "start="):
			t, err := time.Parse(time.RFC3339, strings.TrimPrefix(f, "start="))
			if err == nil {
				start = t
			}
		case strings.HasPrefix(f, "end="):
			t, err := time.Parse(time.RFC3339, strings.TrimPrefix(f, "end="))
			if err == nil {
				end = t
			}
		case strings.HasPrefix(f, "payload="):
			payloadParts = append(payloadParts, strings.TrimPrefix(f, "payload="))
		default:
			payloadParts = append(payloadParts, f)
		}
	}
	payload = strings.Join(payloadParts, " ")
	return
}
