package emqutiti

// TopicSnapshot represents a topic and its subscription state for persistence.
type TopicSnapshot struct {
	Title  string `toml:"title"`
	Active bool   `toml:"active"`
}

// Snapshot returns a serializable copy of the current topics.
func (c *topicsComponent) Snapshot() []TopicSnapshot {
	out := make([]TopicSnapshot, len(c.items))
	for i, t := range c.items {
		out[i] = TopicSnapshot{Title: t.title, Active: t.subscribed}
	}
	return out
}

// SetSnapshot replaces the current topics with the provided snapshot.
func (c *topicsComponent) SetSnapshot(ts []TopicSnapshot) {
	c.items = make([]topicItem, len(ts))
	for i, t := range ts {
		c.items[i] = topicItem{title: t.Title, subscribed: t.Active}
	}
}
