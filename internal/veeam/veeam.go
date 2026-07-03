package veeam

import (
	"fmt"

	core "github.com/fjacquet/licenses-exporter-core"
)

// NewSources builds one Source per configured Enterprise Manager server.
func NewSources(cfg VeeamConfig, trace bool) ([]core.Source, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	var out []core.Source
	for _, s := range cfg.Servers {
		pw, err := core.ResolveSecret(s.Password, s.PasswordFile)
		if err != nil {
			return nil, fmt.Errorf("veeam server %q: %w", s.Instance, err)
		}
		out = append(out, &source{
			instance: s.Instance,
			host:     s.Host,
			username: s.Username,
			password: pw,
			insecure: s.InsecureSkipVerify,
			trace:    trace,
		})
	}
	return out, nil
}
