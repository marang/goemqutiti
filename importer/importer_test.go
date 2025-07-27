package importer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadCSVBuildTopic(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "data.csv")
	err := os.WriteFile(file, []byte("serial_number,payload\n123,hello\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := ReadFile(file)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	topic := BuildTopic("device/{serial_number}/status", rows[0])
	if topic != "device/123/status" {
		t.Fatalf("got topic %s", topic)
	}
	if rows[0]["payload"] != "hello" {
		t.Fatalf("payload mismatch")
	}
}
