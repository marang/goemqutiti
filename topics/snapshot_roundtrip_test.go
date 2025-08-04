package topics

import "testing"

func TestSnapshotRoundTripPublish(t *testing.T) {
	c := newTestComponent()
	c.Items = []Item{{Name: "foo", Subscribed: true, Publish: true}}
	snap := c.Snapshot()
	if len(snap) != 1 || !snap[0].Publish {
		t.Fatalf("publish flag not saved: %#v", snap)
	}
	c.Items = nil
	c.SetSnapshot(snap)
	if len(c.Items) != 1 || !c.Items[0].Publish {
		t.Fatalf("publish flag not restored: %#v", c.Items)
	}
}
