package main

import (
	"fmt"
	"strings"

	"github.com/zalando/go-keyring"
)

// RetrievePasswordFromKeyring resolves a keyring:<service>/<user> reference.
func RetrievePasswordFromKeyring(password string) (string, error) {
	if !strings.HasPrefix(password, "keyring:") {
		return "", fmt.Errorf("password does not reference keyring")
	}
	parts := strings.SplitN(password, ":", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid keyring reference: %s", password)
	}
	serviceUsername := strings.SplitN(parts[1], "/", 2)
	if len(serviceUsername) != 2 {
		return "", fmt.Errorf("invalid keyring format: %s", parts[1])
	}
	pw, err := keyring.Get(serviceUsername[0], serviceUsername[1])
	if err != nil {
		return "", fmt.Errorf("failed to retrieve password from keyring for %s/%s: %w", serviceUsername[0], serviceUsername[1], err)
	}
	return pw, nil
}

// savePasswordToKeyring stores a password in the system keyring.
func savePasswordToKeyring(service, username, password string) error {
	return keyring.Set("emqutiti-"+service, username, password)
}
