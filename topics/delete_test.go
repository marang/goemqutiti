package topics

import "testing"

func TestRemoveTopic(t *testing.T) {
	c := newTestComponent()
	c.Items = []Item{{Name: "a", Subscribed: true}, {Name: "b", Subscribed: false}}
	c.SetSelected(0)
	c.RemoveTopic(0)
	if len(c.Items) != 1 || c.Items[0].Name != "b" {
		t.Fatalf("topic not removed: %#v", c.Items)
	}
}
