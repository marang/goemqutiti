package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/marang/emqutiti/constants"
)

// SelectField allows choosing from a fixed list of options.
type SelectField struct {
	options  []string
	Index    int
	focused  bool
	readOnly bool
}

// NewSelectField creates a new select field with the given value and options.
// It returns an error if no options are provided.
func NewSelectField(val string, opts []string) (*SelectField, error) {
	if len(opts) == 0 {
		return nil, fmt.Errorf("no options provided")
	}
	idx := 0
	for i, o := range opts {
		if o == val {
			idx = i
			break
		}
	}
	return &SelectField{options: opts, Index: idx}, nil
}

func (s *SelectField) Focus() {
	if !s.readOnly {
		s.focused = true
	}
}

func (s *SelectField) Blur() { s.focused = false }

func (s *SelectField) SetReadOnly(ro bool) {
	s.readOnly = ro
	if ro {
		s.Blur()
	}
}

// ReadOnly reports whether the field is read only.
func (s *SelectField) ReadOnly() bool { return s.readOnly }

func (s *SelectField) Update(msg tea.Msg) tea.Cmd {
	if !s.focused || s.readOnly {
		return nil
	}
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case constants.KeyLeft, constants.KeyH:
			s.Index--
		case constants.KeyRight, constants.KeyL, constants.KeySpaceBar:
			s.Index++
		}
		if s.Index < 0 {
			s.Index = len(s.options) - 1
		}
		if s.Index >= len(s.options) {
			s.Index = 0
		}
	}
	return nil
}

func (s *SelectField) View() string {
	if len(s.options) == 0 {
		val := "-"
		if s.focused {
			return FocusedStyle.Render(val)
		}
		return BlurredStyle.Render(val)
	}
	val := s.options[s.Index]
	switch {
	case s.readOnly:
		return BlurredStyle.Render(val)
	case s.focused:
		return FocusedStyle.Render(val)
	default:
		return val
	}
}

func (s *SelectField) Value() string {
	if len(s.options) == 0 {
		return ""
	}
	return s.options[s.Index]
}

// OptionsView renders the available options when the field is focused.
// The current selection is highlighted.
func (s *SelectField) OptionsView() string {
	if !s.focused || len(s.options) == 0 {
		return ""
	}
	items := make([]string, len(s.options))
	for i, opt := range s.options {
		st := lipgloss.NewStyle().Foreground(ColBlue)
		if i == s.Index {
			st = st.Foreground(ColPink).Bold(true)
		}
		items[i] = st.Render(opt)
	}
	return strings.Join(items, "\n")
}
