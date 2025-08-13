package emqutiti

import (
	"fmt"

	"github.com/marang/emqutiti/connections"
	"github.com/marang/emqutiti/constants"
	"github.com/marang/emqutiti/importer"
)

// initImporter bootstraps the importer for a selected profile.
func initImporter(m *model) error {
	if importFile == "" {
		return nil
	}
	var p *connections.Profile
	if profileName != "" {
		for i := range m.connections.Manager.Profiles {
			if m.connections.Manager.Profiles[i].Name == profileName {
				p = &m.connections.Manager.Profiles[i]
				break
			}
		}
	} else if m.connections.Manager.DefaultProfileName != "" {
		for i := range m.connections.Manager.Profiles {
			if m.connections.Manager.Profiles[i].Name == m.connections.Manager.DefaultProfileName {
				p = &m.connections.Manager.Profiles[i]
				break
			}
		}
	}
	if p == nil && len(m.connections.Manager.Profiles) > 0 {
		p = &m.connections.Manager.Profiles[0]
	}
	if p == nil {
		return nil
	}
	cfg := *p
	if cfg.FromEnv {
		connections.ApplyEnvVars(&cfg)
	}
	connections.ApplyDefaultPassword(&cfg)
	client, err := NewMQTTClient(cfg, nil)
	if err != nil {
		return fmt.Errorf("connect error: %w", err)
	}
	m.mqttClient = client
	m.connections.Active = cfg.Name
	m.importer = importer.New(client, importFile)
	m.components[constants.ModeImporter] = m.importer
	m.SetMode(constants.ModeImporter)
	return nil
}
