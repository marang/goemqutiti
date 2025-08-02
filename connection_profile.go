package emqutiti

import (
	"fmt"
	"strings"

	"github.com/zalando/go-keyring"
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
