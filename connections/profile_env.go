package connections

import (
	"os"
	"reflect"
	"strconv"
	"strings"
)

// sanitizeEnvName converts a profile name to a valid environment variable name
// by replacing non-alphanumeric characters with underscores and converting to uppercase.
func sanitizeEnvName(name string) string {
	upper := strings.ToUpper(name)
	var b strings.Builder
	for _, r := range upper {
		if r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	return b.String()
}

// EnvPrefix returns the prefix used for environment variables derived from a
// profile name.
func EnvPrefix(name string) string { return "EMQUTITI_" + sanitizeEnvName(name) + "_" }

// ApplyEnvVars loads profile fields from environment variables when FromEnv is
// set. It looks for variables with the prefix
// EMQUTITI_<PROFILENAME>_<FIELD> where <PROFILENAME> is the profile name
// converted to uppercase with non-alphanumeric characters replaced by
// underscores. For backward compatibility, it also checks the old GOEMQUTITI_
// prefix.
func ApplyEnvVars(p *Profile) {
	if !p.FromEnv {
		return
	}

	prefix := EnvPrefix(p.Name)
	vars := make(map[string]string)
	for _, env := range os.Environ() {
		switch {
		case strings.HasPrefix(env, prefix):
			kv := strings.SplitN(env[len(prefix):], "=", 2)
			if len(kv) == 2 {
				vars[strings.ToLower(kv[0])] = kv[1]
			}
		case strings.HasPrefix(env, "GO"+prefix):
			kv := strings.SplitN(env[len(prefix)+2:], "=", 2)
			if len(kv) == 2 {
				vars[strings.ToLower(kv[0])] = kv[1]
			}
		}
	}

	pv := reflect.ValueOf(p).Elem()
	pt := pv.Type()
	for i := 0; i < pt.NumField(); i++ {
		f := pt.Field(i)
		tag := f.Tag.Get("env")
		if tag == "" {
			continue
		}
		val, ok := vars[tag]
		if !ok {
			continue
		}
		field := pv.Field(i)
		switch field.Kind() {
		case reflect.String:
			field.SetString(val)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if iv, err := strconv.Atoi(val); err == nil {
				field.SetInt(int64(iv))
			}
		case reflect.Bool:
			if bv, err := strconv.ParseBool(val); err == nil {
				field.SetBool(bv)
			}
		}
	}
}
