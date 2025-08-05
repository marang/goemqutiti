package traces

import (
	"bytes"
	"log"
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
func loadTraces() (out map[string]TracerConfig) {
	out = map[string]TracerConfig{}
	fp, err := connections.DefaultUserConfigFile()
	if err != nil {
		return
	}
	var cfg struct {
		Traces map[string]persistedTrace `toml:"traces"`
	}
	if _, err := toml.DecodeFile(fp, &cfg); err != nil {
		return
	}
	for k, v := range cfg.Traces {
		var start, end time.Time
		if v.Start != "" {
			t, err := time.Parse(time.RFC3339, v.Start)
			if err != nil {
				log.Printf("invalid start time for trace %q: %v", k, err)
			} else {
				start = t
			}
		}
		if v.End != "" {
			t, err := time.Parse(time.RFC3339, v.End)
			if err != nil {
				log.Printf("invalid end time for trace %q: %v", k, err)
			} else {
				end = t
			}
		}
		out[k] = TracerConfig{Profile: v.Profile, Topics: v.Topics, Start: start, End: end, Key: k}
	}
	return
}

// saveTraces updates the Traces section in config.toml.
func saveTraces(data map[string]TracerConfig) error {
	fp, err := connections.DefaultUserConfigFile()
	if err != nil {
		return err
	}
	cfg := map[string]interface{}{}
	toml.DecodeFile(fp, &cfg) // ignore errors for new files
	traces := map[string]interface{}{}
	for k, v := range data {
		sub := map[string]interface{}{"profile": v.Profile, "topics": v.Topics}
		if !v.Start.IsZero() {
			sub["start"] = v.Start.Format(time.RFC3339)
		}
		if !v.End.IsZero() {
			sub["end"] = v.End.Format(time.RFC3339)
		}
		traces[k] = sub
	}
	cfg["traces"] = traces
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(cfg); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(fp), os.ModePerm); err != nil {
		return err
	}
	if err := os.WriteFile(fp, buf.Bytes(), 0644); err != nil {
		return err
	}
	return nil
}

// addTrace merges a single trace configuration into the existing file.
func addTrace(cfg TracerConfig) error {
	traces := loadTraces()
	traces[cfg.Key] = cfg
	return saveTraces(traces)
}

// removeTrace deletes a trace from the configuration file.
func removeTrace(key string) error {
	traces := loadTraces()
	delete(traces, key)
	return saveTraces(traces)
}
