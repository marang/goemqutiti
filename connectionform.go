package main

import (
	"fmt"
	"reflect"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/marang/goemqutiti/ui"
)

type connectionForm struct {
	Form
	index   int  // -1 for new
	fromEnv bool // current state of env loading
}

type fieldType int

const (
	ftText fieldType = iota
	ftPassword
	ftBool
	ftSelect
)

type fieldDef struct {
	key, label, placeholder string
	fieldType               fieldType
	options                 []string
}

var formFields = []fieldDef{
	{key: "Name", label: "Name", placeholder: "Name", fieldType: ftText},
	{key: "Schema", label: "Schema", placeholder: "Schema", fieldType: ftSelect, options: []string{"tcp", "ssl", "ws", "wss"}},
	{key: "Host", label: "Host", placeholder: "Host", fieldType: ftText},
	{key: "Port", label: "Port", placeholder: "Port", fieldType: ftText},
	{key: "ClientID", label: "Client ID", placeholder: "Client ID", fieldType: ftText},
	{key: "RandomIDSuffix", label: "Random ID suffix", placeholder: "Random ID suffix", fieldType: ftBool},
	{key: "Username", label: "Username", placeholder: "Username", fieldType: ftText},
	{key: "Password", label: "Password", fieldType: ftPassword},
	{key: "FromEnv", label: "Load from env", placeholder: "Values from env", fieldType: ftBool},
	{key: "SSL", label: "SSL/TLS", placeholder: "SSL/TLS", fieldType: ftBool},
	{key: "MQTTVersion", label: "MQTT Version", placeholder: "MQTT Version", fieldType: ftSelect, options: []string{"3", "4", "5"}},
	{key: "ConnectTimeout", label: "Connect Timeout (s)", placeholder: "Connect Timeout (s)", fieldType: ftText},
	{key: "KeepAlive", label: "Keep Alive (s)", placeholder: "Keep Alive (s)", fieldType: ftText},
	{key: "QoS", label: "QoS", placeholder: "QoS", fieldType: ftSelect, options: []string{"0", "1", "2"}},
	{key: "AutoReconnect", label: "Auto Reconnect", placeholder: "Auto Reconnect", fieldType: ftBool},
	{key: "ReconnectPeriod", label: "Reconnect Period (s)", placeholder: "Reconnect Period (s)", fieldType: ftText},
	{key: "CleanStart", label: "Clean Start", placeholder: "Clean Start", fieldType: ftBool},
	{key: "SessionExpiry", label: "Session Expiry (s)", placeholder: "Session Expiry (s)", fieldType: ftText},
	{key: "ReceiveMaximum", label: "Receive Maximum", placeholder: "Receive Maximum", fieldType: ftText},
	{key: "MaximumPacketSize", label: "Maximum Packet Size", placeholder: "Maximum Packet Size", fieldType: ftText},
	{key: "TopicAliasMaximum", label: "Topic Alias Maximum", placeholder: "Topic Alias Maximum", fieldType: ftText},
	{key: "RequestResponseInfo", label: "Request Response Info", placeholder: "Request Response Info", fieldType: ftBool},
	{key: "RequestProblemInfo", label: "Request Problem Info", placeholder: "Request Problem Info", fieldType: ftBool},
	{key: "LastWillEnabled", label: "Use Last Will", placeholder: "Use Last Will", fieldType: ftBool},
	{key: "LastWillTopic", label: "Last Will Topic", placeholder: "Last Will Topic", fieldType: ftText},
	{key: "LastWillQos", label: "Last Will QoS", placeholder: "Last Will QoS", fieldType: ftSelect, options: []string{"0", "1", "2"}},
	{key: "LastWillRetain", label: "Last Will Retain", placeholder: "Last Will Retain", fieldType: ftBool},
	{key: "LastWillPayload", label: "Last Will Payload", placeholder: "Last Will Payload", fieldType: ftText},
}

var fieldIndex = func() map[string]int {
	m := make(map[string]int, len(formFields))
	for i, fd := range formFields {
		m[fd.key] = i
	}
	return m
}()

// newConnectionForm builds a form populated from the given profile.
// idx is -1 when creating a new profile.
func newConnectionForm(p Profile, idx int) connectionForm {
	if p.FromEnv {
		ApplyEnvVars(&p)
	}
	pwKey := ""
	if p.Name != "" && p.Username != "" {
		pwKey = fmt.Sprintf("keyring:emqutiti-%s/%s", p.Name, p.Username)
	}
	rv := reflect.ValueOf(p)
	fields := make([]formField, len(formFields))
	for i, fd := range formFields {
		placeholder := fd.placeholder
		if fd.key == "Password" && pwKey != "" {
			placeholder = pwKey
		}
		fv := rv.FieldByName(fd.key)
		var strVal string
		var boolVal bool
		switch fv.Kind() {
		case reflect.String:
			strVal = fv.String()
		case reflect.Int:
			strVal = fmt.Sprintf("%d", fv.Int())
		case reflect.Bool:
			boolVal = fv.Bool()
			strVal = fmt.Sprintf("%v", boolVal)
		}
		switch fd.fieldType {
		case ftBool:
			fields[i] = newCheckField(boolVal)
		case ftSelect:
			fields[i] = newSelectField(strVal, fd.options)
		case ftPassword:
			fields[i] = newTextField(strVal, placeholder, true)
		default:
			fields[i] = newTextField(strVal, placeholder)
		}
	}
	if p.FromEnv {
		idxName := fieldIndex["Name"]
		idxFromEnv := fieldIndex["FromEnv"]
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
	idxFromEnv := fieldIndex["FromEnv"]
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
	for i, in := range f.fields {
		label := formFields[i].label
		if i == f.focus {
			label = ui.FocusedStyle.Render(label)
		}
		s += fmt.Sprintf("%s: %s\n", label, in.View())
	}
	idxFromEnv := fieldIndex["FromEnv"]
	idxName := fieldIndex["Name"]
	if chk, ok := f.fields[idxFromEnv].(*checkField); ok && chk.value {
		prefix := EnvPrefix(f.fields[idxName].Value())
		s += ui.InfoStyle.Render("Values loaded from env vars: "+prefix+"<FIELD>") + "\n"
	}
	s += "\n" + ui.InfoStyle.Render("[enter] save  [esc] cancel")
	return s
}

// Profile builds a Profile struct from the form values.
func (f connectionForm) Profile() Profile {
	p := Profile{}
	rv := reflect.ValueOf(&p).Elem()
	for i, fd := range formFields {
		field := rv.FieldByName(fd.key)
		val := f.fields[i].Value()
		switch field.Kind() {
		case reflect.String:
			field.SetString(val)
		case reflect.Int:
			var iv int
			fmt.Sscan(val, &iv)
			field.SetInt(int64(iv))
		case reflect.Bool:
			var bv bool
			fmt.Sscan(val, &bv)
			field.SetBool(bv)
		}
	}
	return p
}
