package connections

// TopicSnapshot represents a topic and its subscription state for persistence.
type TopicSnapshot struct {
	Title      string `toml:"title"`
	Subscribed bool   `toml:"subscribed"`
	Publish    bool   `toml:"publish"`
}

// PayloadSnapshot represents a stored payload for persistence.
type PayloadSnapshot struct {
	Topic   string `toml:"topic"`
	Payload string `toml:"payload"`
}
