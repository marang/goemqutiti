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
