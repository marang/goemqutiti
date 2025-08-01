package main

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

// TestApplyEnvVars sets env vars for all Profile fields and ensures they are applied.
func TestApplyEnvVars(t *testing.T) {
	p := Profile{Name: "test", FromEnv: true}
	prefix := EnvPrefix(p.Name)

	rt := reflect.TypeOf(p)
	rv := reflect.ValueOf(&p).Elem()

	var envs []string
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if f.Name == "FromEnv" {
			continue
		}
		tag := f.Tag.Get("toml")
		if tag == "" {
			continue
		}
		envName := prefix + strings.ToUpper(strings.ReplaceAll(tag, "-", "_"))
		switch f.Type.Kind() {
		case reflect.String:
			os.Setenv(envName, "x")
			envs = append(envs, envName)
		case reflect.Int:
			os.Setenv(envName, "1")
			envs = append(envs, envName)
		case reflect.Bool:
			os.Setenv(envName, "true")
			envs = append(envs, envName)
		default:
			t.Fatalf("unsupported kind %s", f.Type.Kind())
		}
	}
	t.Cleanup(func() {
		for _, e := range envs {
			os.Unsetenv(e)
		}
	})

	ApplyEnvVars(&p)

	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if f.Name == "FromEnv" {
			continue
		}
		tag := f.Tag.Get("toml")
		if tag == "" {
			continue
		}
		field := rv.Field(i)
		switch f.Type.Kind() {
		case reflect.String:
			if field.String() != "x" {
				t.Errorf("field %s not set", f.Name)
			}
		case reflect.Int:
			if field.Int() != 1 {
				t.Errorf("field %s not set", f.Name)
			}
		case reflect.Bool:
			if field.Bool() != true {
				t.Errorf("field %s not set", f.Name)
			}
		}
	}
}
