package files

import (
	"os"
	"path/filepath"
)

// DataDir returns the base data directory for the given profile.
// If the profile is empty, "default" is used.
// The directory is placed under ~/.emqutiti/data.
func DataDir(profile string) string {
	if profile == "" {
		profile = "default"
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("data", profile)
	}
	return filepath.Join(home, ".emqutiti", "data", profile)
}

// EnsureDir creates the directory with 0755 permissions if it does not exist.
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}
