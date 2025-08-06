package emqutiti

import "testing"

func TestDisconnectClosesMessageChan(t *testing.T) {
	ch := make(chan MQTTMessage)
	c := &MQTTClient{MessageChan: ch}
	c.Disconnect()
	if _, ok := <-ch; ok {
		t.Fatalf("expected MessageChan to be closed")
	}
}
