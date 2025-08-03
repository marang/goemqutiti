package topics

import "testing"

func TestScroll(t *testing.T) {
	c := newTestComponent()
	c.VP.SetContent("a\nb\nc\nd\ne\nf\n")
	c.VP.Height = 2
	c.Scroll(1)
	if c.VP.YOffset == 0 {
		t.Fatalf("expected scroll to move viewport")
	}
}
