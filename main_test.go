package emqutiti

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMain creates a temporary home with an empty config file so tests run
// without noisy warnings about missing user configuration.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "emqutiti_test")
	if err != nil {
		os.Exit(1)
	}
	os.Setenv("HOME", dir)
	cfgDir := filepath.Join(dir, ".config", "emqutiti")
	os.MkdirAll(cfgDir, 0o755)
	os.WriteFile(filepath.Join(cfgDir, "config.toml"), []byte{}, 0o644)
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}
