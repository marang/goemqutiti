package tracer

import "github.com/marang/goemqutiti/history"

// LoadCounts returns per-topic counts for the given trace key aggregated by
// the provided subscription topics.
func LoadCounts(profile, key string, topics []string) (map[string]int, error) {
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
	for _, t := range topics {
		counts[t] = 0
	}
	for _, m := range msgs {
		for _, sub := range topics {
			if Match(sub, m.Topic) {
				counts[sub]++
			}
		}
	}
	return counts, nil
}
