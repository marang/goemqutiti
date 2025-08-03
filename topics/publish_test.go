package topics

import "testing"

func TestToggleTopic(t *testing.T) {
	c := newTestComponent()
	c.Items = []Item{{Name: "foo", Subscribed: true}}
	c.ToggleTopic(0)
	if c.Items[0].Subscribed {
		t.Fatalf("topic not toggled: %#v", c.Items[0])
	}
}
