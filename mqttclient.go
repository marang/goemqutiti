package main

import (
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTClient struct {
	Client mqtt.Client
}

func NewMQTTClient(broker, clientId, username string, password string) (*MQTTClient, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientId)
	opts.SetUsername(username)

	// Convert password to string only when setting it in the options
	if len(password) > 0 {
		opts.SetPassword(password)
	}

	opts.SetCleanSession(true)
	opts.OnConnect = func(client mqtt.Client) {
		log.Println("Connected to MQTT broker")
	}
	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		log.Printf("Connection lost: %v", err)
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("failed to connect: %w", token.Error())
	}

	return &MQTTClient{Client: client}, nil
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
