package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type statusMessage string

type connectResult struct {
	client  *MQTTClient
	profile Profile
	err     error
}

// connectBroker attempts to connect to the given profile and reports status on the channel.
func connectBroker(p Profile, ch chan string) tea.Cmd {
	return func() tea.Msg {
		if ch != nil {
			brokerURL := fmt.Sprintf("%s://%s:%d", p.Schema, p.Host, p.Port)
			ch <- fmt.Sprintf("Connecting to %s", brokerURL)
		}
		client, err := NewMQTTClient(p, ch)
		return connectResult{client: client, profile: p, err: err}
	}
}

// listenMessages waits for incoming MQTT messages on the provided channel.
func listenMessages(ch chan MQTTMessage) tea.Cmd {
	return func() tea.Msg {
		if ch == nil {
			return nil
		}
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}

// listenStatus retrieves status updates from the status channel.
func listenStatus(ch chan string) tea.Cmd {
	return func() tea.Msg {
		if ch == nil {
			return nil
		}
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return statusMessage(msg)
	}
}

// flushStatus drains all pending messages from the status channel.
func flushStatus(ch chan string) {
	if ch == nil {
		return
	}
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}
