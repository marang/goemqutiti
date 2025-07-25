package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type formField interface {
	Focus()
	Blur()
	Update(msg tea.Msg) tea.Cmd
	View() string
	Value() string
}

type textField struct{ textinput.Model }

func newTextField(value, placeholder string) *textField {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.SetValue(value)
	return &textField{ti}
}

func (t *textField) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	t.Model, cmd = t.Model.Update(msg)
	return cmd
}

func (t *textField) Value() string { return t.Model.Value() }

type checkField struct {
	value   bool
	focused bool
}

func newCheckField(val bool) *checkField { return &checkField{value: val} }

func (c *checkField) Focus() { c.focused = true }
func (c *checkField) Blur()  { c.focused = false }

func (c *checkField) Update(msg tea.Msg) tea.Cmd {
	switch m := msg.(type) {
	case tea.KeyMsg:
		if m.String() == " " {
			c.value = !c.value
		}
	case tea.MouseMsg:
		if m.Type == tea.MouseLeft {
			c.value = !c.value
		}
	}
	return nil
}

func (c *checkField) View() string {
	box := "[ ]"
	if c.value {
		box = "[x]"
	}
	if c.focused {
		return focusedStyle.Render(box)
	}
	return box
}

func (c *checkField) Value() string { return fmt.Sprintf("%v", c.value) }

func (t *textField) Focus()       { t.Model.Focus() }
func (t *textField) Blur()        { t.Model.Blur() }
func (t *textField) View() string { return t.Model.View() }

type connectionForm struct {
	fields []formField
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
	values := []string{
		p.Name,
		p.Schema,
		p.Host,
		fmt.Sprintf("%d", p.Port),
		p.ClientID,
		p.Username,
		p.Password,
		"", // placeholder for checkbox
		p.MQTTVersion,
		fmt.Sprintf("%d", p.ConnectTimeout),
		fmt.Sprintf("%d", p.KeepAlive),
		"", // checkbox
		fmt.Sprintf("%d", p.ReconnectPeriod),
		"", // checkbox
		fmt.Sprintf("%d", p.SessionExpiry),
		fmt.Sprintf("%d", p.ReceiveMaximum),
		fmt.Sprintf("%d", p.MaximumPacketSize),
		fmt.Sprintf("%d", p.TopicAliasMaximum),
		"", // checkbox
		"", // checkbox
		"", // checkbox
		p.LastWillTopic,
		fmt.Sprintf("%d", p.LastWillQos),
		"", // checkbox
		p.LastWillPayload,
	}

	placeholders := []string{
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

	fields := make([]formField, len(values))
	boolIndices := map[int]bool{
		idxSSL:           true,
		idxAutoReconnect: true,
		idxCleanStart:    true,
		idxRequestResp:   true,
		idxRequestProb:   true,
		idxWillEnable:    true,
		idxWillRetain:    true,
	}

	boolValues := []bool{
		p.SSL,
		p.AutoReconnect,
		p.CleanStart,
		p.RequestResponseInfo,
		p.RequestProblemInfo,
		p.LastWillEnabled,
		p.LastWillRetain,
	}
	bi := 0
	for i := range fields {
		if boolIndices[i] {
			fields[i] = newCheckField(boolValues[bi])
			bi++
		} else {
			fields[i] = newTextField(values[i], placeholders[i])
		}
	}
	fields[0].Focus()
	return connectionForm{fields: fields, focus: 0, index: idx}
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
				f.focus = len(f.fields) - 1
			}
			if f.focus >= len(f.fields) {
				f.focus = 0
			}
		}
	case tea.MouseMsg:
		if msg.Type == tea.MouseLeft {
			// crude calculation: rows correspond to field order
			if msg.Y >= 1 && msg.Y-1 < len(f.fields) {
				f.focus = msg.Y - 1
			}
		}
	}
	for i := range f.fields {
		if i == f.focus {
			f.fields[i].Focus()
		} else {
			f.fields[i].Blur()
		}
	}
	if len(f.fields) > 0 {
		f.fields[f.focus].Update(msg)
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
	for i, in := range f.fields {
		s += fmt.Sprintf("%s: %s\n", labels[i], in.View())
	}
	s += "\nPress Enter to save or Esc to cancel"
	return border.Render(s)
}

func (f connectionForm) Profile() Profile {
	vals := make([]string, len(f.fields))
	for i, in := range f.fields {
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
