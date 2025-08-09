package history

import (
	"fmt"
	"testing"
	"time"

	"github.com/marang/emqutiti/proxy"
)

func TestOpenStoreProxy(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	p, err := proxy.StartProxy("127.0.0.1:0")
	if err != nil {
		t.Fatalf("start proxy: %v", err)
	}
	SetProxyAddr(p.Addr())
	t.Cleanup(p.Stop)

	st, err := openStore("test")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	msg := Message{Timestamp: time.Now(), Topic: "t1", Payload: "p1", Kind: "pub", Retained: false}
	if err := st.Append(msg); err != nil {
		t.Fatalf("append: %v", err)
	}
	key := fmt.Sprintf("%s/%020d", msg.Topic, msg.Timestamp.UnixNano())
	if err := st.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	st2, err := openStore("test")
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer st2.Close()
	msgs := st2.Search(false, []string{"t1"}, time.Time{}, time.Time{}, "")
	if len(msgs) != 1 || msgs[0].Topic != "t1" || msgs[0].Payload != "p1" {
		t.Fatalf("expected persisted message for key %s, got %v", key, msgs)
	}
}
