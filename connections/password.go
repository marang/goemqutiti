package connections

import "os"

// ApplyDefaultPassword assigns the EMQUTITI_DEFAULT_PASSWORD environment variable
// to the profile's password when the profile is not loaded from the environment
// and has no existing password.
func ApplyDefaultPassword(p *Profile) {
	if p == nil {
		return
	}
	if !p.FromEnv && p.Password == "" {
		if env := os.Getenv("EMQUTITI_DEFAULT_PASSWORD"); env != "" {
			p.Password = env
		}
	}
}
