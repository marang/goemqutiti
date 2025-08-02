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

func TestProfileBrokerURL(t *testing.T) {
	p := Profile{Schema: "mqtt", Host: "example.com", Port: 1883}
	if got := p.BrokerURL(); got != "mqtt://example.com:1883" {
		t.Fatalf("BrokerURL() = %q", got)
	}
}

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
