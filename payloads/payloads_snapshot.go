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
	items := make([]Item, len(ps))
	for i, item := range ps {
		items[i] = Item{Topic: item.Topic, Payload: item.Payload}
	}
	p.items = items
}
