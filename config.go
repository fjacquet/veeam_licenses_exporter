package main

import (
	core "github.com/fjacquet/licenses-exporter-core"
	"github.com/fjacquet/veeam_licenses_exporter/internal/veeam"
)

// Config is the exporter's full config: the shared core.Base (collection + otlp)
// inline, plus the vendor-specific veeam block.
type Config struct {
	core.Base `yaml:",inline"`
	Veeam     veeam.VeeamConfig `yaml:"veeam"`
}

// loadConfig parses the file and builds the sources — the single closure body
// core.Main calls at startup and on every reload.
func loadConfig(path string, trace bool) (core.Base, []core.Source, error) {
	var cfg Config
	if err := core.LoadYAML(path, &cfg); err != nil {
		return core.Base{}, nil, err
	}
	if err := cfg.Validate(); err != nil {
		return core.Base{}, nil, err
	}
	sources, err := veeam.NewSources(cfg.Veeam, trace)
	if err != nil {
		return core.Base{}, nil, err
	}
	return cfg.Base, sources, nil
}
