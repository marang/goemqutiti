package tracer

import "github.com/marang/goemqutiti/history"

// HasData reports whether any messages are stored for the given trace key.
func HasData(profile, key string) (bool, error) {
	idx, err := history.OpenTrace(profile)
	if err != nil {
		return false, err
	}
	defer idx.Close()
	keys, err := idx.TraceKeys(key)
	if err != nil {
		return false, err
	}
	return len(keys) > 0, nil
}

// ClearData deletes all messages stored for the trace key.
func ClearData(profile, key string) error {
	idx, err := history.OpenTrace(profile)
	if err != nil {
		return err
	}
	defer idx.Close()
	return idx.DeleteTrace(key)
}
