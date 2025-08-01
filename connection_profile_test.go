package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/zalando/go-keyring"

	"github.com/marang/goemqutiti/internal/files"
)

// TestApplyEnvVars sets env vars for all Profile fields and ensures they are applied.
func TestApplyEnvVars(t *testing.T) {
	p := Profile{Name: "test", FromEnv: true}
	prefix := EnvPrefix(p.Name)

	rt := reflect.TypeOf(p)
	rv := reflect.ValueOf(&p).Elem()

	var envs []string
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if f.Name == "FromEnv" {
			continue
		}
		tag := f.Tag.Get("toml")
		if tag == "" {
			continue
		}
		envName := prefix + strings.ToUpper(strings.ReplaceAll(tag, "-", "_"))
		switch f.Type.Kind() {
		case reflect.String:
			os.Setenv(envName, "x")
			envs = append(envs, envName)
		case reflect.Int:
			os.Setenv(envName, "1")
			envs = append(envs, envName)
		case reflect.Bool:
			os.Setenv(envName, "true")
			envs = append(envs, envName)
		default:
			t.Fatalf("unsupported kind %s", f.Type.Kind())
		}
	}
	t.Cleanup(func() {
		for _, e := range envs {
			os.Unsetenv(e)
		}
	})

	ApplyEnvVars(&p)

	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if f.Name == "FromEnv" {
			continue
		}
		tag := f.Tag.Get("toml")
		if tag == "" {
			continue
		}
		field := rv.Field(i)
		switch f.Type.Kind() {
		case reflect.String:
			if field.String() != "x" {
				t.Errorf("field %s not set", f.Name)
			}
		case reflect.Int:
			if field.Int() != 1 {
				t.Errorf("field %s not set", f.Name)
			}
		case reflect.Bool:
			if field.Bool() != true {
				t.Errorf("field %s not set", f.Name)
			}
		}
	}
}

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

func TestSavePasswordToKeyring(t *testing.T) {
	keyring.MockInit()
	if err := savePasswordToKeyring("svc", "user", "pw"); err != nil {
		t.Fatalf("savePasswordToKeyring: %v", err)
	}
	got, err := keyring.Get("emqutiti-svc", "user")
	if err != nil || got != "pw" {
		t.Fatalf("got %q err %v", got, err)
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

func TestDeleteProfileData(t *testing.T) {
	dir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", oldHome)

	base := files.DataDir("p1")
	os.MkdirAll(filepath.Join(base, "history"), 0755)
	os.MkdirAll(filepath.Join(base, "traces"), 0755)

	deleteProfileData("p1")

	if _, err := os.Stat(filepath.Join(base, "history")); !os.IsNotExist(err) {
		t.Fatalf("history not deleted")
	}
	if _, err := os.Stat(filepath.Join(base, "traces")); !os.IsNotExist(err) {
		t.Fatalf("traces not deleted")
	}
}

func TestPersistProfileChange(t *testing.T) {
	keyring.MockInit()
	dir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", oldHome)

	profiles := []Profile{}
	p := Profile{Name: "test", Username: "user", Password: "secret"}
	if err := persistProfileChange(&profiles, "test", p, -1); err != nil {
		t.Fatalf("persistProfileChange: %v", err)
	}

	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(profiles))
	}
	if profiles[0].Password != "keyring:emqutiti-test/user" {
		t.Fatalf("password not rewritten: %v", profiles[0].Password)
	}

	cfgPath, _ := DefaultUserConfigFile()
	var cfg userConfig
	if _, err := toml.DecodeFile(cfgPath, &cfg); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(cfg.Profiles) != 1 || cfg.Profiles[0].Password != "keyring:emqutiti-test/user" {
		t.Fatalf("config not written: %#v", cfg)
	}
	if cfg.DefaultProfileName != "test" {
		t.Fatalf("default not saved: %#v", cfg)
	}
	pw, err := keyring.Get("emqutiti-test", "user")
	if err != nil || pw != "secret" {
		t.Fatalf("keyring not saved: %q %v", pw, err)
	}
}
func TestDefaultPasswordEnvOverride(t *testing.T) {
	p := Profile{Name: "test", Password: "orig", FromEnv: false}
	os.Setenv("EMQUTITI_DEFAULT_PASSWORD", "envpw")
	t.Cleanup(func() { os.Unsetenv("EMQUTITI_DEFAULT_PASSWORD") })
	if env := os.Getenv("EMQUTITI_DEFAULT_PASSWORD"); env != "" && !p.FromEnv {
		p.Password = env
	}
	if p.Password != "envpw" {
		t.Fatalf("expected override, got %q", p.Password)
	}
}
