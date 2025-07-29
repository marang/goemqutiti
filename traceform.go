package main

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/marang/goemqutiti/ui"
)

type traceForm struct {
	fields []formField
	focus  int
	errMsg string
}

const (
	idxTraceKey = iota
	idxTraceProfile
	idxTraceTopics
	idxTraceStart
	idxTraceEnd
)

func newTraceForm(profiles []string, current string, topics []string) traceForm {
	tf := traceForm{}
	keyField := newTextField("", "Key")
	profileField := newSelectField(current, profiles)
	topicsField := newTextField(strings.Join(topics, ","), "Topics")
	startField := newTextField("", "2006-01-02T15:04:05Z")
	endField := newTextField("", "2006-01-02T15:04:05Z")
	tf.fields = []formField{keyField, profileField, topicsField, startField, endField}
	tf.focus = 0
	tf.fields[0].Focus()
	return tf
}

func (f traceForm) Init() tea.Cmd { return textinput.Blink }

func (f traceForm) Update(msg tea.Msg) (traceForm, tea.Cmd) {
	var cmd tea.Cmd
	switch m := msg.(type) {
	case tea.KeyMsg:
		switch m.String() {
		case "tab", "shift+tab", "up", "down", "k", "j":
			step := 1
			if m.String() == "shift+tab" || m.String() == "up" || m.String() == "k" {
				step = -1
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
		if m.Action == tea.MouseActionPress && m.Button == tea.MouseButtonLeft {
			if m.Y >= 1 && m.Y-1 < len(f.fields) {
				f.focus = m.Y - 1
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

func (f traceForm) View() string {
	labels := []string{"Key", "Profile", "Topics", "Start", "End"}
	var b strings.Builder
	for i, fld := range f.fields {
		label := labels[i]
		if i == f.focus {
			label = ui.FocusedStyle.Render(label)
		}
		b.WriteString(label + ": " + fld.View() + "\n")
	}
	if f.errMsg != "" {
		b.WriteString("\n" + ui.ErrorStyle.Render(f.errMsg))
	}
	b.WriteString("\n" + ui.InfoStyle.Render("[enter] save  [esc] cancel"))
	return b.String()
}

func (f traceForm) Config() TracerConfig {
	vals := make([]string, len(f.fields))
	for i, fld := range f.fields {
		vals[i] = fld.Value()
	}
	cfg := TracerConfig{}
	cfg.Key = strings.TrimSpace(vals[idxTraceKey])
	cfg.Profile = vals[idxTraceProfile]
	if t := strings.TrimSpace(vals[idxTraceTopics]); t != "" {
		parts := strings.Split(t, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		cfg.Topics = parts
	}
	if s := strings.TrimSpace(vals[idxTraceStart]); s != "" {
		if tm, err := time.Parse(time.RFC3339, s); err == nil {
			cfg.Start = tm
		}
	}
	if e := strings.TrimSpace(vals[idxTraceEnd]); e != "" {
		if tm, err := time.Parse(time.RFC3339, e); err == nil {
			cfg.End = tm
		}
	}
	return cfg
}
