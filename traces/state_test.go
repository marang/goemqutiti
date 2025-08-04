package traces

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/marang/emqutiti/connections"
)

func TestSaveTracesPreservesExistingConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	cfgPath, err := connections.DefaultUserConfigFile()
	if err != nil {
		t.Fatalf("cfg path: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(cfgPath, []byte("default_profile = \"foo\"\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	data := map[string]TracerConfig{
		"t1": {Profile: "p1", Topics: []string{"a"}, Start: time.Unix(0, 0)},
	}
	saveTraces(data)
	out, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, "default_profile") {
		t.Fatalf("missing existing fields: %s", s)
	}
	if !strings.Contains(s, "[traces]") {
		t.Fatalf("missing traces section: %s", s)
	}
}
