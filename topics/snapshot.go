package topics

import connections "github.com/marang/emqutiti/connections"

// TopicSnapshot represents a topic and its subscription state for persistence.
type TopicSnapshot = connections.TopicSnapshot

// Snapshot returns a serializable copy of the current topics.
func (c *Component) Snapshot() []connections.TopicSnapshot {
	out := make([]connections.TopicSnapshot, len(c.Items))
	for i, t := range c.Items {
		out[i] = connections.TopicSnapshot{Title: t.Name, Subscribed: t.Subscribed, Publish: t.Publish}
	}
	return out
}

// SetSnapshot replaces the current topics with the provided snapshot.
func (c *Component) SetSnapshot(ts []connections.TopicSnapshot) {
	c.Items = make([]Item, len(ts))
	for i, t := range ts {
		c.Items[i] = Item{Name: t.Title, Subscribed: t.Subscribed, Publish: t.Publish}
	}
}
