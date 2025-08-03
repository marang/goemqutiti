package traces

import (
	"bytes"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/marang/emqutiti/connections"
)

type persistedTrace struct {
	Profile string   `toml:"profile"`
	Topics  []string `toml:"topics"`
	Start   string   `toml:"start"`
	End     string   `toml:"end"`
}

// loadTraces retrieves planned traces from config.toml.
func loadTraces() map[string]TracerConfig {
	fp, err := connections.DefaultUserConfigFile()
	if err != nil {
		return map[string]TracerConfig{}
	}
	var cfg struct {
		Traces map[string]persistedTrace `toml:"traces"`
	}
	if _, err := toml.DecodeFile(fp, &cfg); err != nil {
		return map[string]TracerConfig{}
	}
	out := make(map[string]TracerConfig)
	for k, v := range cfg.Traces {
		var start, end time.Time
		if v.Start != "" {
			start, _ = time.Parse(time.RFC3339, v.Start)
		}
		if v.End != "" {
			end, _ = time.Parse(time.RFC3339, v.End)
		}
		out[k] = TracerConfig{Profile: v.Profile, Topics: v.Topics, Start: start, End: end, Key: k}
	}
	return out
}

// saveTraces updates the Traces section in config.toml.
func saveTraces(data map[string]TracerConfig) {
	fp, err := connections.DefaultUserConfigFile()
	if err != nil {
		return
	}
	var cfg struct {
		Traces map[string]persistedTrace `toml:"traces"`
	}
	toml.DecodeFile(fp, &cfg) // ignore errors for new files
	cfg.Traces = make(map[string]persistedTrace)
	for k, v := range data {
		pt := persistedTrace{Profile: v.Profile, Topics: v.Topics}
		if !v.Start.IsZero() {
			pt.Start = v.Start.Format(time.RFC3339)
		}
		if !v.End.IsZero() {
			pt.End = v.End.Format(time.RFC3339)
		}
		cfg.Traces[k] = pt
	}
	var buf bytes.Buffer
	toml.NewEncoder(&buf).Encode(cfg)
	os.MkdirAll(filepath.Dir(fp), os.ModePerm)
	os.WriteFile(fp, buf.Bytes(), 0644)
}

// addTrace merges a single trace configuration into the existing file.
func addTrace(cfg TracerConfig) {
	traces := loadTraces()
	traces[cfg.Key] = cfg
	saveTraces(traces)
}

// removeTrace deletes a trace from the configuration file.
func removeTrace(key string) {
	traces := loadTraces()
	delete(traces, key)
	saveTraces(traces)
}
