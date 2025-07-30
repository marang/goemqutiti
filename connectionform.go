package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/marang/goemqutiti/config"
	"github.com/marang/goemqutiti/ui"
)

type connectionForm struct {
	Form
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
	cf := connectionForm{Form: Form{fields: fields, focus: 0}, index: idx, fromEnv: p.FromEnv}
	cf.ApplyFocus()
	return cf
}

// Init sets up the text input blink command.
func (f connectionForm) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles keyboard and mouse events for the form.
func (f connectionForm) Update(msg tea.Msg) (connectionForm, tea.Cmd) {
	var cmd tea.Cmd
	switch m := msg.(type) {
	case tea.KeyMsg:
		f.CycleFocus(m)
	case tea.MouseMsg:
		if m.Action == tea.MouseActionPress && m.Button == tea.MouseButtonLeft {
			if m.Y >= 1 && m.Y-1 < len(f.fields) {
				f.focus = m.Y - 1
			}
		}
	}
	f.ApplyFocus()
	if len(f.fields) > 0 {
		cmd = f.fields[f.focus].Update(msg)
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
