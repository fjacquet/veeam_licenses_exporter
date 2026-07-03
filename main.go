package main

import (
	core "github.com/fjacquet/licenses-exporter-core"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// version is injected via -ldflags at build time (see Makefile `make cli`).
var version = "dev"

func main() {
	var (
		cfgPath string
		addr    string
		debug   bool
		once    bool
		trace   bool
	)
	root := &cobra.Command{
		Use:   "veeam_licenses_exporter",
		Short: "Veeam (Enterprise Manager) license Prometheus + OTLP exporter",
		RunE: func(_ *cobra.Command, _ []string) error {
			return core.Main(core.App{
				Version:    version,
				Addr:       addr,
				Once:       once,
				Debug:      debug,
				Trace:      trace,
				ConfigPath: cfgPath,
				Load:       func() (core.Base, []core.Source, error) { return loadConfig(cfgPath) },
			})
		},
	}
	root.Flags().StringVar(&cfgPath, "config", "config.yaml", "path to config.yaml")
	root.Flags().StringVar(&addr, "web.listen-address", ":9107", "metrics listen address")
	root.Flags().BoolVar(&debug, "debug", false, "debug logging")
	root.Flags().BoolVar(&once, "once", false, "run one collection cycle and exit")
	root.Flags().BoolVar(&trace, "trace", false, "log repo-owned API responses (session token redacted; SDK tracing n/a)")
	if err := root.Execute(); err != nil {
		logrus.WithError(err).Fatal("exporter failed")
	}
}
