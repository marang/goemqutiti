package emqutiti

import (
	"strings"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type fakeToken struct {
	done bool
	err  error
}

func (f *fakeToken) Wait() bool                       { return f.done }
func (f *fakeToken) WaitTimeout(d time.Duration) bool { return f.done }
func (f *fakeToken) Done() <-chan struct{} {
	ch := make(chan struct{})
	if f.done {
		close(ch)
	}
	return ch
}
func (f *fakeToken) Error() error { return f.err }

var _ mqtt.Token = (*fakeToken)(nil)

func TestDisconnectClosesMessageChan(t *testing.T) {
	ch := make(chan MQTTMessage)
	c := &MQTTClient{MessageChan: ch}
	c.Disconnect()
	if _, ok := <-ch; ok {
		t.Fatalf("expected MessageChan to be closed")
	}
}

func TestWaitTokenSuccess(t *testing.T) {
	tok := &fakeToken{done: true}
	if err := waitToken(tok, time.Second, "publish"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestWaitTokenTimeout(t *testing.T) {
	tok := &fakeToken{done: false}
	to := 100 * time.Millisecond
	err := waitToken(tok, to, "subscribe")
	if err == nil || !strings.Contains(err.Error(), "subscribe timeout") {
		t.Fatalf("expected timeout error, got %v", err)
	}
}
