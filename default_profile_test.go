package emqutiti

import (
	"testing"

	connections "github.com/marang/emqutiti/connections"
)

// Test that Init connects using the default profile when no profile flag is set.
func TestInitConnectsDefaultProfile(t *testing.T) {
	origProfile := profileName
	profileName = ""
	defer func() { profileName = origProfile }()

	conn := connections.NewConnectionsModel()
	conn.Profiles = []connections.Profile{{Name: "p1", Schema: "tcp", Host: "h", Port: 1883}}
	conn.DefaultProfileName = "p1"

	m, _ := initialModel(&conn)
	cmd := m.Init()
	if m.connections.Manager.Statuses["p1"] != "connecting" {
		t.Fatalf("expected status 'connecting', got %q", m.connections.Manager.Statuses["p1"])
	}
	if cmd == nil {
		t.Fatalf("expected connect command")
	}
}
