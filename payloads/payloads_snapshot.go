package payloads

// Snapshot returns a serializable copy of the current payloads.
func (p *Component) Snapshot() []Snapshot {
	out := make([]Snapshot, len(p.items))
	for i, item := range p.items {
		out[i] = Snapshot{Topic: item.Topic, Payload: item.Payload}
	}
	return out
}

// SetSnapshot replaces the current payloads with the provided snapshot.
func (p *Component) SetSnapshot(ps []Snapshot) {
	seen := make(map[Item]struct{}, len(ps))
	items := make([]Item, 0, len(ps))
	for _, snap := range ps {
		item := Item{Topic: snap.Topic, Payload: snap.Payload}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		items = append(items, item)
	}
	p.SetItems(items)
}
