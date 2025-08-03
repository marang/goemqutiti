package emqutiti

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
	"unsafe"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/marang/emqutiti/history"
	"github.com/marang/emqutiti/ui"
)

// historyPreviewLimit limits preview length for history payloads.
const historyPreviewLimit = 256

// historyState holds the internal state for the history component.
type historyState struct {
	list            list.Model
	items           []history.Item
	store           history.Store
	selectionAnchor int
	showArchived    bool
	filterForm      *historyFilterForm
	filterQuery     string
	detail          viewport.Model
	detailItem      history.Item
}

// historyComponent provides history browsing and filtering.
type historyComponent struct {
	*historyState
	m history.Model
}

// historyModelAdapter satisfies history.Model by delegating to the main model.
type historyModelAdapter struct{ *model }

func (a historyModelAdapter) SetMode(mode history.Mode) tea.Cmd {
	if am, ok := mode.(appMode); ok {
		return a.model.setMode(am)
	}
	return nil
}
func (a historyModelAdapter) PreviousMode() history.Mode  { return a.model.previousMode() }
func (a historyModelAdapter) CurrentMode() history.Mode   { return a.model.currentMode() }
func (a historyModelAdapter) SetFocus(id string) tea.Cmd  { return a.model.SetFocus(id) }
func (a historyModelAdapter) Width() int                  { return a.model.Width() }
func (a historyModelAdapter) Height() int                 { return a.model.Height() }
func (a historyModelAdapter) OverlayHelp(s string) string { return a.model.OverlayHelp(s) }

// newHistoryComponent constructs a history component bound to the model using
// the exported constructor from the history package.
func newHistoryComponent(m *model, hs historyState) *historyComponent {
	comp := history.NewComponent(historyModelAdapter{m}, hs.store)
	return (*historyComponent)(unsafe.Pointer(comp))
}

// Init performs no initialization and returns nil.
func (h *historyComponent) Init() tea.Cmd { return nil }

// Update updates the history list. The caller must ensure the list has focus.
func (h *historyComponent) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	h.list, cmd = h.list.Update(msg)
	return cmd
}

// View renders no standalone view for the history component.
func (h *historyComponent) View() string { return "" }

// Focus gives focus to the history component. Currently a no-op.
func (h *historyComponent) Focus() tea.Cmd { return nil }

// Blur removes focus from the history component. Currently a no-op.
func (h *historyComponent) Blur() {}

// UpdateDetail handles input when viewing a long history payload.
func (h *historyComponent) UpdateDetail(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return h.m.SetMode(h.m.PreviousMode())
		case "ctrl+d":
			return tea.Quit
		}
	}
	h.detail, cmd = h.detail.Update(msg)
	return cmd
}

// UpdateFilter handles the history filter form interaction.
func (h *historyComponent) UpdateFilter(msg tea.Msg) tea.Cmd {
	if h.filterForm == nil {
		return nil
	}
	switch t := msg.(type) {
	case tea.KeyMsg:
		switch t.String() {
		case "esc":
			h.filterForm = nil
			cmd := tea.Batch(h.m.SetMode(h.m.PreviousMode()), h.m.SetFocus(history.ID))
			return cmd
		case "enter":
			q := h.filterForm.query()
			topics, start, end, payload := parseHistoryQuery(q)
			var msgs []history.Message
			if h.showArchived {
				msgs = h.store.Search(true, topics, start, end, payload)
			} else {
				msgs = h.store.Search(false, topics, start, end, payload)
			}
			var items []list.Item
			h.items, items = messagesToHistoryItems(msgs)
			h.list.SetItems(items)
			h.list.FilterInput.SetValue("")
			h.list.SetFilterState(list.Unfiltered)
			h.filterQuery = q
			h.filterForm = nil
			cmd := tea.Batch(h.m.SetMode(h.m.PreviousMode()), h.m.SetFocus(history.ID))
			return cmd
		}
	}
	f, cmd := h.filterForm.Update(msg)
	h.filterForm = &f
	return cmd
}

// ViewDetail renders the full payload of a history message.
func (h *historyComponent) ViewDetail() string {
	lines := strings.Split(h.detail.View(), "\n")
	help := ui.InfoStyle.Render("[esc] back")
	lines = append(lines, help)
	content := strings.Join(lines, "\n")
	sp := -1.0
	if h.detail.Height < lipgloss.Height(content) {
		sp = h.detail.ScrollPercent()
	}
	view := ui.LegendBox(content, "Message", h.m.Width()-2, h.m.Height()-2, ui.ColGreen, true, sp)
	return h.m.OverlayHelp(view)
}

