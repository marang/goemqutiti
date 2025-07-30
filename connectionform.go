package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/marang/goemqutiti/config"
	"github.com/marang/goemqutiti/ui"
)

type formField interface {
	Focus()
	Blur()
	Update(msg tea.Msg) tea.Cmd
	View() string
	Value() string
}

type textField struct {
	textinput.Model
	readOnly bool
}

// newTextField creates a textField with the given value and placeholder. If opts[0] is true the field is masked for password entry.
func newTextField(value, placeholder string, opts ...bool) *textField {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.SetValue(value)
	if len(opts) > 0 && opts[0] {
		ti.EchoMode = textinput.EchoPassword
	}
	return &textField{Model: ti}
}

// setReadOnly marks the field read only and blurs it when activated.
func (t *textField) setReadOnly(ro bool) {
	t.readOnly = ro
	if ro {
		t.Blur()
	}
}

// Update forwards messages to the text input unless the field is read only.
func (t *textField) Update(msg tea.Msg) tea.Cmd {
	if t.readOnly {
		return nil
	}
	var cmd tea.Cmd
	t.Model, cmd = t.Model.Update(msg)
	return cmd
}

// Value returns the text content of the field.
func (t *textField) Value() string { return t.Model.Value() }

type checkField struct {
	value    bool
	focused  bool
	readOnly bool
}

type selectField struct {
	options  []string
	index    int
	focused  bool
	readOnly bool
}

// newSelectField creates a selectField with the given options and selects the provided value.
func newSelectField(val string, opts []string) *selectField {
	idx := 0
	for i, o := range opts {
		if o == val {
			idx = i
			break
		}
	}
	return &selectField{options: opts, index: idx}
}

// Focus sets the field as focused unless it is read only.
func (s *selectField) Focus() {
	if !s.readOnly {
		s.focused = true
	}
}

// Blur removes focus from the field.
func (s *selectField) Blur() { s.focused = false }

// setReadOnly disables interaction and clears focus when true.
func (s *selectField) setReadOnly(ro bool) {
	s.readOnly = ro
	if ro {
		s.Blur()
	}
}

// Update handles key presses when the field is focused.
func (s *selectField) Update(msg tea.Msg) tea.Cmd {
	if !s.focused || s.readOnly {
		return nil
	}
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "left", "h":
			s.index--
		case "right", "l", " ":
			s.index++
		}
		if s.index < 0 {
			s.index = len(s.options) - 1
		}
		if s.index >= len(s.options) {
			s.index = 0
		}
	}
	return nil
}

// View renders the current option and highlights it when focused.
func (s *selectField) View() string {
	val := s.options[s.index]
	if s.focused {
		return ui.FocusedStyle.Render(val)
	}
	return val
}

// Value returns the currently selected option.
func (s *selectField) Value() string { return s.options[s.index] }

// newCheckField returns a checkbox field with the given value.
func newCheckField(val bool) *checkField { return &checkField{value: val} }

// Focus sets the checkbox as active for input.
func (c *checkField) Focus() { c.focused = true }

// Blur removes focus from the checkbox.
func (c *checkField) Blur() { c.focused = false }

// setReadOnly prevents toggling the checkbox value.
func (c *checkField) setReadOnly(ro bool) { c.readOnly = ro }

// Update toggles the checkbox state on keyboard or mouse input.
func (c *checkField) Update(msg tea.Msg) tea.Cmd {
	switch m := msg.(type) {
	case tea.KeyMsg:
		if !c.readOnly && m.String() == " " {
			c.value = !c.value
		}
	case tea.MouseMsg:
		if !c.readOnly && m.Action == tea.MouseActionPress && m.Button == tea.MouseButtonLeft {
			c.value = !c.value
		}
	}
	return nil
}

// View renders the checkbox with focus styling when selected.
func (c *checkField) View() string {
	box := "[ ]"
	if c.value {
		box = "[x]"
	}
	if c.focused {
		return ui.FocusedStyle.Render(box)
	}
	return box
}

// Value returns "true" or "false" for the checkbox state.
func (c *checkField) Value() string { return fmt.Sprintf("%v", c.value) }

// Focus gives keyboard focus to the text field.
// Blur removes keyboard focus from the text field.
// View returns the underlying textinput view.

func (t *textField) Focus()       { t.Model.Focus() }
func (t *textField) Blur()        { t.Model.Blur() }
func (t *textField) View() string { return t.Model.View() }

type connectionForm struct {
	fields  []formField
	focus   int
	index   int  // -1 for new
	fromEnv bool // current state of env loading
}

