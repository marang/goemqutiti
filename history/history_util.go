package history

import "github.com/charmbracelet/bubbles/list"

// MessagesToItems converts a slice of messages into history items and a
// matching slice of list items for use with the history list.
func MessagesToItems(msgs []Message) ([]Item, []list.Item) {
	hitems := make([]Item, len(msgs))
	litems := make([]list.Item, len(msgs))
	for i, m := range msgs {
		hi := Item{
			Timestamp: m.Timestamp,
			Topic:     m.Topic,
			Payload:   m.Payload,
			Kind:      m.Kind,
			Archived:  m.Archived,
		}
		hitems[i] = hi
		litems[i] = hi
	}
	return hitems, litems
}

// ApplyFilter parses the query and retrieves matching messages from the
// store.
func ApplyFilter(q string, store Store, archived bool) ([]Item, []list.Item) {
	if store == nil {
		return nil, nil
	}
	topics, start, end, payload := parseHistoryQuery(q)
	var msgs []Message
	if archived {
		msgs = store.Search(true, topics, start, end, payload)
	} else {
		msgs = store.Search(false, topics, start, end, payload)
	}
	return MessagesToItems(msgs)
}
