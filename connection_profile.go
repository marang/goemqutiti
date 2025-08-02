package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/zalando/go-keyring"

	"github.com/marang/goemqutiti/internal/files"
)

// Profile defines a broker connection.
type Profile struct {
	Name                string `toml:"name" env:"name"`
	Schema              string `toml:"schema" env:"schema"`
	Host                string `toml:"host" env:"host"`
	Port                int    `toml:"port" env:"port"`
	ClientID            string `toml:"client_id" env:"client_id"`
	Username            string `toml:"username" env:"username"`
	Password            string `toml:"password" env:"password"`
	FromEnv             bool   `toml:"from_env"`
	SSL                 bool   `toml:"ssl_tls" env:"ssl_tls"`
	MQTTVersion         string `toml:"mqtt_version" env:"mqtt_version"`
	ConnectTimeout      int    `toml:"connect_timeout" env:"connect_timeout"`
	KeepAlive           int    `toml:"keep_alive" env:"keep_alive"`
	QoS                 int    `toml:"qos" env:"qos"`
	AutoReconnect       bool   `toml:"auto_reconnect" env:"auto_reconnect"`
	ReconnectPeriod     int    `toml:"reconnect_period" env:"reconnect_period"`
	CleanStart          bool   `toml:"clean_start" env:"clean_start"`
	SessionExpiry       int    `toml:"session_expiry_interval" env:"session_expiry_interval"`
	ReceiveMaximum      int    `toml:"receive_maximum" env:"receive_maximum"`
	MaximumPacketSize   int    `toml:"maximum_packet_size" env:"maximum_packet_size"`
	TopicAliasMaximum   int    `toml:"topic_alias_maximum" env:"topic_alias_maximum"`
	RequestResponseInfo bool   `toml:"request_response_info" env:"request_response_info"`
	RequestProblemInfo  bool   `toml:"request_problem_info" env:"request_problem_info"`
	LastWillEnabled     bool   `toml:"last_will_enabled" env:"last_will_enabled"`
	LastWillTopic       string `toml:"last_will_topic" env:"last_will_topic"`
	LastWillQos         int    `toml:"last_will_qos" env:"last_will_qos"`
	LastWillRetain      bool   `toml:"last_will_retain" env:"last_will_retain"`
	LastWillPayload     string `toml:"last_will_payload" env:"last_will_payload"`
	RandomIDSuffix      bool   `toml:"random_id_suffix" env:"random_id_suffix"`
}

// BrokerURL returns the formatted broker URL.
func (p Profile) BrokerURL() string {
	return fmt.Sprintf("%s://%s:%d", p.Schema, p.Host, p.Port)
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
func EnvPrefix(name string) string { return "EMQUTITI_" + sanitizeEnvName(name) + "_" }

func applyEnvTags(prefix, oldPrefix string, v any) {
	rv := reflect.ValueOf(v).Elem()
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		tag := f.Tag.Get("env")
		if tag == "" {
			continue
		}
		key := strings.ToUpper(strings.ReplaceAll(tag, "-", "_"))
		envName := prefix + key
		val, ok := os.LookupEnv(envName)
		if !ok {
			val, ok = os.LookupEnv(oldPrefix + key)
		}
		if !ok {
			continue
		}
		field := rv.Field(i)
		switch field.Kind() {
		case reflect.String:
			field.SetString(val)
		case reflect.Int:
			if iv, err := strconv.Atoi(val); err == nil {
				field.SetInt(int64(iv))
			}
		case reflect.Bool:
			if bv, err := strconv.ParseBool(val); err == nil {
				field.SetBool(bv)
			}
		}
	}
}

// ApplyEnvVars loads profile fields from environment variables when FromEnv is set.
func ApplyEnvVars(p *Profile) {
	if !p.FromEnv {
		return
	}
	prefix := EnvPrefix(p.Name)
	// For a limited time, also check the old GOEMQUTITI_ prefix for
	// backward compatibility.
	oldPrefix := "GO" + prefix
	applyEnvTags(prefix, oldPrefix, p)
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
func savePasswordToKeyring(service, username, password string) error {
	return keyring.Set("emqutiti-"+service, username, password)
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
func persistProfileChange(profiles *[]Profile, defaultName string, p Profile, idx int) error {
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
		if err := savePasswordToKeyring(p.Name, p.Username, plain); err != nil {
			return err
		}
	}
	return nil
}
