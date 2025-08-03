package traces

import (
	"os"
	"strings"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type fakeMessage struct {
	topic   string
	payload []byte
}

func (f fakeMessage) Duplicate() bool   { return false }
func (f fakeMessage) Qos() byte         { return 0 }
func (f fakeMessage) Retained() bool    { return false }
func (f fakeMessage) Topic() string     { return f.topic }
func (f fakeMessage) MessageID() uint16 { return 0 }
func (f fakeMessage) Payload() []byte   { return f.payload }
func (f fakeMessage) Ack()              {}

type fakeClient struct {
	subs  map[string]mqtt.MessageHandler
	subCh chan struct{}
}

func newFakeClient() *fakeClient {
	return &fakeClient{subs: make(map[string]mqtt.MessageHandler), subCh: make(chan struct{}, 1)}
}

func (f *fakeClient) Subscribe(topic string, qos byte, cb mqtt.MessageHandler) error {
	f.subs[topic] = cb
	select {
	case f.subCh <- struct{}{}:
	default:
	}
	return nil
}
func (f *fakeClient) Unsubscribe(topic string) error { delete(f.subs, topic); return nil }
func (f *fakeClient) Disconnect()                    {}

func (f *fakeClient) publish(topic, payload string) {
	if cb, ok := f.subs[topic]; ok {
		cb(nil, fakeMessage{topic: topic, payload: []byte(payload)})
	}
}

func TestTraceStartAndStore(t *testing.T) {
	dir := t.TempDir()
	os.Setenv("HOME", dir)

	cfg := TracerConfig{
		Profile: "test",
		Topics:  []string{"a"},
		Start:   time.Now().Add(-time.Millisecond),
		End:     time.Now().Add(200 * time.Millisecond),
		Key:     "k1",
	}
	fc := newFakeClient()
	tr := newTracer(cfg, fc)
	if err := tr.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	time.Sleep(20 * time.Millisecond)
	fc.publish("a", "one")
	fc.publish("a", "two")
	time.Sleep(100 * time.Millisecond)
	tr.Stop()
	time.Sleep(200 * time.Millisecond)

	keys, err := tracerKeys("test", "k1")
	if err != nil {
		t.Fatalf("trace keys: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
	prefix := "trace/k1/a/"
	for _, k := range keys {
		if !strings.HasPrefix(k, prefix) {
			t.Fatalf("bad key %s", k)
		}
	}
}
