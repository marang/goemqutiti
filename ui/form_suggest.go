package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/marang/emqutiti/constants"
)

// SuggestField is a text input with auto-completion suggestions.
type SuggestField struct {
	*TextField
	options     []string
	suggestions []string
	sel         int
}

// NewSuggestField creates a SuggestField with the given options and placeholder.
func NewSuggestField(opts []string, placeholder string) *SuggestField {
	tf := NewTextField("", placeholder)
	return &SuggestField{
		TextField:   tf,
		options:     append([]string(nil), opts...),
		suggestions: nil,
		sel:         -1,
	}
}

// Update processes key messages to cycle suggestions while forwarding other
// messages to the underlying text field.
func (s *SuggestField) Update(msg tea.Msg) tea.Cmd {
	switch m := msg.(type) {
	case tea.KeyMsg:
		switch m.String() {
		case constants.KeyTab, constants.KeyDown:
			if len(s.suggestions) > 0 {
				s.sel = (s.sel + 1) % len(s.suggestions)
				s.SetValue(s.suggestions[s.sel])
				s.CursorEnd()
				return nil
			}
		case constants.KeyShiftTab, constants.KeyUp:
			if len(s.suggestions) > 0 {
				s.sel--
				if s.sel < 0 {
					s.sel = len(s.suggestions) - 1
				}
				s.SetValue(s.suggestions[s.sel])
				s.CursorEnd()
				return nil
			}
		case constants.KeyEnter, constants.KeySpaceBar, constants.KeySpace:
			if len(s.suggestions) > 0 && s.sel >= 0 {
				s.SetValue(s.suggestions[s.sel])
				s.suggestions = nil
				s.sel = -1
				s.CursorEnd()
				return nil
			}
		}
	}
	cmd := s.TextField.Update(msg)
	if s.Focused() {
		prefix := s.Value()
		s.suggestions = s.suggestions[:0]
		s.sel = -1
		for _, o := range s.options {
			if prefix == "" || strings.HasPrefix(o, prefix) {
				s.suggestions = append(s.suggestions, o)
				if len(s.suggestions) == 5 {
					break
				}
			}
		}
	}
	return cmd
}

// SuggestionsView renders the suggestion list for the field.
func (s *SuggestField) SuggestionsView() string {
	if !s.Focused() || len(s.suggestions) == 0 {
		return ""
	}
	items := make([]string, len(s.suggestions))
	for i, sug := range s.suggestions {
		st := lipgloss.NewStyle().Foreground(ColBlue)
		if i == s.sel {
			st = st.Foreground(ColPink)
		}
		items[i] = st.Render(sug)
	}
	return strings.Join(items, "\n")
}

// WantsKey reports whether the field wants to handle the key itself to cycle
// suggestions instead of moving focus.
func (s *SuggestField) WantsKey(k tea.KeyMsg) bool {
	switch k.String() {
	case constants.KeyTab, constants.KeyShiftTab, constants.KeyUp, constants.KeyDown, constants.KeyEnter, constants.KeySpaceBar, constants.KeySpace:
		return len(s.suggestions) > 0
	default:
		return s.TextField.WantsKey(k)
	}
}
