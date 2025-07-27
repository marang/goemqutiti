package importer

import (
	"encoding/json"
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

func TestRowToJSON(t *testing.T) {
	row := map[string]string{"A": "1", "B": "2"}
	mapping := map[string]string{"A": "alpha", "B": ""}
	data, err := RowToJSON(row, mapping)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var got map[string]string
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	if got["alpha"] != "1" || got["B"] != "2" {
		t.Fatalf("got %v", got)
	}
}
