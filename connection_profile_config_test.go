package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/zalando/go-keyring"
)

func TestLoadConfig(t *testing.T) {
	keyring.MockInit()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")
	data := `[[profiles]]
name = "p1"
username = "u1"
password = "keyring:svc/u1"

[[profiles]]
name = "env"
from_env = true
`
	if err := os.WriteFile(cfgPath, []byte(data), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := keyring.Set("svc", "u1", "secret"); err != nil {
		t.Fatalf("keyring set: %v", err)
	}
	os.Setenv("EMQUTITI_ENV_HOST", "example.com")
	os.Setenv("EMQUTITI_ENV_PORT", "1884")
	t.Cleanup(func() {
		os.Unsetenv("EMQUTITI_ENV_HOST")
		os.Unsetenv("EMQUTITI_ENV_PORT")
	})

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig error: %v", err)
	}
	if len(cfg.Profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(cfg.Profiles))
	}
	if cfg.Profiles[0].Password != "secret" {
		t.Errorf("keyring password not resolved: %v", cfg.Profiles[0].Password)
	}
	p := cfg.Profiles[1]
	if p.Host != "example.com" || p.Port != 1884 {
		t.Errorf("env vars not applied: %+v", p)
	}
}

func TestLoadConfigKeyringError(t *testing.T) {
	keyring.MockInit()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")
	data := `[[profiles]]
name = "p1"
username = "u1"
password = "keyring:svc/u1"
`
	if err := os.WriteFile(cfgPath, []byte(data), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if _, err := LoadConfig(cfgPath); err == nil {
		t.Fatal("expected error for missing keyring entry")
	}
}

func TestSaveConfig(t *testing.T) {
	dir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", oldHome)

	profiles := []Profile{{Name: "a"}}
	saveConfig(profiles, "a")
	cfgPath, _ := DefaultUserConfigFile()
	var cfg userConfig
	if _, err := toml.DecodeFile(cfgPath, &cfg); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if cfg.DefaultProfileName != "a" || len(cfg.Profiles) != 1 || cfg.Profiles[0].Name != "a" {
		t.Fatalf("unexpected cfg: %#v", cfg)
	}
}

func TestLoadProfile(t *testing.T) {
	keyring.MockInit()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.toml")
	data := `default_profile = "p2"
[[profiles]]
name = "p1"
[[profiles]]
name = "p2"
`
	if err := os.WriteFile(cfgPath, []byte(data), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	p, err := LoadProfile("p1", cfgPath)
	if err != nil || p.Name != "p1" {
		t.Fatalf("LoadProfile explicit: %v %#v", err, p)
	}
	p, err = LoadProfile("", cfgPath)
	if err != nil || p.Name != "p2" {
		t.Fatalf("LoadProfile default: %v %#v", err, p)
	}
}
