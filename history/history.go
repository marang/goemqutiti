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

func baseDir(profile string) string {
	if profile == "" {
		profile = "default"
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("data", profile)
	}
	return filepath.Join(home, ".emqutiti", "data", profile)
}

// HistoryDir returns the directory for the history database.
func HistoryDir(profile string) string {
	return filepath.Join(baseDir(profile), "history")
}

// TraceDir returns the directory for the trace database.
func TraceDir(profile string) string {
	return filepath.Join(baseDir(profile), "traces")
}

// DefaultDir is kept for backward compatibility and returns HistoryDir.
func DefaultDir(profile string) string { return HistoryDir(profile) }

// Open opens (or creates) a persistent message index for the given profile.
// If profile is empty, "default" is used.
func Open(profile string) (*Index, error) {
	if profile == "" {
		profile = "default"
	}
	path := HistoryDir(profile)
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

// OpenTrace opens the trace database for the given profile.
func OpenTrace(profile string) (*Index, error) {
	if profile == "" {
		profile = "default"
	}
	path := TraceDir(profile)
	os.MkdirAll(path, 0755)
	opts := badger.DefaultOptions(path).WithLogger(nil)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	idx := &Index{db: db}
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

// AddTrace stores a trace message under the given key without adding it to the
// in-memory list.
func (i *Index) AddTrace(key string, msg Message) {
	if i.db == nil {
		return
	}
	dbKey := []byte(fmt.Sprintf("trace/%s/%s/%020d", key, msg.Topic, msg.Timestamp.UnixNano()))
	val, _ := json.Marshal(msg)
	i.db.Update(func(txn *badger.Txn) error {
		return txn.Set(dbKey, val)
	})
}

// TraceMessages returns all messages stored for the given trace key.
func (i *Index) TraceMessages(key string) ([]Message, error) {
	if i.db == nil {
		return nil, nil
	}
	prefix := []byte(fmt.Sprintf("trace/%s/", key))
	var msgs []Message
	err := i.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			if err := item.Value(func(val []byte) error {
				var m Message
				if err := json.Unmarshal(val, &m); err != nil {
					return err
				}
				msgs = append(msgs, m)
				return nil
			}); err != nil {
				return err
			}
		}
		return nil
	})
	return msgs, err
}

// TraceKeys returns all database keys for the given trace key.
func (i *Index) TraceKeys(key string) ([]string, error) {
	if i.db == nil {
		return nil, nil
	}
	prefix := []byte(fmt.Sprintf("trace/%s/", key))
	var out []string
	err := i.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			out = append(out, string(it.Item().Key()))
		}
		return nil
	})
	return out, err
}

// DeleteTrace removes all stored messages for the given trace key.
func (i *Index) DeleteTrace(key string) error {
	if i.db == nil {
		return nil
	}
	prefix := []byte(fmt.Sprintf("trace/%s/", key))
	return i.db.Update(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			if err := txn.Delete(it.Item().KeyCopy(nil)); err != nil {
				return err
			}
		}
		return nil
	})
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
