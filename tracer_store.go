package emqutiti

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/dgraph-io/badger/v4"

	"github.com/marang/emqutiti/internal/files"
)

// openTracerStore opens the trace database for the profile.
// When readonly is true, the database is opened in read-only mode.
func openTracerStore(profile string, readonly bool) (*badger.DB, error) {
	if profile == "" {
		profile = "default"
	}
	path := filepath.Join(files.DataDir(profile), "traces")
	if err := files.EnsureDir(path); err != nil {
		return nil, err
	}
	opts := badger.DefaultOptions(path).WithLogger(nil)
	if readonly {
		opts = opts.WithReadOnly(true)
	}
	return badger.Open(opts)
}

// Add stores a trace message under the given key.
func tracerAdd(profile, key string, msg TracerMessage) error {
	db, err := openTracerStore(profile, false)
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
func tracerMessages(profile, key string) ([]TracerMessage, error) {
	db, err := openTracerStore(profile, true)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	prefix := []byte(fmt.Sprintf("trace/%s/", key))
	var msgs []TracerMessage
	err = db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			if err := item.Value(func(val []byte) error {
				var m TracerMessage
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
func tracerKeys(profile, key string) ([]string, error) {
	db, err := openTracerStore(profile, true)
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
func tracerDelete(profile, key string) error {
	db, err := openTracerStore(profile, false)
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
func tracerHasData(profile, key string) (bool, error) {
	keys, err := tracerKeys(profile, key)
	if err != nil {
		return false, err
	}
	return len(keys) > 0, nil
}

// ClearData deletes all messages stored for the trace key.
func tracerClearData(profile, key string) error {
	return tracerDelete(profile, key)
}
