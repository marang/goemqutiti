package emqutiti

import "testing"

func TestCycleFocusNext(t *testing.T) {
	m, _ := initialModel(nil)
	if _, ok := m.cycleFocus(focusNext); !ok {
		t.Fatalf("cycleFocus next should return true")
	}
	if m.focus.Index() != 1 {
		t.Fatalf("focus index got %d, want 1", m.focus.Index())
	}
}

func TestCycleFocusPrevWraps(t *testing.T) {
	m, _ := initialModel(nil)
	if _, ok := m.cycleFocus(focusPrev); !ok {
		t.Fatalf("cycleFocus prev should return true")
	}
	want := len(m.ui.focusOrder) - 1
	if m.focus.Index() != want {
		t.Fatalf("focus index got %d, want %d", m.focus.Index(), want)
	}
}
