package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type connectionForm struct {
	inputs []textinput.Model
	focus  int
	index  int // -1 for new
}

const (
	idxName = iota
	idxSchema
	idxHost
	idxPort
	idxClientID
	idxUsername
	idxPassword
	idxSSL
	idxMQTTVersion
	idxConnectTimeout
	idxKeepAlive
	idxAutoReconnect
	idxReconnectPeriod
	idxCleanStart
	idxSessionExpiry
	idxReceiveMaximum
	idxMaximumPacket
	idxTopicAlias
	idxRequestResp
	idxRequestProb
	idxWillEnable
	idxWillTopic
	idxWillQos
	idxWillRetain
	idxWillPayload
)

func newConnectionForm(p Profile, idx int) connectionForm {
	fields := []string{
		p.Name,
		p.Schema,
		p.Host,
		fmt.Sprintf("%d", p.Port),
		p.ClientID,
		p.Username,
		p.Password,
		fmt.Sprintf("%v", p.SSL),
		p.MQTTVersion,
		fmt.Sprintf("%d", p.ConnectTimeout),
		fmt.Sprintf("%d", p.KeepAlive),
		fmt.Sprintf("%v", p.AutoReconnect),
		fmt.Sprintf("%d", p.ReconnectPeriod),
		fmt.Sprintf("%v", p.CleanStart),
		fmt.Sprintf("%d", p.SessionExpiry),
		fmt.Sprintf("%d", p.ReceiveMaximum),
		fmt.Sprintf("%d", p.MaximumPacketSize),
		fmt.Sprintf("%d", p.TopicAliasMaximum),
		fmt.Sprintf("%v", p.RequestResponseInfo),
		fmt.Sprintf("%v", p.RequestProblemInfo),
		fmt.Sprintf("%v", p.LastWillEnabled),
		p.LastWillTopic,
		fmt.Sprintf("%d", p.LastWillQos),
		fmt.Sprintf("%v", p.LastWillRetain),
		p.LastWillPayload,
	}
	inputs := make([]textinput.Model, len(fields))
	placeholders := []string{
		"Name",
		"Schema",
		"Host",
		"Port",
		"Client ID",
		"Username",
		"Password",
		"SSL/TLS (true/false)",
		"MQTT Version",
		"Connect Timeout",
		"Keep Alive",
		"Auto Reconnect (true/false)",
		"Reconnect Period",
		"Clean Start (true/false)",
		"Session Expiry",
		"Receive Maximum",
		"Maximum Packet Size",
		"Topic Alias Maximum",
		"Request Response Info (true/false)",
		"Request Problem Info (true/false)",
		"Use Last Will (true/false)",
		"Last Will Topic",
		"Last Will QoS",
		"Last Will Retain (true/false)",
		"Last Will Payload",
	}
	for i := range inputs {
		ti := textinput.New()
		ti.Placeholder = placeholders[i]
		ti.SetValue(fields[i])
		if i == 0 {
			ti.Focus()
		}
		inputs[i] = ti
	}
	return connectionForm{inputs: inputs, focus: 0, index: idx}
}

func (f connectionForm) Init() tea.Cmd {
	return textinput.Blink
}

func (f connectionForm) Update(msg tea.Msg) (connectionForm, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab":
			if msg.String() == "shift+tab" {
				f.focus--
			} else {
				f.focus++
			}
			if f.focus < 0 {
				f.focus = len(f.inputs) - 1
			}
			if f.focus >= len(f.inputs) {
				f.focus = 0
			}
		}
	}
	for i := range f.inputs {
		if i == f.focus {
			f.inputs[i].Focus()
		} else {
			f.inputs[i].Blur()
		}
		f.inputs[i], _ = f.inputs[i].Update(msg)
	}
	return f, cmd
}

func (f connectionForm) View() string {
	border := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1).BorderForeground(lipgloss.Color("63"))
	var s string
	labels := []string{
		"Name",
		"Schema",
		"Host",
		"Port",
		"Client ID",
		"Username",
		"Password",
		"SSL/TLS",
		"MQTT Version",
		"Connect Timeout",
		"Keep Alive",
		"Auto Reconnect",
		"Reconnect Period",
		"Clean Start",
		"Session Expiry",
		"Receive Maximum",
		"Maximum Packet Size",
		"Topic Alias Maximum",
		"Request Response Info",
		"Request Problem Info",
		"Use Last Will",
		"Last Will Topic",
		"Last Will QoS",
		"Last Will Retain",
		"Last Will Payload",
	}
	for i, in := range f.inputs {
		s += fmt.Sprintf("%s: %s\n", labels[i], in.View())
	}
	s += "\nPress Enter to save or Esc to cancel"
	return border.Render(s)
}

func (f connectionForm) Profile() Profile {
	vals := make([]string, len(f.inputs))
	for i, in := range f.inputs {
		vals[i] = in.Value()
	}
	p := Profile{}
	p.Name = vals[idxName]
	p.Schema = vals[idxSchema]
	p.Host = vals[idxHost]
	fmt.Sscan(vals[idxPort], &p.Port)
	p.ClientID = vals[idxClientID]
	p.Username = vals[idxUsername]
	p.Password = vals[idxPassword]
	fmt.Sscan(vals[idxSSL], &p.SSL)
	p.MQTTVersion = vals[idxMQTTVersion]
	fmt.Sscan(vals[idxConnectTimeout], &p.ConnectTimeout)
	fmt.Sscan(vals[idxKeepAlive], &p.KeepAlive)
	fmt.Sscan(vals[idxAutoReconnect], &p.AutoReconnect)
	fmt.Sscan(vals[idxReconnectPeriod], &p.ReconnectPeriod)
	fmt.Sscan(vals[idxCleanStart], &p.CleanStart)
	fmt.Sscan(vals[idxSessionExpiry], &p.SessionExpiry)
	fmt.Sscan(vals[idxReceiveMaximum], &p.ReceiveMaximum)
	fmt.Sscan(vals[idxMaximumPacket], &p.MaximumPacketSize)
	fmt.Sscan(vals[idxTopicAlias], &p.TopicAliasMaximum)
	fmt.Sscan(vals[idxRequestResp], &p.RequestResponseInfo)
	fmt.Sscan(vals[idxRequestProb], &p.RequestProblemInfo)
	fmt.Sscan(vals[idxWillEnable], &p.LastWillEnabled)
	p.LastWillTopic = vals[idxWillTopic]
	fmt.Sscan(vals[idxWillQos], &p.LastWillQos)
	fmt.Sscan(vals[idxWillRetain], &p.LastWillRetain)
	p.LastWillPayload = vals[idxWillPayload]
	return p
}
