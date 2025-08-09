package history

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"

	"github.com/marang/emqutiti/internal/files"
)

// Message holds a timestamped MQTT message with optional payload text.
type Message struct {
	Timestamp time.Time
	Topic     string
	Payload   string
	Kind      string
	Archived  bool
	Retained  bool
}

// store stores messages in memory and optionally persists them to disk.
type store struct {
	mu   sync.RWMutex
	msgs []Message
	db   *badger.DB
}

// openStore opens (or creates) a persistent message index for the given profile.
// If profile is empty, "default" is used.
func openStore(profile string) (Store, error) {
	if profile == "" {
		profile = "default"
	}
	path := filepath.Join(files.DataDir(profile), "history")
	if err := files.EnsureDir(path); err != nil {
		return nil, err
	}
	opts := badger.DefaultOptions(path).WithLogger(nil)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	idx := &store{db: db}
	// Load existing messages
	if err := db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			var m Message
			if err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &m)
			}); err != nil {
				return err
			}
			idx.msgs = append(idx.msgs, m)
		}
		return nil
	}); err != nil {
		db.Close()
		return nil, err
	}
	return idx, nil
}

// Close closes the underlying database.
func (i *store) Close() error {
	if i.db != nil {
		return i.db.Close()
	}
	return nil
}

// Append adds a message to the store.
func (i *store) Append(msg Message) error {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.msgs = append(i.msgs, msg)
	if i.db != nil {
		// Keys use the format <topic>/<timestamp>. Slashes in the topic are
		// safe because BadgerDB treats keys as byte strings and doesn't
		// interpret '/' as a directory separator. Keeping slashes allows
		// prefix queries on hierarchical topics.
		key := []byte(fmt.Sprintf("%s/%020d", msg.Topic, msg.Timestamp.UnixNano()))
		val, err := json.Marshal(msg)
		if err != nil {
			return err
		}
		if err := i.db.Update(func(txn *badger.Txn) error {
			return txn.Set(key, val)
		}); err != nil {
			return err
		}
	}
	return nil
}

// Delete removes a message with the given key from the index.
// The key should use the format "<topic>/<timestamp>" matching Add.
func (i *store) Delete(key string) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.db != nil {
		if err := i.db.Update(func(txn *badger.Txn) error {
			return txn.Delete([]byte(key))
		}); err != nil {
			return err
		}
	}

	for idx, m := range i.msgs {
		k := fmt.Sprintf("%s/%020d", m.Topic, m.Timestamp.UnixNano())
		if k == key {
			i.msgs = append(i.msgs[:idx], i.msgs[idx+1:]...)
			break
		}
	}
	return nil
}

// Archive marks a message as archived without deleting it.
// The key should use the format "<topic>/<timestamp>" matching Add.
func (i *store) Archive(key string) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	for idx, m := range i.msgs {
		k := fmt.Sprintf("%s/%020d", m.Topic, m.Timestamp.UnixNano())
		if k == key {
			if i.db != nil {
				m.Archived = true
				val, err := json.Marshal(m)
				if err != nil {
					return err
				}
				if err := i.db.Update(func(txn *badger.Txn) error {
					return txn.Set([]byte(key), val)
				}); err != nil {
					return err
				}
				i.msgs[idx] = m
			} else {
				m.Archived = true
				i.msgs[idx] = m
			}
			return nil
		}
	}
	return fmt.Errorf("message %s not found", key)
}

// Search returns messages matching the provided filters. Zero timestamps
// disable the corresponding time constraints. When archived is true, only
// archived messages are returned.
func (i *store) Search(archived bool, topics []string, start, end time.Time, payload string) []Message {
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
		if m.Archived != archived {
			continue
		}
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

// Count reports the number of stored messages. When archived is true,
// only archived messages are counted; otherwise only unarchived messages
// are included.
func (i *store) Count(archived bool) int {
	i.mu.RLock()
	defer i.mu.RUnlock()
	c := 0
	for _, m := range i.msgs {
		if m.Archived == archived {
			c++
		}
	}
	return c
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
