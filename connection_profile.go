package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/zalando/go-keyring"

	"github.com/marang/goemqutiti/internal/files"
)

// Profile defines a broker connection.
type Profile struct {
	Name                string `toml:"name"`
	Schema              string `toml:"schema"`
	Host                string `toml:"host"`
	Port                int    `toml:"port"`
	ClientID            string `toml:"client_id"`
	Username            string `toml:"username"`
	Password            string `toml:"password"`
	FromEnv             bool   `toml:"from_env"`
	SSL                 bool   `toml:"ssl_tls"`
	MQTTVersion         string `toml:"mqtt_version"`
	ConnectTimeout      int    `toml:"connect_timeout"`
	KeepAlive           int    `toml:"keep_alive"`
	QoS                 int    `toml:"qos"`
	AutoReconnect       bool   `toml:"auto_reconnect"`
	ReconnectPeriod     int    `toml:"reconnect_period"`
	CleanStart          bool   `toml:"clean_start"`
	SessionExpiry       int    `toml:"session_expiry_interval"`
	ReceiveMaximum      int    `toml:"receive_maximum"`
	MaximumPacketSize   int    `toml:"maximum_packet_size"`
	TopicAliasMaximum   int    `toml:"topic_alias_maximum"`
	RequestResponseInfo bool   `toml:"request_response_info"`
	RequestProblemInfo  bool   `toml:"request_problem_info"`
	LastWillEnabled     bool   `toml:"last_will_enabled"`
	LastWillTopic       string `toml:"last_will_topic"`
	LastWillQos         int    `toml:"last_will_qos"`
	LastWillRetain      bool   `toml:"last_will_retain"`
	LastWillPayload     string `toml:"last_will_payload"`
	RandomIDSuffix      bool   `toml:"random_id_suffix"`
}

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

// RetrievePasswordFromKeyring resolves a keyring:<service>/<user> reference.
func RetrievePasswordFromKeyring(password string) (string, error) {
	if !strings.HasPrefix(password, "keyring:") {
		return "", fmt.Errorf("password does not reference keyring")
	}
	parts := strings.SplitN(password, ":", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid keyring reference: %s", password)
	}
	serviceUsername := strings.SplitN(parts[1], "/", 2)
	if len(serviceUsername) != 2 {
		return "", fmt.Errorf("invalid keyring format: %s", parts[1])
	}
	pw, err := keyring.Get(serviceUsername[0], serviceUsername[1])
	if err != nil {
		return "", fmt.Errorf("failed to retrieve password from keyring for %s/%s: %w", serviceUsername[0], serviceUsername[1], err)
	}
	return pw, nil
}

