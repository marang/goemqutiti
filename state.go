package main

import (
	"bytes"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/marang/goemqutiti/config"
)

// persistedTopic mirrors topicItem for persistence in the config file.
type persistedTopic struct {
	Title  string `toml:"title"`
	Active bool   `toml:"active"`
}

type persistedPayload struct {
	Topic   string `toml:"topic"`
	Payload string `toml:"payload"`
}

type persistedConn struct {
	Topics   []persistedTopic   `toml:"topics"`
	Payloads []persistedPayload `toml:"payloads"`
}

type persistedTrace struct {
	Profile string   `toml:"profile"`
	Topics  []string `toml:"topics"`
	Start   string   `toml:"start"`
	End     string   `toml:"end"`
}

// userConfig represents the structure stored in config.toml.
type userConfig struct {
	DefaultProfileName string                    `toml:"default_profile"`
	Profiles           []Profile                 `toml:"profiles"`
	Saved              map[string]persistedConn  `toml:"saved"`
	Traces             map[string]persistedTrace `toml:"traces"`
}

// loadState retrieves saved topics and payloads from config.toml.
func loadState() map[string]connectionData {
	fp, err := config.DefaultUserConfigFile()
	if err != nil {
		return map[string]connectionData{}
	}
	var cfg userConfig
	if _, err := toml.DecodeFile(fp, &cfg); err != nil {
		return map[string]connectionData{}
	}
	out := make(map[string]connectionData)
	for k, v := range cfg.Saved {
		var topics []topicItem
		for _, t := range v.Topics {
			topics = append(topics, topicItem{title: t.Title, subscribed: t.Active})
		}
		var payloads []payloadItem
		for _, p := range v.Payloads {
			payloads = append(payloads, payloadItem{topic: p.Topic, payload: p.Payload})
		}
		out[k] = connectionData{Topics: topics, Payloads: payloads}
	}
	return out
}

// writeConfig writes the entire configuration back to disk.
func writeConfig(cfg userConfig) {
	fp, err := config.DefaultUserConfigFile()
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
func saveState(data map[string]connectionData) {
	fp, err := config.DefaultUserConfigFile()
	if err != nil {
		return
	}
	var cfg userConfig
	toml.DecodeFile(fp, &cfg) // ignore errors for new files
	cfg.Saved = make(map[string]persistedConn)
	for k, v := range data {
		var topics []persistedTopic
		for _, t := range v.Topics {
			topics = append(topics, persistedTopic{Title: t.title, Active: t.subscribed})
		}
		var payloads []persistedPayload
		for _, p := range v.Payloads {
			payloads = append(payloads, persistedPayload{Topic: p.topic, Payload: p.payload})
		}
		cfg.Saved[k] = persistedConn{Topics: topics, Payloads: payloads}
	}
	writeConfig(cfg)
}

// loadTraces retrieves planned traces from config.toml.
func loadTraces() map[string]TracerConfig {
	fp, err := config.DefaultUserConfigFile()
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
	fp, err := config.DefaultUserConfigFile()
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
