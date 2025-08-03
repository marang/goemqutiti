package traces

import (
	"testing"
	"time"
)

func TestHasDataAndClear(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	cfg := TracerConfig{Profile: "test", Topics: []string{"a"}, Start: time.Now().Add(-time.Millisecond), End: time.Now().Add(200 * time.Millisecond), Key: "k1"}
	fc := newFakeClient()
	tr := newTracer(cfg, fc)
	if err := tr.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	select {
	case <-fc.subCh:
	case <-time.After(time.Second):
		t.Fatalf("subscribe timeout")
	}
	fc.publish("a", "one")
	tr.Stop()

	has, err := tracerHasData("test", "k1")
	if err != nil || !has {
		t.Fatalf("expected data, err=%v", err)
	}
	if err := tracerClearData("test", "k1"); err != nil {
		t.Fatalf("clear: %v", err)
	}
	has, err = tracerHasData("test", "k1")
	if err != nil || has {
		t.Fatalf("expected no data")
	}
}
