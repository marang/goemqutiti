package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/zalando/go-keyring"
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

func envPrefix(name string) string { return "GOEMQUTITI_" + sanitizeEnvName(name) + "_" }

// ApplyEnvVars loads profile fields from environment variables when FromEnv is set.
func ApplyEnvVars(p *Profile) {
	if !p.FromEnv {
		return
	}
	prefix := envPrefix(p.Name)
	rv := reflect.ValueOf(p).Elem()
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if f.Name == "FromEnv" {
			continue
		}
		tag := f.Tag.Get("toml")
		if tag == "" {
			continue
		}
		envName := prefix + strings.ToUpper(strings.ReplaceAll(tag, "-", "_"))
		val, ok := os.LookupEnv(envName)
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
