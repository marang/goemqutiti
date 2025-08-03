package emqutiti

import connections "github.com/marang/emqutiti/connections"

// PayloadSnapshot represents a stored payload for persistence.
type PayloadSnapshot = connections.PayloadSnapshot

// Snapshot returns a serializable copy of the current payloads.
func (p *payloadsComponent) Snapshot() []connections.PayloadSnapshot {
	out := make([]connections.PayloadSnapshot, len(p.items))
	for i, item := range p.items {
		out[i] = connections.PayloadSnapshot{Topic: item.topic, Payload: item.payload}
	}
	return out
}

// SetSnapshot replaces the current payloads with the provided snapshot.
func (p *payloadsComponent) SetSnapshot(ps []connections.PayloadSnapshot) {
	items := make([]payloadItem, len(ps))
	for i, item := range ps {
		items[i] = payloadItem{topic: item.Topic, payload: item.Payload}
	}
	p.items = items
}
