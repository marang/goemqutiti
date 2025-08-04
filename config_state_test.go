package emqutiti

import (
	connections "github.com/marang/emqutiti/connections"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/marang/emqutiti/traces"
)

func TestLoadFromConfig(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config.toml")
	data := `default_profile = "local"

[[profiles]]
name = "local"
schema = "tcp"
host = "localhost"
port = 1883
username = "user"
password = "secret"
`
	if err := os.WriteFile(cfg, []byte(data), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	c, err := connections.LoadFromConfig(cfg)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if len(c.Profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(c.Profiles))
	}
	p := c.Profiles[0]
	if p.Name != "local" || p.Password != "secret" || p.Port != 1883 {
		t.Fatalf("unexpected profile: %+v", p)
	}
}

func TestLoadFromConfigEnv(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "config.toml")
	data := `[[profiles]]
name = "test"
from_env = true
`
	if err := os.WriteFile(cfg, []byte(data), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	os.Setenv("EMQUTITI_TEST_HOST", "example.com")
	os.Setenv("EMQUTITI_TEST_PORT", "1884")
	os.Setenv("EMQUTITI_TEST_SCHEMA", "ssl")
	defer func() {
		os.Unsetenv("EMQUTITI_TEST_HOST")
		os.Unsetenv("EMQUTITI_TEST_PORT")
		os.Unsetenv("EMQUTITI_TEST_SCHEMA")
	}()

	c, err := connections.LoadFromConfig(cfg)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if len(c.Profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(c.Profiles))
	}
	p := c.Profiles[0]
	if p.Host != "example.com" || p.Port != 1884 || p.Schema != "ssl" {
		t.Fatalf("env vars not applied: %+v", p)
	}
}

func TestSaveLoadState(t *testing.T) {
	dir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", oldHome)

	data := map[string]connections.ConnectionSnapshot{
		"p1": {
			Topics:   []connections.TopicSnapshot{{Title: "foo", Subscribed: true, Publish: true}},
			Payloads: []connections.PayloadSnapshot{{Topic: "foo", Payload: "bar"}},
		},
	}
	connections.SaveState(data)
	got := connections.LoadState()
	if !reflect.DeepEqual(got, data) {
		t.Fatalf("state mismatch: %#v != %#v", got, data)
	}
}

func TestSaveLoadTraces(t *testing.T) {
	dir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", oldHome)

	start := time.Date(2025, time.July, 28, 18, 25, 21, 0, time.UTC)
	end := start.Add(time.Hour)
	data := map[string]traces.TracerConfig{
		"t1": {Profile: "p1", Topics: []string{"a"}, Start: start, End: end, Key: "t1"},
	}
	store := traces.FileStore{}
	store.SaveTraces(data)
	got := store.LoadTraces()
	// Compare only relevant fields
	if len(got) != len(data) {
		t.Fatalf("trace count mismatch")
	}
	for k, v := range data {
		g := got[k]
		if g.Profile != v.Profile || len(g.Topics) != len(v.Topics) || g.Topics[0] != v.Topics[0] || !g.Start.Equal(v.Start) || !g.End.Equal(v.End) {
			t.Fatalf("trace mismatch: %#v != %#v", g, v)
		}
	}
}
