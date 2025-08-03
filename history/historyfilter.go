package history

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/marang/emqutiti/ui"
)

const (
	idxFilterTopic = iota
	idxFilterPayload
	idxFilterStart
	idxFilterEnd
)

// historyFilterForm captures filter inputs for history searches.
type historyFilterForm struct {
	ui.Form
	topic   *ui.SuggestField
	payload *ui.TextField
	start   *ui.TextField
	end     *ui.TextField
}

// Topic returns the topic field.
func (f *historyFilterForm) Topic() *ui.SuggestField { return f.topic }

// Payload returns the payload field.
func (f *historyFilterForm) Payload() *ui.TextField { return f.payload }

// Start returns the start time field.
func (f *historyFilterForm) Start() *ui.TextField { return f.start }

// End returns the end time field.
func (f *historyFilterForm) End() *ui.TextField { return f.end }

// newHistoryFilterForm builds a form with optional prefilled values.
// Start and end remain blank when zero, allowing searches across all time.
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

// NewFilterForm builds a history filter form with optional prefilled values.
func NewFilterForm(topics []string, topic, payload string, start, end time.Time) historyFilterForm {
	return newHistoryFilterForm(topics, topic, payload, start, end)
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
