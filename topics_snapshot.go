package emqutiti

import connections "github.com/marang/emqutiti/connections"

// TopicSnapshot represents a topic and its subscription state for persistence.
type TopicSnapshot = connections.TopicSnapshot

// Snapshot returns a serializable copy of the current topics.
func (c *topicsComponent) Snapshot() []connections.TopicSnapshot {
	out := make([]connections.TopicSnapshot, len(c.items))
	for i, t := range c.items {
		out[i] = connections.TopicSnapshot{Title: t.title, Active: t.subscribed}
	}
	return out
}

// SetSnapshot replaces the current topics with the provided snapshot.
func (c *topicsComponent) SetSnapshot(ts []connections.TopicSnapshot) {
	c.items = make([]topicItem, len(ts))
	for i, t := range ts {
		c.items[i] = topicItem{title: t.Title, subscribed: t.Active}
	}
}
