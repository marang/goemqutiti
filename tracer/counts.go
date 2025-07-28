package tracer

import "github.com/marang/goemqutiti/history"

// LoadCounts returns per-topic message counts for the given trace key.
func LoadCounts(profile, key string) (map[string]int, error) {
	idx, err := history.OpenTrace(profile)
	if err != nil {
		return nil, err
	}
	defer idx.Close()
	msgs, err := idx.TraceMessages(key)
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int)
	for _, m := range msgs {
		counts[m.Topic]++
	}
	return counts, nil
}
