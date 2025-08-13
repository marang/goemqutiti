package layout

import "testing"

func TestTopicsInputWidth(t *testing.T) {
	m := Manager{}
	widths := []int{20, 40, 80}
	for _, w := range widths {
		want := w - 7
		got := m.TopicsInputWidth(w)
		if got != want {
			t.Fatalf("width %d: got %d, want %d", w, got, want)
		}
	}
}

func TestHistorySizeDefault(t *testing.T) {
	m := Manager{}
	w, h := m.HistorySize(60, 30)
	if w != 56 {
		t.Fatalf("width got %d, want 56", w)
	}
	wantH := (30-1)/3 + 10
	if h != wantH {
		t.Fatalf("height got %d, want %d", h, wantH)
	}
	if m.History.Height != wantH {
		t.Fatalf("stored height %d, want %d", m.History.Height, wantH)
	}
}

func TestTraceHeightDefault(t *testing.T) {
	m := Manager{}
	h := m.TraceHeight(40)
	if h != 34 {
		t.Fatalf("height got %d, want 34", h)
	}
	if m.Trace.Height != 34 {
		t.Fatalf("stored height %d, want 34", m.Trace.Height)
	}
}
