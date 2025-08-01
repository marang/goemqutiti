package main

import "os"

// OverridePasswordFromEnv applies environment variables when FromEnv is set
// and otherwise overrides the password using the MQTT_PASSWORD variable.
func OverridePasswordFromEnv(p *Profile) {
	if p.FromEnv {
		ApplyEnvVars(p)
	} else if env := os.Getenv("MQTT_PASSWORD"); env != "" {
		p.Password = env
	}
}
