package tracer

import (
	"os"
	"testing"
	"time"
)

func TestHasDataAndClear(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("HOME", dir)

	cfg := Config{Profile: "test", Topics: []string{"a"}, Start: time.Now().Add(-time.Millisecond), End: time.Now().Add(200 * time.Millisecond), Key: "k1"}
	fc := newFakeClient()
	tr := New(cfg, fc)
	if err := tr.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	time.Sleep(20 * time.Millisecond)
	fc.publish("a", "one")
	tr.Stop()
	time.Sleep(5 * time.Millisecond)

	has, err := HasData("test", "k1")
	if err != nil || !has {
		t.Fatalf("expected data, err=%v", err)
	}
	if err := ClearData("test", "k1"); err != nil {
		t.Fatalf("clear: %v", err)
	}
	has, err = HasData("test", "k1")
	if err != nil || has {
		t.Fatalf("expected no data")
	}
}
