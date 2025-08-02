package main

import (
	"testing"

	"github.com/zalando/go-keyring"
)

func TestSavePasswordToKeyring(t *testing.T) {
	keyring.MockInit()
	if err := savePasswordToKeyring("svc", "user", "pw"); err != nil {
		t.Fatalf("savePasswordToKeyring: %v", err)
	}
	got, err := keyring.Get("emqutiti-svc", "user")
	if err != nil || got != "pw" {
		t.Fatalf("got %q err %v", got, err)
	}
}

func TestRetrievePasswordFromKeyring(t *testing.T) {
	keyring.MockInit()
	if err := keyring.Set("svc", "user", "secret"); err != nil {
		t.Fatalf("keyring set: %v", err)
	}
	pw, err := RetrievePasswordFromKeyring("keyring:svc/user")
	if err != nil {
		t.Fatalf("RetrievePasswordFromKeyring: %v", err)
	}
	if pw != "secret" {
		t.Fatalf("expected secret, got %q", pw)
	}
}

func TestRetrievePasswordFromKeyringInvalid(t *testing.T) {
	if _, err := RetrievePasswordFromKeyring("plain"); err == nil {
		t.Fatal("expected error for non-keyring password")
	}
}
