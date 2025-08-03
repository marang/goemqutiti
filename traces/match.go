package traces

import "strings"

// tracerMatch reports whether the topic matches the MQTT subscription filter.
func tracerMatch(filter, topic string) bool {
	fp := strings.Split(filter, "/")
	tp := strings.Split(topic, "/")
	for i := 0; i < len(fp); i++ {
		if fp[i] == "#" {
			return true
		}
		if i >= len(tp) {
			return false
		}
		if fp[i] != "+" && fp[i] != tp[i] {
			return false
		}
	}
	return len(fp) == len(tp)
}
