package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTClient struct {
	Client     mqtt.Client
	StatusChan chan string
}

func NewMQTTClient(p Profile, statusChan chan string) (*MQTTClient, error) {
	opts := mqtt.NewClientOptions()
	brokerURL := fmt.Sprintf("%s://%s:%d", p.Schema, p.Host, p.Port)
	opts.AddBroker(brokerURL)
	opts.SetClientID(p.ClientID)
	opts.SetUsername(p.Username)

	if len(p.Password) > 0 {
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
	opts.OnConnect = func(client mqtt.Client) {
		log.Println("Connected to MQTT broker")
		if statusChan != nil {
			statusChan <- "Connected to MQTT broker"
		}
	}
	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		log.Printf("Connection lost: %v", err)
		if statusChan != nil {
			statusChan <- fmt.Sprintf("Connection lost: %v", err)
		}
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("failed to connect: %w", token.Error())
	}

	return &MQTTClient{Client: client, StatusChan: statusChan}, nil
}

func (m *MQTTClient) Publish(topic string, qos byte, retained bool, payload interface{}) error {
	token := m.Client.Publish(topic, qos, retained, payload)
	token.Wait()
	if token.Error() != nil {
		return fmt.Errorf("publish failed: %w", token.Error())
	}
	return nil
}

func (m *MQTTClient) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) error {
	token := m.Client.Subscribe(topic, qos, callback)
	token.Wait()
	if token.Error() != nil {
		return fmt.Errorf("subscribe failed: %w", token.Error())
	}
	return nil
}

func (m *MQTTClient) Unsubscribe(topic string) error {
	token := m.Client.Unsubscribe(topic)
	token.Wait()
	if token.Error() != nil {
		return fmt.Errorf("unsubscribe failed: %w", token.Error())
	}
	return nil
}
