package connections

import (
	"os"
	"strconv"
	"strings"

	mapstructure "github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
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

type profileEnvSetter func(*Profile, string)

// profileEnvSetters handles bespoke conversions where default decoding is not
// sufficient. Most fields are handled directly by mapstructure.
var profileEnvSetters = map[string]profileEnvSetter{
	"port": func(p *Profile, v string) {
		if iv, err := strconv.Atoi(v); err == nil {
			p.Port = iv
		}
	},
	"ssl_tls": func(p *Profile, v string) {
		if bv, err := strconv.ParseBool(v); err == nil {
			p.SSL = bv
		}
	},
	"skip_tls_verify": func(p *Profile, v string) {
		if bv, err := strconv.ParseBool(v); err == nil {
			p.SkipTLSVerify = bv
		}
	},
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
	"publish_timeout": func(p *Profile, v string) {
		if iv, err := strconv.Atoi(v); err == nil {
			p.PublishTimeout = iv
		}
	},
	"subscribe_timeout": func(p *Profile, v string) {
		if iv, err := strconv.Atoi(v); err == nil {
			p.SubscribeTimeout = iv
		}
	},
	"unsubscribe_timeout": func(p *Profile, v string) {
		if iv, err := strconv.Atoi(v); err == nil {
			p.UnsubscribeTimeout = iv
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
	"random_id_suffix": func(p *Profile, v string) {
		if bv, err := strconv.ParseBool(v); err == nil {
			p.RandomIDSuffix = bv
		}
	},
}

// ApplyEnvVars loads profile fields from environment variables when FromEnv is set.
// It looks for variables with the prefix EMQUTITI_<PROFILENAME>_<FIELD> where
// <PROFILENAME> is the profile name converted to uppercase with non-alphanumeric
// characters replaced by underscores.
// For backward compatibility, it also checks the old GOEMQUTITI_ prefix.
func ApplyEnvVars(p *Profile) {
	if !p.FromEnv {
		return
	}

	v := viper.New()
	v.SetEnvPrefix(EnvPrefix(p.Name))
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()

	prefix := EnvPrefix(p.Name)
	for _, env := range os.Environ() {
		switch {
		case strings.HasPrefix(env, prefix):
			kv := strings.SplitN(env[len(prefix):], "=", 2)
			if len(kv) == 2 {
				v.Set(strings.ToLower(kv[0]), kv[1])
			}
		case strings.HasPrefix(env, "GO"+prefix):
			kv := strings.SplitN(env[len(prefix)+2:], "=", 2)
			if len(kv) == 2 {
				v.Set(strings.ToLower(kv[0]), kv[1])
			}
		}
	}

	_ = v.Unmarshal(p, func(dc *mapstructure.DecoderConfig) {
		dc.TagName = "env"
		dc.ZeroFields = false
	})

	for tag, setter := range profileEnvSetters {
		if v.IsSet(tag) {
			setter(p, v.GetString(tag))
		}
	}
}
