package emqutiti

import (
	"strings"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	mqttoptions "github.com/marang/emqutiti/mqttclient"
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

func TestWithAuthOption(t *testing.T) {
	opts := mqtt.NewClientOptions()
	mqttoptions.WithAuth("user", "pass")(opts)
	if opts.Username != "user" || opts.Password != "pass" {
		t.Fatalf("auth option not applied")
	}
}

func TestWithAuthDefaults(t *testing.T) {
	opts := mqtt.NewClientOptions()
	mqttoptions.WithAuth("", "")(opts)
	if opts.Username != "" || opts.Password != "" {
		t.Fatalf("expected defaults for username and password")
	}
}

func TestWithTimeouts(t *testing.T) {
	opts := mqtt.NewClientOptions()
	mqttoptions.WithTimeouts(10, 20)(opts)
	if opts.ConnectTimeout != 10*time.Second || opts.KeepAlive != 20 {
		t.Fatalf("timeouts not applied: got %v and %d", opts.ConnectTimeout, opts.KeepAlive)
	}
}

func TestWithTimeoutsDefaults(t *testing.T) {
	opts := mqtt.NewClientOptions()
	mqttoptions.WithTimeouts(0, 0)(opts)
	if opts.ConnectTimeout != 30*time.Second || opts.KeepAlive != 30 {
		t.Fatalf("expected default timeouts, got %v and %d", opts.ConnectTimeout, opts.KeepAlive)
	}
}
