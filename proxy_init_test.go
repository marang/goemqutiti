package emqutiti

import (
	"net"
	"os"
	"testing"

	connections "github.com/marang/emqutiti/connections"
)

func TestInitProxyWritesConfig(t *testing.T) {
	dir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", oldHome)

	addr, p := initProxy()
	if p == nil {
		t.Fatalf("proxy not started")
	}
	defer p.Stop()
	if addr == "" {
		t.Fatalf("no addr returned")
	}
	if got := connections.LoadProxyAddr(); got != addr {
		t.Fatalf("config addr %q != %q", got, addr)
	}
	// ensure port is listening
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial proxy: %v", err)
	}
	conn.Close()
}
