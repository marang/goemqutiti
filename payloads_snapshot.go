package emqutiti

// PayloadSnapshot represents a stored payload for persistence.
type PayloadSnapshot struct {
	Topic   string `toml:"topic"`
	Payload string `toml:"payload"`
}

// Snapshot returns a serializable copy of the current payloads.
func (p *payloadsComponent) Snapshot() []PayloadSnapshot {
	out := make([]PayloadSnapshot, len(p.items))
	for i, item := range p.items {
		out[i] = PayloadSnapshot{Topic: item.topic, Payload: item.payload}
	}
	return out
}

// SetSnapshot replaces the current payloads with the provided snapshot.
func (p *payloadsComponent) SetSnapshot(ps []PayloadSnapshot) {
	items := make([]payloadItem, len(ps))
	for i, s := range ps {
		items[i] = payloadItem{topic: s.Topic, payload: s.Payload}
	}
	p.SetItems(items)
}
