package emqutiti

import "testing"

func TestCalcTopicsInputWidth(t *testing.T) {
	widths := []int{20, 40, 80}
	for _, w := range widths {
		want := calcMessageWidth(w) - 3
		got := calcTopicsInputWidth(w)
		if got != want {
			t.Fatalf("width %d: got %d, want %d", w, got, want)
		}
	}
}

func TestCycleFocusNext(t *testing.T) {
	m, _ := initialModel(nil)
	if _, ok := m.cycleFocus(focusNext); !ok {
		t.Fatalf("cycleFocus next should return true")
	}
	if m.focus.Index() != 2 {
		t.Fatalf("focus index got %d, want 2", m.focus.Index())
	}
}

func TestCycleFocusPrevWraps(t *testing.T) {
	m, _ := initialModel(nil)
	m.SetFocus(idTopic)
	if _, ok := m.cycleFocus(focusPrev); !ok {
		t.Fatalf("cycleFocus prev should return true")
	}
	want := len(m.ui.focusOrder) - 1
	if m.focus.Index() != want {
		t.Fatalf("focus index got %d, want %d", m.focus.Index(), want)
	}
}
