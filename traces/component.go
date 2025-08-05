package traces

import (
	"fmt"
	connections "github.com/marang/emqutiti/connections"
	"io"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/marang/emqutiti/focus"
	"github.com/marang/emqutiti/ui"
)

// traceItem represents a single trace configuration and its runtime tracer.
type traceItem struct {
	key    string
	cfg    TracerConfig
	tracer *Tracer
	counts map[string]int
	loaded bool
}

func (t *traceItem) FilterValue() string { return t.key }
func (t *traceItem) Title() string       { return t.key }
func (t *traceItem) Description() string {
	status := "stopped"
	if t.tracer != nil {
		if t.tracer.Running() {
			status = "running"
		} else if t.tracer.Planned() {
			status = "planned"
		}
	} else if time.Now().Before(t.cfg.Start) {
		status = "planned"
	}
	var parts []string
	counts := t.counts
	if t.tracer != nil {
		counts = t.tracer.Counts()
	} else if !t.loaded {
		if c, err := tracerLoadCounts(t.cfg.Profile, t.cfg.Key, t.cfg.Topics); err == nil {
			t.counts = c
			t.loaded = true
			counts = c
		}
	}
	for _, tp := range t.cfg.Topics {
		parts = append(parts, fmt.Sprintf("%s:%d", tp, counts[tp]))
	}
	if len(parts) > 0 {
		status += " " + strings.Join(parts, " ")
	}
	var times []string
	if !t.cfg.Start.IsZero() {
		times = append(times, t.cfg.Start.Format(time.RFC3339))
	}
	if !t.cfg.End.IsZero() {
		times = append(times, t.cfg.End.Format(time.RFC3339))
	}
	if len(times) > 0 {
		status += " " + strings.Join(times, " -> ")
	}
	return status
}

// state groups state related to tracing functionality.
type State struct {
	list    list.Model
	items   []*traceItem
	form    *traceForm
	view    list.Model
	viewKey string
}

// Component implements the traces interface for managing traces. It owns the
// tracing state but delegates broader navigation and history logging to the
// root model.
type Component struct {
	*State
	api   API
	store Store
}

func NewComponent(api API, ts State, store Store) *Component {
	return &Component{State: &ts, api: api, store: store}
}

func (t *Component) Init() tea.Cmd { return nil }

func (t *Component) View() string { return t.viewTraces() }

func (t *Component) Focus() tea.Cmd { return nil }

func (t *Component) Blur() {}

// Focusables satisfies FocusableSet; the base model provides the trace list focusable.
func (t *Component) Focusables() map[string]focus.Focusable { return map[string]focus.Focusable{} }

// List exposes the trace configuration list model.
func (t *Component) List() *list.Model { return &t.list }

// ViewList exposes the trace message list model.
func (t *Component) ViewList() *list.Model { return &t.view }

// traceMsgItem holds a trace message with its sequence number.
type traceMsgItem struct {
	idx int
	msg TracerMessage
}

// FilterValue implements list.Item for filtering.
func (t traceMsgItem) FilterValue() string { return t.msg.Payload }

// Title implements list.Item.
func (t traceMsgItem) Title() string { return t.msg.Topic }

// Description implements list.Item.
func (t traceMsgItem) Description() string { return t.msg.Payload }

// MsgDelegate renders trace messages with numbering and timestamp.
type MsgDelegate struct{ T *Component }

// Height returns the line height for an item.
func (d MsgDelegate) Height() int { return 2 }

// Spacing satisfies list.ItemDelegate.
func (d MsgDelegate) Spacing() int { return 0 }

// Update is a no-op for this delegate.
func (d MsgDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

// Render draws a trace message item with numbering and timestamp.
func (d MsgDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	it := item.(traceMsgItem)
	width := m.Width()
	innerWidth := width - 2
	if innerWidth < 0 {
		innerWidth = 0
	}
	ts := it.msg.Timestamp.Format("2006-01-02T15:04:05.000000000Z07:00")
	header := fmt.Sprintf("%d %s %s:", it.idx, ts, it.msg.Topic)
	lines := []string{lipgloss.PlaceHorizontal(innerWidth, lipgloss.Left,
		lipgloss.NewStyle().Foreground(ui.ColBlue).Render(header))}
	for _, l := range strings.Split(it.msg.Payload, "\n") {
		wrapped := ansi.Wrap(l, innerWidth, " ")
		for _, wl := range strings.Split(wrapped, "\n") {
			lines = append(lines, lipgloss.PlaceHorizontal(innerWidth, lipgloss.Left,
				lipgloss.NewStyle().Foreground(ui.ColSub).Render(wl)))
		}
	}
	barColor := ui.ColDarkGray
	if index == d.T.view.Index() {
		barColor = ui.ColPurple
	}
	bar := lipgloss.NewStyle().Foreground(barColor)
	lines = ui.FormatHistoryLines(lines, width, bar)
	fmt.Fprint(w, strings.Join(lines, "\n"))
}

type traceTickMsg struct{}

// traceTicker schedules periodic refresh events while traces run.
func traceTicker() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg { return traceTickMsg{} })
}

