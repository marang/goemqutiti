package traces

import "time"

// Message holds a timestamped MQTT message used for traces.
type TracerMessage struct {
	Timestamp time.Time
	Topic     string
	Payload   string
	Kind      string
}
