package connections

import (
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"

	"github.com/marang/emqutiti/internal/files"
)

// saveConfig persists profiles and default selection to config.toml.
func saveConfig(profiles []Profile, defaultName string) {
	saved := LoadState()
	cfg := userConfig{
		DefaultProfileName: defaultName,
		Profiles:           profiles,
		Saved:              saved,
	}
	writeConfig(cfg)
}

// savePasswordToKeyring stores a password in the system keyring.
func savePasswordToKeyring(service, username, password string) error {
	return keyring.Set("emqutiti-"+service, username, password)
}

// deleteProfileData removes profile-specific persisted history and traces and
// returns any cleanup errors.
func deleteProfileData(name string) error {
	historyPath := filepath.Join(files.DataDir(name), "history")
	historyErr := os.RemoveAll(historyPath)
	if historyErr != nil {
		log.Printf("Error removing %s: %v", historyPath, historyErr)
	}

	tracesPath := filepath.Join(files.DataDir(name), "traces")
	tracesErr := os.RemoveAll(tracesPath)
	if tracesErr != nil {
		log.Printf("Error removing %s: %v", tracesPath, tracesErr)
	}

	return errors.Join(historyErr, tracesErr)
}

// persistProfileChange applies a profile update, saves config and keyring.
func persistProfileChange(profiles *[]Profile, defaultName string, p Profile, idx int) error {
	plain := p.Password
	if !p.FromEnv {
		p.Password = "keyring:emqutiti-" + p.Name + "/" + p.Username
	} else {
		p.Password = ""
	}
	if idx >= 0 && idx < len(*profiles) {
		(*profiles)[idx] = p
	} else {
		*profiles = append(*profiles, p)
	}
	saveConfig(*profiles, defaultName)
	if !p.FromEnv {
		if err := savePasswordToKeyring(p.Name, p.Username, plain); err != nil {
			return err
		}
	}
	return nil
}
