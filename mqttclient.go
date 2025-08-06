package emqutiti

import (
	"context"
	"crypto/tls"
	"fmt"
	connections "github.com/marang/emqutiti/connections"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const defaultTokenTimeout = 5 * time.Second

type MQTTMessage struct {
	Topic   string
	Payload string
}

type MQTTClient struct {
	Client mqtt.Client
	// MessageChan receives published messages. It is closed when
	// Disconnect is called, so consumers must handle channel
	// closure.
	MessageChan        chan MQTTMessage
	publishTimeout     time.Duration
	subscribeTimeout   time.Duration
	unsubscribeTimeout time.Duration
}

// waitToken blocks until the MQTT token completes or the timeout expires.
// It returns any error from the token or a timeout error.
func waitToken(token mqtt.Token, timeout time.Duration, action string) error {
	if timeout <= 0 {
		timeout = defaultTokenTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan struct{})
	go func() {
		token.Wait()
		close(done)
	}()

	select {
	case <-done:
		if err := token.Error(); err != nil {
			return fmt.Errorf("%s failed: %w", action, err)
		}
		return nil
	case <-ctx.Done():
		return fmt.Errorf("%s timeout after %v", action, timeout)
	}
}

// NewMQTTClient creates and configures a new MQTT client based on the profile
// details. Status updates are delivered via the provided callback.
func NewMQTTClient(p connections.Profile, fn statusFunc) (*MQTTClient, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(p.BrokerURL())
	cid := p.ClientID
	if p.RandomIDSuffix {
		cid = fmt.Sprintf("%s-%d", cid, time.Now().UnixNano())
	}
	opts.SetClientID(cid)
	opts.SetUsername(p.Username)

	if len(p.Password) > 0 {
		opts.SetPassword(p.Password)
	}

	if p.MQTTVersion != "" {
		ver, err := strconv.Atoi(p.MQTTVersion)
		if err != nil {
			return nil, fmt.Errorf("invalid MQTT version %q: %w", p.MQTTVersion, err)
		}
		if ver != 0 {
			opts.SetProtocolVersion(uint(ver))
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
		opts.SetTLSConfig(&tls.Config{InsecureSkipVerify: p.SkipTLSVerify})
	}
	opts.OnConnect = func(client mqtt.Client) {
		if fn != nil {
			fn("Connected to MQTT broker")
		}
	}
	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		if fn != nil {
			fn(fmt.Sprintf("Connection lost: %v", err))
		}
	}

	msgChan := make(chan MQTTMessage, 20)
	opts.SetDefaultPublishHandler(func(client mqtt.Client, m mqtt.Message) {
		msgChan <- MQTTMessage{Topic: m.Topic(), Payload: string(m.Payload())}
	})

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("failed to connect: %w", token.Error())
	}

	pubTimeout := time.Duration(p.PublishTimeout) * time.Second
	subTimeout := time.Duration(p.SubscribeTimeout) * time.Second
	unsubTimeout := time.Duration(p.UnsubscribeTimeout) * time.Second

	return &MQTTClient{
		Client:             client,
		MessageChan:        msgChan,
		publishTimeout:     pubTimeout,
		subscribeTimeout:   subTimeout,
		unsubscribeTimeout: unsubTimeout,
	}, nil
}

// Publish sends the payload to the given topic using the underlying client.
// It waits for the publish token to complete and returns any error from the
// broker.
func (m *MQTTClient) Publish(topic string, qos byte, retained bool, payload interface{}) error {
	token := m.Client.Publish(topic, qos, retained, payload)
	return waitToken(token, m.publishTimeout, "publish")
}

// Subscribe registers callback for messages on topic at the specified QoS.
// The method blocks until the broker acknowledges the subscription and
// returns an error if the request fails.
func (m *MQTTClient) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) error {
	token := m.Client.Subscribe(topic, qos, callback)
	return waitToken(token, m.subscribeTimeout, "subscribe")
}

// Unsubscribe removes the subscription for the topic. It waits for
// completion and returns an error if the unsubscribe request fails.
func (m *MQTTClient) Unsubscribe(topic string) error {
	token := m.Client.Unsubscribe(topic)
	return waitToken(token, m.unsubscribeTimeout, "unsubscribe")
}

// Disconnect cleanly closes the connection to the broker and closes
// MessageChan. Consumers must handle channel closure.
func (m *MQTTClient) Disconnect() {
	if m.Client != nil && m.Client.IsConnected() {
		// Allow up to 250 milliseconds for pending work to complete.
		m.Client.Disconnect(250)
	}
	if m.MessageChan != nil {
		close(m.MessageChan)
		m.MessageChan = nil
	}
}
