package veeam

import (
	"context"

	core "github.com/fjacquet/licenses-exporter-core"
	"github.com/sirupsen/logrus"
)

type source struct {
	instance string
	host     string
	username string
	password string
	insecure bool
}

func (s *source) Vendor() string   { return vendor }
func (s *source) Instance() string { return s.instance }

// Collect logs into Enterprise Manager, reads the licensing resource, logs out
// (best-effort), and parses the result — stateless per cycle. A logout failure is
// warned, not fatal, so operators see potential EM session leaks.
func (s *source) Collect(ctx context.Context) ([]core.Sample, error) {
	c := newEMClient(s.host, s.insecure)
	if err := c.login(ctx, s.username, s.password); err != nil {
		return nil, err
	}
	defer func() {
		if err := c.logout(); err != nil {
			logrus.WithFields(logrus.Fields{"vendor": vendor, "instance": s.instance}).WithError(err).Warn("veeam EM logout failed")
		}
	}()
	raw, err := c.licensing(ctx)
	if err != nil {
		return nil, err
	}
	return parseLicense(raw, s.instance)
}