// ViewFilter displays the history filter form.
func (h *historyComponent) ViewFilter() string {
	if h.filterForm == nil {
		return ""
	}
	content := lipgloss.NewStyle().Padding(1, 2).Render(h.filterForm.View())
	box := ui.LegendBox(content, "Filter", h.m.Width()/2, 0, ui.ColBlue, true, -1)
	return lipgloss.Place(h.m.Width(), h.m.Height(), lipgloss.Center, lipgloss.Center, box)
}

// Focusables exposes focusable elements for the history component.
func (h *historyComponent) Focusables() map[string]Focusable {
	return map[string]Focusable{}
}

// Append stores a message in the history list and optional store.
func (h *historyComponent) Append(topic, payload, kind, logText string) {
	ts := time.Now()
	text := payload
	if kind == "log" {
		text = logText
	}
	hi := history.Item{Timestamp: ts, Topic: topic, Payload: text, Kind: kind, Archived: false}
	if h.store != nil {
		h.store.Append(history.Message{Timestamp: ts, Topic: topic, Payload: payload, Kind: kind, Archived: false})
	}
	if !h.showArchived {
		if h.filterQuery != "" {
			var items []list.Item
			h.items, items = applyHistoryFilter(h.filterQuery, h.store, h.showArchived)
			h.list.SetItems(items)
			h.list.Select(len(items) - 1)
		} else {
			h.items = append(h.items, hi)
			items := make([]list.Item, len(h.items))
			for i, it := range h.items {
				items[i] = it
			}
			h.list.SetItems(items)
			h.list.Select(len(items) - 1)
		}
	}
}

// Scroll forwards scroll events to the history list.
func (h *historyComponent) Scroll(msg tea.MouseMsg) tea.Cmd {
	var cmd tea.Cmd
	h.list, cmd = h.list.Update(msg)
	return cmd
}

// HandleSelection updates history selection based on index and shift key.
func (h *historyComponent) HandleSelection(idx int, shift bool) {
	h.list.Select(idx)
	if shift {
		if h.selectionAnchor == -1 {
			h.selectionAnchor = h.list.Index()
			if h.selectionAnchor >= 0 && h.selectionAnchor < len(h.items) {
				v := true
				h.items[h.selectionAnchor].IsSelected = &v
			}
		}
		h.updateSelectionRange(idx)
	} else {
		for i := range h.items {
			h.items[i].IsSelected = nil
		}
		h.selectionAnchor = -1
	}
}

// HandleClick selects a history item based on mouse position.
func (h *historyComponent) HandleClick(msg tea.MouseMsg, top, vpYOffset int) {
	idx := h.indexAt(msg.Y, top, vpYOffset)
	if idx >= 0 {
		h.HandleSelection(idx, msg.Shift)
	}
}

// UpdateSelectionRange selects history entries from the anchor to idx.
func (h *historyComponent) UpdateSelectionRange(idx int) { h.updateSelectionRange(idx) }

func (h *historyComponent) indexAt(y, top, vpYOffset int) int {
	rel := y - (top + 1) + vpYOffset
	if rel < 0 {
		return -1
	}
	hgt := 2 // history delegate height
	idx := rel / hgt
	start := h.list.Paginator.Page * h.list.Paginator.PerPage
	i := start + idx
	if i >= len(h.list.Items()) || i < 0 {
		return -1
	}
	return i
}

func (h *historyComponent) updateSelectionRange(idx int) {
	start := h.selectionAnchor
	end := idx
	if start > end {
		start, end = end, start
	}
	for i := range h.items {
		h.items[i].IsSelected = nil
	}
	for i := start; i <= end && i < len(h.items); i++ {
		v := true
		h.items[i].IsSelected = &v
	}
}

// historyDelegate renders history items in the list.
type historyDelegate struct{}

// Height returns the fixed height for history entries.
func (d historyDelegate) Height() int { return 2 }

// Spacing returns the row spacing for history entries.
func (d historyDelegate) Spacing() int { return 0 }

