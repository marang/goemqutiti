package main

import (
	"os"
	"testing"
)

func TestNewConnectionFormEnvReadOnly(t *testing.T) {
	os.Setenv("GOEMQUTITI_ENV_HOST", "envhost")
	os.Setenv("GOEMQUTITI_ENV_PORT", "1884")
	t.Cleanup(func() {
		os.Unsetenv("GOEMQUTITI_ENV_HOST")
		os.Unsetenv("GOEMQUTITI_ENV_PORT")
	})
	cf := newConnectionForm(Profile{Name: "env", FromEnv: true}, 0)
	hostField, ok := cf.fields[fieldIndex["Host"]].(*textField)
	if !ok || hostField.Value() != "envhost" {
		t.Fatalf("host not loaded from env: %v", hostField)
	}
	portField := cf.fields[fieldIndex["Port"]].(*textField)
	if portField.Value() != "1884" {
		t.Fatalf("port not loaded from env: %s", portField.Value())
	}
	idxName := fieldIndex["Name"]
	idxFromEnv := fieldIndex["FromEnv"]
	for i, fld := range cf.fields {
		if i == idxName || i == idxFromEnv {
			continue
		}
		switch f := fld.(type) {
		case *textField:
			if !f.readOnly {
				t.Errorf("field %s not read-only", formFields[i].key)
			}
		case *selectField:
			if !f.readOnly {
				t.Errorf("field %s not read-only", formFields[i].key)
			}
		case *checkField:
			if !f.readOnly {
				t.Errorf("field %s not read-only", formFields[i].key)
			}
		}
	}
}

func TestNewConnectionFormPasswordPlaceholder(t *testing.T) {
	cf := newConnectionForm(Profile{Name: "test", Username: "user", Password: "secret"}, 0)
	tf, ok := cf.fields[fieldIndex["Password"]].(*textField)
	if !ok {
		t.Fatalf("password field not text")
	}
	if tf.Model.Placeholder != "keyring:emqutiti-test/user" {
		t.Fatalf("unexpected placeholder %s", tf.Model.Placeholder)
	}
}

func TestConnectionFormProfile(t *testing.T) {
	cf := newConnectionForm(Profile{}, -1)
	cf.fields[fieldIndex["Name"]].(*textField).SetValue("n1")
	cf.fields[fieldIndex["Port"]].(*textField).SetValue("1883")
	cf.fields[fieldIndex["AutoReconnect"]].(*checkField).value = true
	cf.fields[fieldIndex["QoS"]].(*selectField).index = 2

	p, err := cf.Profile()
	if err != nil {
		t.Fatalf("profile: %v", err)
	}
	if p.Name != "n1" || p.Port != 1883 || !p.AutoReconnect || p.QoS != 2 {
		t.Fatalf("unexpected profile: %#v", p)
	}
}

func TestConnectionFormProfileInvalidInt(t *testing.T) {
	cf := newConnectionForm(Profile{}, -1)
	cf.fields[fieldIndex["Port"]].(*textField).SetValue("abc")
	p, err := cf.Profile()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if p.Port != 0 {
		t.Fatalf("expected port 0, got %d", p.Port)
	}
}
