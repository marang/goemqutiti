package main

import (
	"testing"

	"github.com/zalando/go-keyring"
)

func TestKeyring(t *testing.T) {
	service := "TestService"
	username := "testuser"
	password := "testpassword"

	if err := keyring.Set(service, username, password); err != nil {
		t.Fatal("setup error:", err)
	}

	t.Cleanup(func() {
		_ = keyring.Delete(service, username)
	})

	retrievedPassword, err := keyring.Get(service, username)
	if err != nil {
		t.Fatal("get error:", err)
	}
	if retrievedPassword != password {
		t.Fatalf("expected %q, got %q", password, retrievedPassword)
	}

	if err := keyring.Delete(service, username); err != nil {
		t.Fatal("delete error:", err)
	}

	if _, err := keyring.Get(service, username); err == nil {
		t.Fatal("expected error retrieving deleted password")
	}
}
