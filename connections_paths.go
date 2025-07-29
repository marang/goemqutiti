package main

import (
	"os"
	"path/filepath"
)

// dataDir returns the base data directory for the given profile.
func dataDir(profile string) string {
	if profile == "" {
		profile = "default"
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("data", profile)
	}
	return filepath.Join(home, ".emqutiti", "data", profile)
}
