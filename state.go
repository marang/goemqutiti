package main

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
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

// userConfig represents the structure stored in config.toml.
type userConfig struct {
	DefaultProfileName string                   `toml:"default_profile"`
	Profiles           []Profile                `toml:"profiles"`
	Saved              map[string]persistedConn `toml:"saved"`
}

// loadState retrieves saved topics and payloads from config.toml.
func loadState() map[string]connectionData {
	fp, err := DefaultUserConfigFile()
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
			topics = append(topics, topicItem{title: t.Title, active: t.Active})
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
func saveState(data map[string]connectionData) {
	fp, err := DefaultUserConfigFile()
	if err != nil {
		return
	}
	var cfg userConfig
	toml.DecodeFile(fp, &cfg) // ignore errors for new files
	cfg.Saved = make(map[string]persistedConn)
	for k, v := range data {
		var topics []persistedTopic
		for _, t := range v.Topics {
			topics = append(topics, persistedTopic{Title: t.title, Active: t.active})
		}
		var payloads []persistedPayload
		for _, p := range v.Payloads {
			payloads = append(payloads, persistedPayload{Topic: p.topic, Payload: p.payload})
		}
		cfg.Saved[k] = persistedConn{Topics: topics, Payloads: payloads}
	}
	writeConfig(cfg)
}
