package main

import (
	"os"
	"path/filepath"
	"testing"

	core "github.com/fjacquet/licenses-exporter-core"
)

func TestLoadConfigParsesBaseAndVeeam(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	yaml := `
collection:
  interval: 3h
otlp:
  endpoint: "otel:4317"
  insecure: true
veeam:
  enabled: true
  servers:
    - instance: em-a
      host: https://em-a.example.com:9398
      username: svc-ro
      password: shhh
      insecureSkipVerify: true
`
	if err := os.WriteFile(path, []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}
	var cfg Config
	if err := core.LoadYAML(path, &cfg); err != nil {
		t.Fatalf("LoadYAML: %v", err)
	}
	if cfg.Collection.Interval.Hours() != 3 {
		t.Errorf("interval = %v, want 3h", cfg.Collection.Interval)
	}
	if cfg.OTLP.Endpoint != "otel:4317" {
		t.Errorf("otlp endpoint = %q, want otel:4317", cfg.OTLP.Endpoint)
	}
	if !cfg.Veeam.Enabled || len(cfg.Veeam.Servers) != 1 || cfg.Veeam.Servers[0].Instance != "em-a" {
		t.Errorf("veeam block not parsed: %+v", cfg.Veeam)
	}
}

func TestLoadReturnsSourcesForEnabledServer(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	yaml := `
collection:
  interval: 2h
veeam:
  enabled: true
  servers:
    - instance: em-a
      host: https://em-a.example.com:9398
      username: svc-ro
      password: shhh
`
	if err := os.WriteFile(path, []byte(yaml), 0o600); err != nil {
		t.Fatal(err)
	}
	base, sources, err := loadConfig(path, false)
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if base.Collection.Interval.Hours() != 2 {
		t.Errorf("interval = %v, want 2h", base.Collection.Interval)
	}
	if len(sources) != 1 || sources[0].Vendor() != "veeam" || sources[0].Instance() != "em-a" {
		t.Fatalf("sources = %+v, want one veeam/em-a", sources)
	}
}