// Update manages the traces list and responds to key presses.
func (t *Component) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case traceTickMsg:
		// refresh
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+d":
			t.SavePlannedTraces()
			return tea.Quit
		case "esc":
			t.SavePlannedTraces()
			return t.api.SetModeClient()
		case "a":
			profs := t.api.Profiles()
			opts := make([]string, len(profs))
			for i, p := range profs {
				opts[i] = p.Name
			}
			topics := t.api.SubscribedTopics()
			f := newTraceForm(opts, t.api.ActiveConnection(), topics)
			t.form = &f
			return tea.Batch(t.api.SetModeEditTrace(), textinput.Blink)
		case "enter":
			i := t.list.Index()
			if i >= 0 && i < len(t.items) {
				it := t.items[i]
				if it.tracer != nil && (it.tracer.Running() || it.tracer.Planned()) {
					t.stopTrace(i)
				} else {
					t.startTrace(i)
				}
			}
		case "v":
			i := t.list.Index()
			if i >= 0 && i < len(t.items) {
				t.loadTraceMessages(i)
				return nil
			}
		case "delete":
			i := t.list.Index()
			if i >= 0 && i < len(t.items) {
				key := t.items[i].key
				t.stopTrace(i)
				t.items = append(t.items[:i], t.items[i+1:]...)
				items := []list.Item{}
				for _, it := range t.items {
					items = append(items, it)
				}
				t.list.SetItems(items)
				if err := removeTrace(key); err != nil {
					t.api.LogHistory("", err.Error(), "log", err.Error())
				}
			}
			if t.anyTraceRunning() {
				return traceTicker()
			}
			return nil
		case "ctrl+shift+up":
			if t.api.TraceHeight() > 1 {
				t.api.SetTraceHeight(t.api.TraceHeight() - 1)
				t.list.SetSize(t.api.Width()-4, t.api.Height()-4)
			}
		case "ctrl+shift+down":
			t.api.SetTraceHeight(t.api.TraceHeight() + 1)
			t.list.SetSize(t.api.Width()-4, t.api.Height()-4)
		}
	}
	t.list, cmd = t.list.Update(msg)
	if t.anyTraceRunning() {
		return tea.Batch(cmd, traceTicker())
	}
	return cmd
}

// UpdateForm handles input for the new trace form.
func (t *Component) UpdateForm(msg tea.Msg) tea.Cmd {
	if t.form == nil {
		return nil
	}
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+d":
			return tea.Quit
		case "esc":
			t.form = nil
			return t.api.SetModeTracer()
		case "enter":
			cfg := t.form.Config()
			if cfg.Key == "" || len(cfg.Topics) == 0 || cfg.Profile == "" {
				t.form.errMsg = "key, profile and topics required"
				return nil
			}
			if cfg.Start.IsZero() {
				cfg.Start = time.Now().Round(time.Second)
				if tf, ok := t.form.Fields[idxTraceStart].(*ui.TextField); ok {
					tf.SetValue(cfg.Start.Format(time.RFC3339))
				}
			}
			if cfg.End.IsZero() {
				cfg.End = cfg.Start.Add(time.Hour)
				if tf, ok := t.form.Fields[idxTraceEnd].(*ui.TextField); ok {
					tf.SetValue(cfg.End.Format(time.RFC3339))
				}
			}
			if t.traceIndex(cfg.Key) >= 0 {
				t.form.errMsg = "trace key exists"
				return nil
			}
			p, err := connections.LoadProfile(cfg.Profile, "")
			if err != nil {
				t.form.errMsg = err.Error()
				return nil
			}
			if p.FromEnv {
				connections.ApplyEnvVars(p)
			} else if env := os.Getenv("EMQUTITI_DEFAULT_PASSWORD"); env != "" {
				p.Password = env
			}
			client, err := t.api.NewClient(*p)
			if err != nil {
				t.form.errMsg = err.Error()
				return nil
			}
			client.Disconnect()
			newItem := &traceItem{key: cfg.Key, cfg: cfg}
			t.items = append(t.items, newItem)
			items := t.list.Items()
			items = append(items, newItem)
			t.list.SetItems(items)
			if err := addTrace(cfg); err != nil {
				t.form.errMsg = err.Error()
				return nil
			}
			t.form = nil
			return t.api.SetModeTracer()
		}
	}
	f, cmd := t.form.Update(msg)
	t.form = &f
	return cmd
}

// UpdateView displays messages captured for a trace.
func (t *Component) UpdateView(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return t.api.SetModeTracer()
		case "ctrl+d":
			return tea.Quit
		case "ctrl+shift+up":
			if t.api.TraceHeight() > 1 {
				t.api.SetTraceHeight(t.api.TraceHeight() - 1)
				t.view.SetSize(t.api.Width()-4, t.api.TraceHeight())
			}
		case "ctrl+shift+down":
			t.api.SetTraceHeight(t.api.TraceHeight() + 1)
			t.view.SetSize(t.api.Width()-4, t.api.TraceHeight())
		}
	}
	t.view, cmd = t.view.Update(msg)
	return cmd
}
