package connections

// TopicSnapshot represents a topic and its subscription state for persistence.
type TopicSnapshot struct {
	Title  string `toml:"title"`
	Active bool   `toml:"active"`
}

// PayloadSnapshot represents a stored payload for persistence.
type PayloadSnapshot struct {
	Topic   string `toml:"topic"`
	Payload string `toml:"payload"`
}
