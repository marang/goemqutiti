package tracer

import "time"

// Message holds a timestamped MQTT message used for traces.
type Message struct {
	Timestamp time.Time
	Topic     string
	Payload   string
	Kind      string
}
