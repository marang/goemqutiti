package emqutiti

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	connections "github.com/marang/emqutiti/connections"
)

type connectResult struct {
	client  *MQTTClient
	profile connections.Profile
	err     error
}

func (r connectResult) Client() connections.Client   { return r.client }
func (r connectResult) Profile() connections.Profile { return r.profile }
func (r connectResult) Err() error                   { return r.err }

// statusFunc reports connection status messages.
type statusFunc func(string)

// connectBroker attempts to connect to the given profile and reports status via callback.
func connectBroker(p connections.Profile, fn statusFunc) tea.Cmd {
	return func() tea.Msg {
		if fn != nil {
			brokerURL := p.BrokerURL()
			fn(fmt.Sprintf("Connecting to %s", brokerURL))
		}
		client, err := NewMQTTClient(p, fn)
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