// Update performs no update and returns nil.
func (d historyDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

// Render prints a history item with its label and payload.
func (d historyDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	hi := item.(history.Item)
	width := m.Width()
	var label string
	ts := hi.Timestamp.Format("2006-01-02 15:04:05.000")
	var lblColor lipgloss.Color
	var msgColor lipgloss.Color
	switch hi.Kind {
	case "sub":
		label = fmt.Sprintf("SUB %s", hi.Topic)
		lblColor = ui.ColPink
		msgColor = ui.ColPub
	case "pub":
		label = fmt.Sprintf("PUB %s", hi.Topic)
		lblColor = ui.ColBlue
		msgColor = ui.ColSub
	default:
		label = ""
		lblColor = ui.ColGray
		msgColor = ui.ColGray
	}
	align := lipgloss.Left
	if hi.Kind == "pub" {
		align = lipgloss.Right
	}
	innerWidth := width - 2
	if innerWidth < 0 {
		innerWidth = 0
	}

	// Render at most two lines so the list height stays consistent
	var lines []string
	if hi.Kind != "log" {
		header := lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Foreground(lblColor).Render(label),
			lipgloss.NewStyle().Foreground(ui.ColGray).Render(" "+ts+":"))
		lines = append(lines, lipgloss.PlaceHorizontal(innerWidth, align, header))
	}
	payload := strings.ReplaceAll(hi.Payload, "\r\n", "\n")
	payload = strings.ReplaceAll(payload, "\n", "\u23ce")
	more := utf8.RuneCountInString(payload) > historyPreviewLimit
	if more {
		payload = ansi.Truncate(payload, historyPreviewLimit, "")
	}
	trunc := ansi.Truncate(hi.Payload, innerWidth, "")
	trunc = strings.NewReplacer("\r\n", "\u23ce", "\n", "\u23ce").Replace(trunc)
	if more || lipgloss.Width(hi.Payload) > innerWidth {
		if lipgloss.Width(trunc) >= innerWidth {
			trunc = ansi.Truncate(trunc, innerWidth-1, "")
		}
		trunc += "\u2026"
	}
	fg := msgColor
	if hi.Kind == "log" && len(lines) == 0 {
		trunc = ts + ": " + trunc
		fg = ui.ColGray
	}
	lines = append(lines, lipgloss.PlaceHorizontal(innerWidth, align,
		lipgloss.NewStyle().Foreground(fg).Render(trunc)))
	if len(lines) < 2 {
		lines = append(lines, lipgloss.PlaceHorizontal(innerWidth, align, ""))
	}
	if hi.IsSelected != nil && *hi.IsSelected {
		for i, l := range lines {
			lines[i] = lipgloss.NewStyle().Background(ui.ColDarkGray).Render(l)
		}
	}
	barColor := ui.ColGray
	if hi.Kind == "log" {
		barColor = ui.ColDarkGray
	}
	if hi.IsSelected != nil && *hi.IsSelected {
		barColor = ui.ColBlue
	}
	if index == m.Index() {
		barColor = ui.ColPurple
	}
	bar := lipgloss.NewStyle().Foreground(barColor)
	lines = ui.FormatHistoryLines(lines, width, bar)
	fmt.Fprint(w, strings.Join(lines, "\n"))
}

// historyFilterForm captures filter inputs for history searches.
type historyFilterForm struct {
	ui.Form
	topic   *ui.SuggestField
	payload *ui.TextField
	start   *ui.TextField
	end     *ui.TextField
}

const (
	idxFilterTopic = iota
	idxFilterPayload
	idxFilterStart
	idxFilterEnd
)

// newHistoryFilterForm builds a form with optional prefilled values.
func newHistoryFilterForm(topics []string, topic, payload string, start, end time.Time) historyFilterForm {
	sort.Strings(topics)
	tf := ui.NewSuggestField(topics, "topic")
	tf.SetValue(topic)

	pf := ui.NewTextField("", "text contains")
	pf.SetValue(payload)

	sf := ui.NewTextField("", "start (RFC3339)")
	if !start.IsZero() {
		sf.SetValue(start.Format(time.RFC3339))
	}

	ef := ui.NewTextField("", "end (RFC3339)")
	if !end.IsZero() {
		ef.SetValue(end.Format(time.RFC3339))
	}

	f := historyFilterForm{
		Form:    ui.Form{Fields: []ui.Field{tf, pf, sf, ef}},
		topic:   tf,
		payload: pf,
		start:   sf,
		end:     ef,
	}
	f.ApplyFocus()
	return f
}

// Update handles focus cycling and topic completion.
func (f historyFilterForm) Update(msg tea.Msg) (historyFilterForm, tea.Cmd) {
	var cmd tea.Cmd
	switch m := msg.(type) {
	case tea.KeyMsg:
		if c, ok := f.Fields[f.Focus].(ui.KeyConsumer); ok && c.WantsKey(m) {
			cmd = f.Fields[f.Focus].Update(msg)
		} else {
			f.CycleFocus(m)
			if len(f.Fields) > 0 {
				cmd = f.Fields[f.Focus].Update(msg)
			}
		}
	case tea.MouseMsg:
		if m.Action == tea.MouseActionPress && m.Button == tea.MouseButtonLeft {
			if m.Y >= 1 && m.Y-1 < len(f.Fields) {
				f.Focus = m.Y - 1
			}
		}
		if len(f.Fields) > 0 {
			cmd = f.Fields[f.Focus].Update(msg)
		}
	}
	f.ApplyFocus()
	return f, cmd
}

