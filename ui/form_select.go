package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/marang/emqutiti/constants"
)

// SelectField allows choosing from a fixed list of options.
type SelectField struct {
	options  []string
	Index    int
	focused  bool
	readOnly bool
}

func NewSelectField(val string, opts []string) *SelectField {
	idx := 0
	for i, o := range opts {
		if o == val {
			idx = i
			break
		}
	}
	return &SelectField{options: opts, Index: idx}
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
	val := s.options[s.Index]
	if s.focused {
		return FocusedStyle.Render(val)
	}
	return val
}

func (s *SelectField) Value() string { return s.options[s.Index] }
