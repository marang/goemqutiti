package focus

import "testing"

// stubFocusable implements focus.Focusable for testing.
type stubFocusable struct{ focused bool }

func (s *stubFocusable) Focus()          { s.focused = true }
func (s *stubFocusable) Blur()           { s.focused = false }
func (s *stubFocusable) IsFocused() bool { return s.focused }
func (s *stubFocusable) View() string    { return "" }

// TestNewFocusMapNilFirst ensures a nil first element doesn't panic or focus others.
func TestNewFocusMapNilFirst(t *testing.T) {
	second := &stubFocusable{}
	fm := NewFocusMap([]Focusable{nil, second})
	if fm.Index() != 0 {
		t.Fatalf("expected index 0, got %d", fm.Index())
	}
	if second.IsFocused() {
		t.Fatalf("second element should not be focused")
	}
}
