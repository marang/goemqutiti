package traces

import (
	"errors"
	"testing"
	"time"

	"github.com/marang/emqutiti/proxy"
)

func TestHasDataAndClear(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	p, err := proxy.StartProxy("127.0.0.1:0")
	if err != nil {
		t.Fatalf("start proxy: %v", err)
	}
	SetProxyAddr(p.Addr())
	t.Cleanup(p.Stop)

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

func TestHasDataEmptyDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	p, err := proxy.StartProxy("127.0.0.1:0")
	if err != nil {
		t.Fatalf("start proxy: %v", err)
	}
	SetProxyAddr(p.Addr())
	t.Cleanup(p.Stop)

	has, err := tracerHasData("test", "k1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if has {
		t.Fatalf("expected no data")
	}
}

func TestTracerAddError(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	p, err := proxy.StartProxy("127.0.0.1:0")
	if err != nil {
		t.Fatalf("start proxy: %v", err)
	}
	SetProxyAddr(p.Addr())
	t.Cleanup(p.Stop)

	old := jsonMarshal
	jsonMarshal = func(any) ([]byte, error) { return nil, errors.New("fail") }
	defer func() { jsonMarshal = old }()

	err = tracerAdd("test", "k1", TracerMessage{Timestamp: time.Now(), Retained: false})
	if err == nil {
		t.Fatalf("expected error")
	}
}