const (
	idxName = iota
	idxSchema
	idxHost
	idxPort
	idxClientID
	idxIDSuffix
	idxUsername
	idxPassword
	idxFromEnv
	idxSSL
	idxMQTTVersion
	idxConnectTimeout
	idxKeepAlive
	idxQoS
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

// newConnectionForm builds a form populated from the given profile.
// idx is -1 when creating a new profile.
func newConnectionForm(p Profile, idx int) connectionForm {
	if p.FromEnv {
		config.ApplyEnvVars(&p)
	}
	pwKey := ""
	if p.Name != "" && p.Username != "" {
		pwKey = fmt.Sprintf("keyring:emqutiti-%s/%s", p.Name, p.Username)
	}
	values := []string{
		p.Name,
		p.Schema,
		p.Host,
		fmt.Sprintf("%d", p.Port),
		p.ClientID,
		"", // checkbox for rand suffix
		p.Username,
		p.Password,
		"", // checkbox: load from env
		"", // SSL checkbox
		p.MQTTVersion,
		fmt.Sprintf("%d", p.ConnectTimeout),
		fmt.Sprintf("%d", p.KeepAlive),
		fmt.Sprintf("%d", p.QoS),
		"", // auto reconnect checkbox
		fmt.Sprintf("%d", p.ReconnectPeriod),
		"", // clean start checkbox
		fmt.Sprintf("%d", p.SessionExpiry),
		fmt.Sprintf("%d", p.ReceiveMaximum),
		fmt.Sprintf("%d", p.MaximumPacketSize),
		fmt.Sprintf("%d", p.TopicAliasMaximum),
		"", // request response info checkbox
		"", // request problem info checkbox
		"", // last will enabled checkbox
		p.LastWillTopic,
		fmt.Sprintf("%d", p.LastWillQos),
		"", // last will retain checkbox
		p.LastWillPayload,
	}

	placeholders := []string{
		"Name",
		"Schema",
		"Host",
		"Port",
		"Client ID",
		"Random ID suffix",
		"Username",
		pwKey,
		"Values from env",
		"SSL/TLS",
		"MQTT Version",
		"Connect Timeout (s)",
		"Keep Alive (s)",
		"QoS",
		"Auto Reconnect",
		"Reconnect Period (s)",
		"Clean Start",
		"Session Expiry (s)",
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
		idxIDSuffix:      true,
		idxFromEnv:       true,
		idxSSL:           true,
		idxAutoReconnect: true,
		idxCleanStart:    true,
		idxRequestResp:   true,
		idxRequestProb:   true,
		idxWillEnable:    true,
		idxWillRetain:    true,
	}

	selectOptions := map[int][]string{
		idxSchema:      {"tcp", "ssl", "ws", "wss"},
		idxMQTTVersion: {"3", "4", "5"},
		idxQoS:         {"0", "1", "2"},
		idxWillQos:     {"0", "1", "2"},
	}

	boolValues := []bool{
		p.RandomIDSuffix,
		p.FromEnv,
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
			continue
		}
		if opts, ok := selectOptions[i]; ok {
			fields[i] = newSelectField(values[i], opts)
			continue
		}
		if i == idxPassword {
			fields[i] = newTextField(values[i], placeholders[i], true)
		} else {
			fields[i] = newTextField(values[i], placeholders[i])
		}
	}
	if p.FromEnv {
		for i, fld := range fields {
			if i == idxName || i == idxFromEnv {
				continue
			}
			if ro, ok := fld.(interface{ setReadOnly(bool) }); ok {
				ro.setReadOnly(true)
			}
		}
	}
	fields[0].Focus()
	return connectionForm{fields: fields, focus: 0, index: idx, fromEnv: p.FromEnv}
}

// Init sets up the text input blink command.
func (f connectionForm) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles keyboard and mouse events for the form.
func (f connectionForm) Update(msg tea.Msg) (connectionForm, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "up", "down", "k", "j":
			step := 0
			switch msg.String() {
			case "shift+tab", "up", "k":
				step = -1
			default:
				step = 1
			}
			f.focus += step
			if f.focus < 0 {
				f.focus = len(f.fields) - 1
			}
			if f.focus >= len(f.fields) {
				f.focus = 0
			}
		}
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
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
	if chk, ok := f.fields[idxFromEnv].(*checkField); ok && chk.value != f.fromEnv {
		p := f.Profile()
		f = newConnectionForm(p, f.index)
		f.focus = idxFromEnv
		f.fromEnv = chk.value
	}
	return f, cmd
}

// View renders the form with labels and field contents.
func (f connectionForm) View() string {
	var s string
	labels := []string{
		"Name",
		"Schema",
		"Host",
		"Port",
		"Client ID",
		"Random ID suffix",
		"Username",
		"Password",
		"Load from env",
		"SSL/TLS",
		"MQTT Version",
		"Connect Timeout (s)",
		"Keep Alive (s)",
		"QoS",
		"Auto Reconnect",
		"Reconnect Period (s)",
		"Clean Start",
		"Session Expiry (s)",
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
		label := labels[i]
		if i == f.focus {
			label = ui.FocusedStyle.Render(label)
		}
		s += fmt.Sprintf("%s: %s\n", label, in.View())
	}
	if chk, ok := f.fields[idxFromEnv].(*checkField); ok && chk.value {
		prefix := config.EnvPrefix(f.fields[idxName].Value())
		s += ui.InfoStyle.Render("Values loaded from env vars: "+prefix+"<FIELD>") + "\n"
	}
	s += "\n" + ui.InfoStyle.Render("[enter] save  [esc] cancel")
	return s
}

// Profile builds a Profile struct from the form values.
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
	fmt.Sscan(vals[idxFromEnv], &p.FromEnv)
	fmt.Sscan(vals[idxSSL], &p.SSL)
	p.MQTTVersion = vals[idxMQTTVersion]
	fmt.Sscan(vals[idxConnectTimeout], &p.ConnectTimeout)
	fmt.Sscan(vals[idxKeepAlive], &p.KeepAlive)
	fmt.Sscan(vals[idxQoS], &p.QoS)
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
	fmt.Sscan(vals[idxIDSuffix], &p.RandomIDSuffix)
	return p
}
