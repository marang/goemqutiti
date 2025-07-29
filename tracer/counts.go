package tracer

// LoadCounts returns per-topic counts for the given trace key aggregated by
// the provided subscription topics.
func LoadCounts(profile, key string, topics []string) (map[string]int, error) {
	msgs, err := Messages(profile, key)
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
