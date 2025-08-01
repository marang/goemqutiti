package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOverridePasswordFromEnvCallSite(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")
	data := `[[profiles]]
name="demo"
password="orig"
`
	if err := os.WriteFile(cfgPath, []byte(data), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	p, err := LoadProfile("demo", cfgPath)
	if err != nil {
		t.Fatalf("load profile: %v", err)
	}
	os.Setenv("MQTT_PASSWORD", "override")
	defer os.Unsetenv("MQTT_PASSWORD")
	OverridePasswordFromEnv(p)
	if p.Password != "override" {
		t.Fatalf("expected override, got %s", p.Password)
	}
}
