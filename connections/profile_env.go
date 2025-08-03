package connections

import (
	"os"
	"strconv"
	"strings"
)

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

type profileEnvSetter func(*Profile, string)

var profileEnvSetters = map[string]profileEnvSetter{
	"name":   func(p *Profile, v string) { p.Name = v },
	"schema": func(p *Profile, v string) { p.Schema = v },
	"host":   func(p *Profile, v string) { p.Host = v },
	"port": func(p *Profile, v string) {
		if iv, err := strconv.Atoi(v); err == nil {
			p.Port = iv
		}
	},
	"client_id": func(p *Profile, v string) { p.ClientID = v },
	"username":  func(p *Profile, v string) { p.Username = v },
	"password":  func(p *Profile, v string) { p.Password = v },
	"ssl_tls": func(p *Profile, v string) {
		if bv, err := strconv.ParseBool(v); err == nil {
			p.SSL = bv
		}
	},
	"mqtt_version": func(p *Profile, v string) { p.MQTTVersion = v },
	"connect_timeout": func(p *Profile, v string) {
		if iv, err := strconv.Atoi(v); err == nil {
			p.ConnectTimeout = iv
		}
	},
	"keep_alive": func(p *Profile, v string) {
		if iv, err := strconv.Atoi(v); err == nil {
			p.KeepAlive = iv
		}
	},
	"qos": func(p *Profile, v string) {
		if iv, err := strconv.Atoi(v); err == nil {
			p.QoS = iv
		}
	},
	"auto_reconnect": func(p *Profile, v string) {
		if bv, err := strconv.ParseBool(v); err == nil {
			p.AutoReconnect = bv
		}
	},
	"reconnect_period": func(p *Profile, v string) {
		if iv, err := strconv.Atoi(v); err == nil {
			p.ReconnectPeriod = iv
		}
	},
	"clean_start": func(p *Profile, v string) {
		if bv, err := strconv.ParseBool(v); err == nil {
			p.CleanStart = bv
		}
	},
	"session_expiry_interval": func(p *Profile, v string) {
		if iv, err := strconv.Atoi(v); err == nil {
			p.SessionExpiry = iv
		}
	},
	"receive_maximum": func(p *Profile, v string) {
		if iv, err := strconv.Atoi(v); err == nil {
			p.ReceiveMaximum = iv
		}
	},
	"maximum_packet_size": func(p *Profile, v string) {
		if iv, err := strconv.Atoi(v); err == nil {
			p.MaximumPacketSize = iv
		}
	},
	"topic_alias_maximum": func(p *Profile, v string) {
		if iv, err := strconv.Atoi(v); err == nil {
			p.TopicAliasMaximum = iv
		}
	},
	"request_response_info": func(p *Profile, v string) {
		if bv, err := strconv.ParseBool(v); err == nil {
			p.RequestResponseInfo = bv
		}
	},
	"request_problem_info": func(p *Profile, v string) {
		if bv, err := strconv.ParseBool(v); err == nil {
			p.RequestProblemInfo = bv
		}
	},
	"last_will_enabled": func(p *Profile, v string) {
		if bv, err := strconv.ParseBool(v); err == nil {
			p.LastWillEnabled = bv
		}
	},
	"last_will_topic": func(p *Profile, v string) { p.LastWillTopic = v },
	"last_will_qos": func(p *Profile, v string) {
		if iv, err := strconv.Atoi(v); err == nil {
			p.LastWillQos = iv
		}
	},
	"last_will_retain": func(p *Profile, v string) {
		if bv, err := strconv.ParseBool(v); err == nil {
			p.LastWillRetain = bv
		}
	},
	"last_will_payload": func(p *Profile, v string) { p.LastWillPayload = v },
	"random_id_suffix": func(p *Profile, v string) {
		if bv, err := strconv.ParseBool(v); err == nil {
			p.RandomIDSuffix = bv
		}
	},
}

// ApplyEnvVars loads profile fields from environment variables when FromEnv is set.
func ApplyEnvVars(p *Profile) {
	if !p.FromEnv {
		return
	}
	prefix := EnvPrefix(p.Name)
	// For a limited time, also check the old GOEMQUTITI_ prefix for
	// backward compatibility.
	oldPrefix := "GO" + prefix
	for tag, setter := range profileEnvSetters {
		key := strings.ToUpper(strings.ReplaceAll(tag, "-", "_"))
		envName := prefix + key
		if val, ok := os.LookupEnv(envName); ok {
			setter(p, val)
			continue
		}
		if val, ok := os.LookupEnv(oldPrefix + key); ok {
			setter(p, val)
		}
	}
}
