package main

import (
	"os"
	"testing"
)

func TestOverridePasswordFromEnv_FromEnv(t *testing.T) {
	p := &Profile{Name: "demo", FromEnv: true}
	os.Setenv("GOEMQUTITI_DEMO_PASSWORD", "envpass")
	os.Setenv("MQTT_PASSWORD", "global")
	defer os.Unsetenv("GOEMQUTITI_DEMO_PASSWORD")
	defer os.Unsetenv("MQTT_PASSWORD")

	OverridePasswordFromEnv(p)
	if p.Password != "envpass" {
		t.Fatalf("expected envpass, got %s", p.Password)
	}
}

func TestOverridePasswordFromEnv_Global(t *testing.T) {
	p := &Profile{Name: "demo"}
	os.Setenv("MQTT_PASSWORD", "global")
	defer os.Unsetenv("MQTT_PASSWORD")

	OverridePasswordFromEnv(p)
	if p.Password != "global" {
		t.Fatalf("expected global, got %s", p.Password)
	}
}

func TestOverridePasswordFromEnv_NoChange(t *testing.T) {
	p := &Profile{Name: "demo", FromEnv: true}
	os.Setenv("MQTT_PASSWORD", "global")
	defer os.Unsetenv("MQTT_PASSWORD")

	OverridePasswordFromEnv(p)
	if p.Password != "" {
		t.Fatalf("expected empty password, got %s", p.Password)
	}
}