// View renders the filter fields with labels.
func (f historyFilterForm) View() string {
	line := fmt.Sprintf("Topic: %s", f.topic.View())
	lines := []string{line}
	if sugg := f.topic.SuggestionsView(); sugg != "" {
		lines = append(lines, sugg)
	}
	lines = append(lines,
		"",
		fmt.Sprintf("Text:  %s", f.payload.View()),
		"",
		fmt.Sprintf("Start: %s", f.start.View()),
		"",
		fmt.Sprintf("End:   %s", f.end.View()),
	)
	return strings.Join(lines, "\n")
}

// query builds a history search string.
func (f historyFilterForm) query() string {
	var parts []string
	if v := f.topic.Value(); v != "" {
		parts = append(parts, "topic="+v)
	}
	if v := f.payload.Value(); v != "" {
		parts = append(parts, "payload="+v)
	}
	if v := f.start.Value(); v != "" {
		parts = append(parts, "start="+v)
	}
	if v := f.end.Value(); v != "" {
		parts = append(parts, "end="+v)
	}
	return strings.Join(parts, " ")
}

// messagesToHistoryItems converts messages into history items and list items.
func messagesToHistoryItems(msgs []history.Message) ([]history.Item, []list.Item) {
	hitems := make([]history.Item, len(msgs))
	litems := make([]list.Item, len(msgs))
	for i, m := range msgs {
		hi := history.Item{
			Timestamp: m.Timestamp,
			Topic:     m.Topic,
			Payload:   m.Payload,
			Kind:      m.Kind,
			Archived:  m.Archived,
		}
		hitems[i] = hi
		litems[i] = hi
	}
	return hitems, litems
}

// applyHistoryFilter parses the query and retrieves matching messages from the store.
func applyHistoryFilter(q string, store history.Store, archived bool) ([]history.Item, []list.Item) {
	if store == nil {
		return nil, nil
	}
	topics, start, end, payload := parseHistoryQuery(q)
	var msgs []history.Message
	if archived {
		msgs = store.Search(true, topics, start, end, payload)
	} else {
		msgs = store.Search(false, topics, start, end, payload)
	}
	return messagesToHistoryItems(msgs)
}

// parseHistoryQuery interprets a filter string.
func parseHistoryQuery(q string) (topics []string, start, end time.Time, payload string) {
	var payloadParts []string
	for _, f := range strings.Fields(q) {
		switch {
		case strings.HasPrefix(f, "topic="):
			ts := strings.TrimPrefix(f, "topic=")
			if ts != "" {
				topics = strings.Split(ts, ",")
			}
		case strings.HasPrefix(f, "start="):
			t, err := time.Parse(time.RFC3339, strings.TrimPrefix(f, "start="))
			if err == nil {
				start = t
			}
		case strings.HasPrefix(f, "end="):
			t, err := time.Parse(time.RFC3339, strings.TrimPrefix(f, "end="))
			if err == nil {
				end = t
			}
		case strings.HasPrefix(f, "payload="):
			payloadParts = append(payloadParts, strings.TrimPrefix(f, "payload="))
		default:
			payloadParts = append(payloadParts, f)
		}
	}
	payload = strings.Join(payloadParts, " ")
	return
}

// historyStore provides an in-memory implementation of history.Store for tests.
type historyStore struct{ msgs []history.Message }

func (s *historyStore) Append(m history.Message) { s.msgs = append(s.msgs, m) }

func (s *historyStore) Search(archived bool, topics []string, start, end time.Time, payload string) []history.Message {
	var out []history.Message
	for _, m := range s.msgs {
		if m.Archived != archived {
			continue
		}
		if len(topics) > 0 {
			match := false
			for _, t := range topics {
				if m.Topic == t {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}
		if !start.IsZero() && m.Timestamp.Before(start) {
			continue
		}
		if !end.IsZero() && m.Timestamp.After(end) {
			continue
		}
		if payload != "" && !strings.Contains(m.Payload, payload) {
			continue
		}
		out = append(out, m)
	}
	return out
}

func (s *historyStore) Delete(string) error  { return nil }
func (s *historyStore) Archive(string) error { return nil }
func (s *historyStore) Count(archived bool) int {
	c := 0
	for _, m := range s.msgs {
		if m.Archived == archived {
			c++
		}
	}
	return c
}
func (s *historyStore) Close() error { return nil }
