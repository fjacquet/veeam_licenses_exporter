package veeam

// VeeamConfig is the Veeam block of the exporter config. Enabled=false (or an
// empty Servers list) yields zero sources — the exporter serves only
// license_build_info. Each server is a Veeam Backup Enterprise Manager host.
type VeeamConfig struct {
	Enabled bool           `yaml:"enabled"`
	Servers []ServerConfig `yaml:"servers"`
}

// ServerConfig is one Enterprise Manager target. Password is an inline ${ENV}
// ref; PasswordFile is a path read at load (ResolveSecret governs precedence).
type ServerConfig struct {
	Instance           string `yaml:"instance"`
	Host               string `yaml:"host"` // https://em-host:9398
	Username           string `yaml:"username"`
	Password           string `yaml:"password"`
	PasswordFile       string `yaml:"passwordFile"`
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify"`
}
