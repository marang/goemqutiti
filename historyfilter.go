package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	idxFilterTopic = iota
	idxFilterStart
	idxFilterEnd
)

// historyFilterForm captures filter inputs for history searches.
type historyFilterForm struct {
	Form
	topic  *textField
	start  *textField
	end    *textField
	topics []string
}

// newHistoryFilterForm builds a form prefilled with the last hour.
func newHistoryFilterForm(topics []string) historyFilterForm {
	end := time.Now()
	start := end.Add(-time.Hour)
	tf := newTextField("", "topic")
	sf := newTextField(start.Format(time.RFC3339), "start (RFC3339)")
	ef := newTextField(end.Format(time.RFC3339), "end (RFC3339)")
	f := historyFilterForm{
		topic:  tf,
		start:  sf,
		end:    ef,
		topics: append([]string(nil), topics...),
	}
	f.fields = []formField{tf, sf, ef}
	sort.Strings(f.topics)
	f.ApplyFocus()
	return f
}

// Update handles focus cycling and topic completion.
func (f historyFilterForm) Update(msg tea.Msg) (historyFilterForm, tea.Cmd) {
	var cmd tea.Cmd
	switch m := msg.(type) {
	case tea.KeyMsg:
		if f.focus == idxFilterTopic && m.String() == "tab" {
			prefix := f.topic.Value()
			for _, t := range f.topics {
				if strings.HasPrefix(t, prefix) {
					f.topic.SetValue(t)
					f.topic.CursorEnd()
					break
				}
			}
			return f, nil
		}
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
	return f, cmd
}

// View renders the filter fields with labels.
func (f historyFilterForm) View() string {
	lines := []string{
		fmt.Sprintf("Topic: %s", f.topic.View()),
		fmt.Sprintf("Start: %s", f.start.View()),
		fmt.Sprintf("End:   %s", f.end.View()),
	}
	return strings.Join(lines, "\n")
}

// query builds a history search string.
func (f historyFilterForm) query() string {
	var parts []string
	if v := f.topic.Value(); v != "" {
		parts = append(parts, "topic="+v)
	}
	if v := f.start.Value(); v != "" {
		parts = append(parts, "start="+v)
	}
	if v := f.end.Value(); v != "" {
		parts = append(parts, "end="+v)
	}
	return strings.Join(parts, " ")
}
