package tracer

import (
	"crypto/tls"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/marang/goemqutiti/config"
)

type Profile = config.Profile

// mqttClient wraps the MQTT connection for the tracer.
type mqttClient struct{ client mqtt.Client }

func newMQTTClient(p Profile) (*mqttClient, error) {
	opts := mqtt.NewClientOptions()
	brokerURL := fmt.Sprintf("%s://%s:%d", p.Schema, p.Host, p.Port)
	opts.AddBroker(brokerURL)
	cid := p.ClientID
	if p.RandomIDSuffix {
		cid = fmt.Sprintf("%s-%d", cid, time.Now().UnixNano())
	}
	opts.SetClientID(cid)
	opts.SetUsername(p.Username)
	if p.Password != "" {
		opts.SetPassword(p.Password)
	}
	if p.MQTTVersion != "" {
		var ver uint
		fmt.Sscan(p.MQTTVersion, &ver)
		if ver != 0 {
			opts.SetProtocolVersion(ver)
		}
	}
	if p.ConnectTimeout > 0 {
		opts.SetConnectTimeout(time.Duration(p.ConnectTimeout) * time.Second)
	}
	if p.KeepAlive > 0 {
		opts.SetKeepAlive(time.Duration(p.KeepAlive) * time.Second)
	}
	opts.SetAutoReconnect(p.AutoReconnect)
	if p.CleanStart {
		opts.SetCleanSession(true)
	} else {
		opts.SetCleanSession(false)
	}
	if p.LastWillEnabled && p.LastWillTopic != "" {
		opts.SetWill(p.LastWillTopic, p.LastWillPayload, byte(p.LastWillQos), p.LastWillRetain)
	}
	if p.SSL {
		opts.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
	}
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("failed to connect: %w", token.Error())
	}
	return &mqttClient{client: client}, nil
}

func (m *mqttClient) Subscribe(topic string, qos byte, cb mqtt.MessageHandler) error {
	token := m.client.Subscribe(topic, qos, cb)
	token.Wait()
	return token.Error()
}
func (m *mqttClient) Unsubscribe(topic string) error {
	token := m.client.Unsubscribe(topic)
	token.Wait()
	return token.Error()
}
func (m *mqttClient) Disconnect() {
	if m.client != nil && m.client.IsConnected() {
		m.client.Disconnect(250)
	}
}

// Run executes the tracer headlessly using configuration from config.toml.
func Run(key, topics, profileName, startStr, endStr string) error {
	if key == "" || topics == "" {
		return fmt.Errorf("-trace and -topics are required")
	}
	var start, end time.Time
	var err error
	if startStr != "" {
		start, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			return fmt.Errorf("invalid start time: %w", err)
		}
	}
	if endStr != "" {
		end, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			return fmt.Errorf("invalid end time: %w", err)
		}
	}
	p, err := config.LoadProfile(profileName, "")
	if err != nil {
		return err
	}
	if env := os.Getenv("MQTT_PASSWORD"); env != "" && !p.FromEnv {
		p.Password = env
	}
	client, err := newMQTTClient(*p)
	if err != nil {
		return fmt.Errorf("connect error: %w", err)
	}
	defer client.Disconnect()

	tlist := strings.Split(topics, ",")
	for i := range tlist {
		tlist[i] = strings.TrimSpace(tlist[i])
	}
	cfg := Config{Profile: p.Name, Topics: tlist, Start: start, End: end, Key: key}
	tr := New(cfg, client)
	if err := tr.Start(); err != nil {
		return fmt.Errorf("trace start: %w", err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	for tr.Planned() || tr.Running() {
		select {
		case <-sig:
			tr.Stop()
		case <-time.After(500 * time.Millisecond):
		}
	}

	for t, c := range tr.Counts() {
		fmt.Printf("%s: %d\n", t, c)
	}
	return nil
}
