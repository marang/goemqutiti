package connections

import (
	"os"
	"testing"
)

func TestSaveLoadProxyAddr(t *testing.T) {
	dir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", oldHome)

	if err := SaveProxyAddr("127.0.0.1:12345"); err != nil {
		t.Fatalf("SaveProxyAddr: %v", err)
	}
	if got := LoadProxyAddr(); got != "127.0.0.1:12345" {
		t.Fatalf("LoadProxyAddr = %q", got)
	}
	// Ensure SaveState preserves proxy address.
	if err := SaveState(map[string]ConnectionSnapshot{}); err != nil {
		t.Fatalf("SaveState: %v", err)
	}
	if got := LoadProxyAddr(); got != "127.0.0.1:12345" {
		t.Fatalf("proxy addr lost after SaveState: %q", got)
	}
}
