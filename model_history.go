package emqutiti

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/marang/emqutiti/constants"
	"github.com/marang/emqutiti/history"
	"github.com/marang/emqutiti/ui"
)

// historyPreviewLimit limits preview length for history payloads.
const historyPreviewLimit = 256

// historyModelAdapter satisfies history.Model by delegating to the main model.
type historyModelAdapter struct{ *model }

func (a historyModelAdapter) SetMode(mode history.Mode) tea.Cmd {
	if am, ok := mode.(constants.AppMode); ok {
		return a.model.SetMode(am)
	}
	return nil
}
func (a historyModelAdapter) PreviousMode() history.Mode  { return a.model.PreviousMode() }
func (a historyModelAdapter) CurrentMode() history.Mode   { return a.model.CurrentMode() }
func (a historyModelAdapter) SetFocus(id string) tea.Cmd  { return a.model.SetFocus(id) }
func (a historyModelAdapter) Width() int                  { return a.model.Width() }
func (a historyModelAdapter) Height() int                 { return a.model.Height() }
func (a historyModelAdapter) OverlayHelp(s string) string { return a.model.OverlayHelp(s) }

// historyFilterForm captures filter inputs for history searches.
type historyFilterForm struct {
	ui.Form
	topic    *ui.SuggestField
	payload  *ui.TextField
	start    *ui.TextField
	end      *ui.TextField
	archived *ui.CheckField
}

const (
	idxFilterTopic = iota
	idxFilterPayload
	idxFilterStart
	idxFilterEnd
	idxFilterArchived
)

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
		"",
		fmt.Sprintf("Archived: %s", f.archived.View()),
	)
	return strings.Join(lines, "\n")
}

// historyStore provides an in-memory implementation of history.Store for tests.
type historyStore struct{ msgs []history.Message }

func (s *historyStore) Append(m history.Message) error {
	s.msgs = append(s.msgs, m)
	return nil
}

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
