package main

import "github.com/charmbracelet/bubbles/list"

// messagesToHistoryItems converts a slice of messages into history items
// and a matching slice of list items for use with the history list.
func messagesToHistoryItems(msgs []Message) ([]historyItem, []list.Item) {
	hitems := make([]historyItem, len(msgs))
	litems := make([]list.Item, len(msgs))
	for i, m := range msgs {
		hi := historyItem{
			timestamp: m.Timestamp,
			topic:     m.Topic,
			payload:   m.Payload,
			kind:      m.Kind,
			archived:  m.Archived,
		}
		hitems[i] = hi
		litems[i] = hi
	}
	return hitems, litems
}
