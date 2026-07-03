package veeam

import (
	"encoding/json"
	"fmt"
	"time"

	core "github.com/fjacquet/licenses-exporter-core"
)

const (
	vendor = "veeam"
	unit   = "instances"
)

// parseLicense maps an Enterprise Manager /api/licensing response to license
// samples. Tolerant / absent-not-zero: unlimited (LicensedInstancesNumber<=0)
// omits seats_total; perpetual / unparseable ExpirationDate omits the expiration
// sample. A malformed body is an error so the source degrades to license_up=0.
func parseLicense(raw []byte, instance string) ([]core.Sample, error) {
	var li licenseInfo
	if err := json.Unmarshal(raw, &li); err != nil {
		return nil, fmt.Errorf("decode licensing response: %w", err)
	}
	product := li.Edition
	if product == "" {
		product = vendor
	}
	var out []core.Sample
	if li.LicensedInstancesNumber != nil && *li.LicensedInstancesNumber > 0 {
		out = append(out, core.SeatSample(core.MetricSeatsTotal, vendor, product, unit, instance, float64(*li.LicensedInstancesNumber)))
	}
	if li.UsedInstancesNumber != nil {
		out = append(out, core.SeatSample(core.MetricSeatsUsed, vendor, product, unit, instance, float64(*li.UsedInstancesNumber)))
	}
	if li.ExpirationDate != "" {
		if t, err := time.Parse(time.RFC3339, li.ExpirationDate); err == nil {
			out = append(out, core.ExpirationSample(vendor, product, instance, float64(t.Unix())))
		}
	}
	return out, nil
}
