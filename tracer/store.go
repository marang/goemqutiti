package tracer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dgraph-io/badger/v4"
)

func dataDir(profile string) string {
	if profile == "" {
		profile = "default"
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("data", profile)
	}
	return filepath.Join(home, ".emqutiti", "data", profile)
}

func traceDir(profile string) string {
	return filepath.Join(dataDir(profile), "traces")
}

func openTraceDB(profile string, readonly bool) (*badger.DB, error) {
	if profile == "" {
		profile = "default"
	}
	path := traceDir(profile)
	os.MkdirAll(path, 0755)
	opts := badger.DefaultOptions(path).WithLogger(nil)
	if readonly {
		opts = opts.WithReadOnly(true)
	}
	return badger.Open(opts)
}

// Add stores a trace message under the given key.
func Add(profile, key string, msg Message) error {
	db, err := openTraceDB(profile, false)
	if err != nil {
		return err
	}
	defer db.Close()

	dbKey := []byte(fmt.Sprintf("trace/%s/%s/%020d", key, msg.Topic, msg.Timestamp.UnixNano()))
	val, _ := json.Marshal(msg)
	return db.Update(func(txn *badger.Txn) error {
		return txn.Set(dbKey, val)
	})
}

// Messages returns all messages stored for the given trace key.
func Messages(profile, key string) ([]Message, error) {
	db, err := openTraceDB(profile, true)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	prefix := []byte(fmt.Sprintf("trace/%s/", key))
	var msgs []Message
	err = db.View(func(txn *badger.Txn) error {
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

// Keys returns all database keys for the given trace key.
func Keys(profile, key string) ([]string, error) {
	db, err := openTraceDB(profile, true)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	prefix := []byte(fmt.Sprintf("trace/%s/", key))
	var out []string
	err = db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			out = append(out, string(it.Item().Key()))
		}
		return nil
	})
	return out, err
}

// Delete removes all stored messages for the given trace key.
func Delete(profile, key string) error {
	db, err := openTraceDB(profile, false)
	if err != nil {
		return err
	}
	defer db.Close()

	prefix := []byte(fmt.Sprintf("trace/%s/", key))
	return db.Update(func(txn *badger.Txn) error {
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

// HasData reports whether any messages are stored for the given trace key.
func HasData(profile, key string) (bool, error) {
	keys, err := Keys(profile, key)
	if err != nil {
		return false, err
	}
	return len(keys) > 0, nil
}

// ClearData deletes all messages stored for the trace key.
func ClearData(profile, key string) error {
	return Delete(profile, key)
}
