package mqttclient

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// ClientOption configures mqtt.ClientOptions.
type ClientOption func(*mqtt.ClientOptions)

// WithBroker sets the broker URL.
func WithBroker(url string) ClientOption {
	return func(o *mqtt.ClientOptions) {
		o.AddBroker(url)
	}
}

// WithClientID sets the client ID, optionally adding a random suffix.
func WithClientID(id string, random bool) ClientOption {
	return func(o *mqtt.ClientOptions) {
		if random {
			id = fmt.Sprintf("%s-%d", id, time.Now().UnixNano())
		}
		o.SetClientID(id)
	}
}

// WithAuth sets the username and password if provided.
func WithAuth(user, pass string) ClientOption {
	return func(o *mqtt.ClientOptions) {
		o.SetUsername(user)
		if pass != "" {
			o.SetPassword(pass)
		}
	}
}

// WithVersion sets the MQTT protocol version if specified.
func WithVersion(ver string) (ClientOption, error) {
	if ver == "" {
		return func(o *mqtt.ClientOptions) {}, nil
	}
	v, err := strconv.Atoi(ver)
	if err != nil {
		return nil, fmt.Errorf("invalid MQTT version %q: %w", ver, err)
	}
	return func(o *mqtt.ClientOptions) {
		if v != 0 {
			o.SetProtocolVersion(uint(v))
		}
	}, nil
}

// WithTimeouts configures connect timeout and keepalive.
func WithTimeouts(connectTimeout, keepAlive int) ClientOption {
	return func(o *mqtt.ClientOptions) {
		if connectTimeout > 0 {
			o.SetConnectTimeout(time.Duration(connectTimeout) * time.Second)
		}
		if keepAlive > 0 {
			o.SetKeepAlive(time.Duration(keepAlive) * time.Second)
		}
	}
}

// WithSession sets auto reconnect and clean session flags.
func WithSession(autoReconnect, cleanStart bool) ClientOption {
	return func(o *mqtt.ClientOptions) {
		o.SetAutoReconnect(autoReconnect)
		o.SetCleanSession(cleanStart)
	}
}

// WithWill sets the last will message if enabled.
func WithWill(enabled bool, topic, payload string, qos int, retain bool) ClientOption {
	return func(o *mqtt.ClientOptions) {
		if enabled && topic != "" {
			o.SetWill(topic, payload, byte(qos), retain)
		}
	}
}

// WithTLS applies TLS configuration when SSL is enabled.
func WithTLS(ssl bool, skipVerify bool, caCertPath, clientCertPath, clientKeyPath string) (ClientOption, error) {
	if !ssl {
		return func(o *mqtt.ClientOptions) {}, nil
	}
	tlsCfg := &tls.Config{InsecureSkipVerify: skipVerify}
	if caCertPath != "" {
		caData, err := os.ReadFile(caCertPath)
		if err != nil {
			return nil, fmt.Errorf("read CA cert: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(caData) {
			return nil, fmt.Errorf("invalid CA cert")
		}
		tlsCfg.RootCAs = pool
	}
	if clientCertPath != "" && clientKeyPath != "" {
		cert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
		if err != nil {
			return nil, fmt.Errorf("load client cert: %w", err)
		}
		tlsCfg.Certificates = []tls.Certificate{cert}
	}
	return func(o *mqtt.ClientOptions) {
		o.SetTLSConfig(tlsCfg)
	}, nil
}
