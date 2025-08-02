package emqutiti

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	DefaultProfile string    `toml:"default_profile"`
	Profiles       []Profile `toml:"profiles"`
}

// DefaultUserConfigFile returns ~/.emqutiti/config.toml.
func DefaultUserConfigFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".emqutiti", "config.toml"), nil
}

// LoadConfig reads profiles from a TOML file and resolves keyring references.
func LoadConfig(filePath string) (*Config, error) {
	var err error
	if filePath == "" {
		if filePath, err = DefaultUserConfigFile(); err != nil {
			return nil, err
		}
	}
	var cfg Config
	if _, err := toml.DecodeFile(filePath, &cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}
	for i := range cfg.Profiles {
		p := &cfg.Profiles[i]
		if p.FromEnv {
			ApplyEnvVars(p)
			continue
		}
		if strings.HasPrefix(p.Password, "keyring:") {
			pw, err := RetrievePasswordFromKeyring(p.Password)
			if err != nil {
				return nil, err
			}
			p.Password = pw
		}
	}
	return &cfg, nil
}

// LoadProfile returns the named profile from the config file, falling back to the default or first profile.
func LoadProfile(name, file string) (*Profile, error) {
	cfg, err := LoadConfig(file)
	if err != nil {
		return nil, err
	}
	var p *Profile
	if name != "" {
		for i := range cfg.Profiles {
			if cfg.Profiles[i].Name == name {
				p = &cfg.Profiles[i]
				break
			}
		}
	} else if cfg.DefaultProfile != "" {
		for i := range cfg.Profiles {
			if cfg.Profiles[i].Name == cfg.DefaultProfile {
				p = &cfg.Profiles[i]
				break
			}
		}
	}
	if p == nil && len(cfg.Profiles) > 0 {
		p = &cfg.Profiles[0]
	}
	if p == nil {
		return nil, fmt.Errorf("no connection profile available")
	}
	return p, nil
}
