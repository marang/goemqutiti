package connections

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// ConnectionSnapshot holds topics and payloads for a connection in config.toml.
type ConnectionSnapshot struct {
	Topics   []TopicSnapshot   `toml:"topics"`
	Payloads []PayloadSnapshot `toml:"payloads"`
}

// userConfig represents the structure stored in config.toml.
type userConfig struct {
	DefaultProfileName string                        `toml:"default_profile"`
	Profiles           []Profile                     `toml:"profiles"`
	Saved              map[string]ConnectionSnapshot `toml:"saved"`
}

// LoadState retrieves saved topics and payloads from config.toml.
func LoadState() map[string]ConnectionSnapshot {
	fp, err := DefaultUserConfigFile()
	if err != nil {
		return map[string]ConnectionSnapshot{}
	}
	var cfg userConfig
	if _, err := toml.DecodeFile(fp, &cfg); err != nil {
		return map[string]ConnectionSnapshot{}
	}
	if cfg.Saved == nil {
		return map[string]ConnectionSnapshot{}
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

// SaveState updates only the Saved section in config.toml.
func SaveState(data map[string]ConnectionSnapshot) {
	fp, err := DefaultUserConfigFile()
	if err != nil {
		return
	}
	var cfg userConfig
	toml.DecodeFile(fp, &cfg) // ignore errors for new files
	cfg.Saved = data
	writeConfig(cfg)
}