func sanitizeEnvName(name string) string {
	upper := strings.ToUpper(name)
	var b strings.Builder
	for _, r := range upper {
		if r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	return b.String()
}

// EnvPrefix returns the prefix used for environment variables derived from a
// profile name.
func EnvPrefix(name string) string { return "GOEMQUTITI_" + sanitizeEnvName(name) + "_" }

type profileEnvSetter func(*Profile, string)

var profileEnvSetters = map[string]profileEnvSetter{
	"name":   func(p *Profile, v string) { p.Name = v },
	"schema": func(p *Profile, v string) { p.Schema = v },
	"host":   func(p *Profile, v string) { p.Host = v },
	"port": func(p *Profile, v string) {
		if iv, err := strconv.Atoi(v); err == nil {
			p.Port = iv
		}
	},
	"client_id": func(p *Profile, v string) { p.ClientID = v },
	"username":  func(p *Profile, v string) { p.Username = v },
	"password":  func(p *Profile, v string) { p.Password = v },
	"ssl_tls": func(p *Profile, v string) {
		if bv, err := strconv.ParseBool(v); err == nil {
			p.SSL = bv
		}
	},
	"mqtt_version": func(p *Profile, v string) { p.MQTTVersion = v },
	"connect_timeout": func(p *Profile, v string) {
		if iv, err := strconv.Atoi(v); err == nil {
			p.ConnectTimeout = iv
		}
	},
	"keep_alive": func(p *Profile, v string) {
		if iv, err := strconv.Atoi(v); err == nil {
			p.KeepAlive = iv
		}
	},
	"qos": func(p *Profile, v string) {
		if iv, err := strconv.Atoi(v); err == nil {
			p.QoS = iv
		}
	},
	"auto_reconnect": func(p *Profile, v string) {
		if bv, err := strconv.ParseBool(v); err == nil {
			p.AutoReconnect = bv
		}
	},
	"reconnect_period": func(p *Profile, v string) {
		if iv, err := strconv.Atoi(v); err == nil {
			p.ReconnectPeriod = iv
		}
	},
	"clean_start": func(p *Profile, v string) {
		if bv, err := strconv.ParseBool(v); err == nil {
			p.CleanStart = bv
		}
	},
	"session_expiry_interval": func(p *Profile, v string) {
		if iv, err := strconv.Atoi(v); err == nil {
			p.SessionExpiry = iv
		}
	},
	"receive_maximum": func(p *Profile, v string) {
		if iv, err := strconv.Atoi(v); err == nil {
			p.ReceiveMaximum = iv
		}
	},
	"maximum_packet_size": func(p *Profile, v string) {
		if iv, err := strconv.Atoi(v); err == nil {
			p.MaximumPacketSize = iv
		}
	},
	"topic_alias_maximum": func(p *Profile, v string) {
		if iv, err := strconv.Atoi(v); err == nil {
			p.TopicAliasMaximum = iv
		}
	},
	"request_response_info": func(p *Profile, v string) {
		if bv, err := strconv.ParseBool(v); err == nil {
			p.RequestResponseInfo = bv
		}
	},
	"request_problem_info": func(p *Profile, v string) {
		if bv, err := strconv.ParseBool(v); err == nil {
			p.RequestProblemInfo = bv
		}
	},
	"last_will_enabled": func(p *Profile, v string) {
		if bv, err := strconv.ParseBool(v); err == nil {
			p.LastWillEnabled = bv
		}
	},
	"last_will_topic": func(p *Profile, v string) { p.LastWillTopic = v },
	"last_will_qos": func(p *Profile, v string) {
		if iv, err := strconv.Atoi(v); err == nil {
			p.LastWillQos = iv
		}
	},
	"last_will_retain": func(p *Profile, v string) {
		if bv, err := strconv.ParseBool(v); err == nil {
			p.LastWillRetain = bv
		}
	},
	"last_will_payload": func(p *Profile, v string) { p.LastWillPayload = v },
	"random_id_suffix": func(p *Profile, v string) {
		if bv, err := strconv.ParseBool(v); err == nil {
			p.RandomIDSuffix = bv
		}
	},
}

// ApplyEnvVars loads profile fields from environment variables when FromEnv is set.
func ApplyEnvVars(p *Profile) {
	if !p.FromEnv {
		return
	}
	prefix := EnvPrefix(p.Name)
	for tag, setter := range profileEnvSetters {
		envName := prefix + strings.ToUpper(strings.ReplaceAll(tag, "-", "_"))
		if val, ok := os.LookupEnv(envName); ok {
			setter(p, val)
		}
	}
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

// saveConfig persists profiles and default selection to config.toml.
func saveConfig(profiles []Profile, defaultName string) {
	saved := loadState()
	cfg := userConfig{
		DefaultProfileName: defaultName,
		Profiles:           profiles,
		Saved:              make(map[string]persistedConn),
	}
	for k, v := range saved {
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

// savePasswordToKeyring stores a password in the system keyring.
func savePasswordToKeyring(service, username, password string) {
	if err := keyring.Set("emqutiti-"+service, username, password); err != nil {
		fmt.Println("Error saving password to keyring:", err)
	}
}

// deleteProfileData removes profile-specific persisted history and traces and
// returns any cleanup errors.
func deleteProfileData(name string) error {
	historyPath := filepath.Join(files.DataDir(name), "history")
	historyErr := os.RemoveAll(historyPath)
	if historyErr != nil {
		log.Printf("Error removing %s: %v", historyPath, historyErr)
	}

	tracesPath := filepath.Join(files.DataDir(name), "traces")
	tracesErr := os.RemoveAll(tracesPath)
	if tracesErr != nil {
		log.Printf("Error removing %s: %v", tracesPath, tracesErr)
	}

	return errors.Join(historyErr, tracesErr)
}

// persistProfileChange applies a profile update, saves config and keyring.
func persistProfileChange(profiles *[]Profile, defaultName string, p Profile, idx int) {
	plain := p.Password
	if !p.FromEnv {
		p.Password = "keyring:emqutiti-" + p.Name + "/" + p.Username
	} else {
		p.Password = ""
	}
	if idx >= 0 && idx < len(*profiles) {
		(*profiles)[idx] = p
	} else {
		*profiles = append(*profiles, p)
	}
	saveConfig(*profiles, defaultName)
	if !p.FromEnv {
		savePasswordToKeyring(p.Name, p.Username, plain)
	}
}
