package traces

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	connections "github.com/marang/emqutiti/connections"
	"github.com/marang/emqutiti/constants"
	"github.com/marang/emqutiti/focus"
	"github.com/marang/emqutiti/history"
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
	list  list.Model
	items []*traceItem
	form  *traceForm
	*history.Component
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

type histModel struct {
	api       API
	prev, cur constants.AppMode
}

func (h *histModel) SetMode(mode history.Mode) tea.Cmd {
	if m, ok := mode.(constants.AppMode); ok {
		h.prev = h.cur
		h.cur = m
		switch m {
		case constants.ModeTracer:
			return h.api.SetModeTracer()
		case constants.ModeViewTrace:
			return h.api.SetModeViewTrace()
		}
	}
	return nil
}

func (h *histModel) PreviousMode() history.Mode { return h.prev }

func (h *histModel) CurrentMode() history.Mode { return h.cur }

func (h *histModel) SetFocus(id string) tea.Cmd { return h.api.SetFocus(id) }

func (h *histModel) Width() int { return h.api.Width() }

func (h *histModel) Height() int { return h.api.Height() }

func (h *histModel) OverlayHelp(v string) string { return h.api.OverlayHelp(v) }

func NewComponent(api API, ts State, store Store) *Component {
	hm := &histModel{api: api, cur: constants.ModeViewTrace, prev: constants.ModeTracer}
	ts.Component = history.NewComponent(hm, nil)
	return &Component{State: &ts, api: api, store: store}
}

func (t *Component) Init() tea.Cmd { return nil }

func (t *Component) View() string { return t.viewTraces() }

func (t *Component) Focus() tea.Cmd { return nil }

func (t *Component) Blur() {}

// Focusables satisfies FocusableSet; the base model provides trace focusables.
func (t *Component) Focusables() map[string]focus.Focusable { return map[string]focus.Focusable{} }

// List exposes the trace configuration list model.
func (t *Component) List() *list.Model { return &t.list }

// ViewList exposes the trace message list model.
func (t *Component) ViewList() *list.Model { return t.Component.List() }

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
				t.Component.List().SetSize(t.api.Width()-4, t.api.TraceHeight())
			}
			return nil
		case "ctrl+shift+down":
			t.api.SetTraceHeight(t.api.TraceHeight() + 1)
			t.Component.List().SetSize(t.api.Width()-4, t.api.TraceHeight())
			return nil
		}
	}
	return t.Component.Update(msg)
}
