package connections

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// LoadProxyAddr returns the proxy address from config.toml.
func LoadProxyAddr() string {
	fp, err := DefaultUserConfigFile()
	if err != nil {
		return ""
	}
	var cfg struct {
		ProxyAddr string `toml:"proxy_addr"`
	}
	if _, err := toml.DecodeFile(fp, &cfg); err != nil {
		return ""
	}
	return cfg.ProxyAddr
}

// SaveProxyAddr stores the proxy address in config.toml.
func SaveProxyAddr(addr string) error {
	fp, err := DefaultUserConfigFile()
	if err != nil {
		return err
	}
	cfg := map[string]interface{}{}
	toml.DecodeFile(fp, &cfg) // ignore errors
	if addr != "" {
		cfg["proxy_addr"] = addr
	} else {
		delete(cfg, "proxy_addr")
	}
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(cfg); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(fp), os.ModePerm); err != nil {
		return err
	}
	return os.WriteFile(fp, buf.Bytes(), 0644)
}
