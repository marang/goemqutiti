package emqutiti

import (
	"bytes"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

// connectionSnapshot holds topics and payloads for a connection in config.toml.
type connectionSnapshot struct {
	Topics   []TopicSnapshot   `toml:"topics"`
	Payloads []PayloadSnapshot `toml:"payloads"`
}

type persistedTrace struct {
	Profile string   `toml:"profile"`
	Topics  []string `toml:"topics"`
	Start   string   `toml:"start"`
	End     string   `toml:"end"`
}

// userConfig represents the structure stored in config.toml.
type userConfig struct {
	DefaultProfileName string                        `toml:"default_profile"`
	Profiles           []Profile                     `toml:"profiles"`
	Saved              map[string]connectionSnapshot `toml:"saved"`
	Traces             map[string]persistedTrace     `toml:"traces"`
}

// loadState retrieves saved topics and payloads from config.toml.
func loadState() map[string]connectionSnapshot {
	fp, err := DefaultUserConfigFile()
	if err != nil {
		return map[string]connectionSnapshot{}
	}
	var cfg userConfig
	if _, err := toml.DecodeFile(fp, &cfg); err != nil {
		return map[string]connectionSnapshot{}
	}
	if cfg.Saved == nil {
		return map[string]connectionSnapshot{}
	}
	return cfg.Saved
}

// writeConfig writes the entire configuration back to disk.
func writeConfig(cfg userConfig) {
	fp, err := DefaultUserConfigFile()
	if err != nil {
		return
	}
	os.MkdirAll(filepath.Dir(fp), os.ModePerm)
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(cfg); err != nil {
		return
	}
	os.WriteFile(fp, buf.Bytes(), 0644)
}

// saveState updates only the Saved section in config.toml.
func saveState(data map[string]connectionSnapshot) {
	fp, err := DefaultUserConfigFile()
	if err != nil {
		return
	}
	var cfg userConfig
	toml.DecodeFile(fp, &cfg) // ignore errors for new files
	cfg.Saved = data
	writeConfig(cfg)
}

// loadTraces retrieves planned traces from config.toml.
func loadTraces() map[string]TracerConfig {
	fp, err := DefaultUserConfigFile()
	if err != nil {
		return map[string]TracerConfig{}
	}
	var cfg userConfig
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
	fp, err := DefaultUserConfigFile()
	if err != nil {
		return
	}
	var cfg userConfig
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
	writeConfig(cfg)
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
