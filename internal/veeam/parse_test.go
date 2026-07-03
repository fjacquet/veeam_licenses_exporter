package veeam

import (
	"testing"

	core "github.com/fjacquet/licenses-exporter-core"
)

func find(samples []core.Sample, name string) (core.Sample, bool) {
	for _, s := range samples {
		if s.Name == name {
			return s, true
		}
	}
	return core.Sample{}, false
}

// TestParseLimitedLicense: a limited, dated license emits seats_total, seats_used,
// and expiration, all labelled vendor=veeam / unit=instances / product=<Edition>.
func TestParseLimitedLicense(t *testing.T) {
	raw := []byte(`{"Edition":"Enterprise Plus","Status":"Valid","ExpirationDate":"2027-01-31T00:00:00Z","LicensedInstancesNumber":100,"UsedInstancesNumber":42}`)
	samples, err := parseLicense(raw, "em-a")
	if err != nil {
		t.Fatalf("parseLicense: %v", err)
	}
	total, ok := find(samples, core.MetricSeatsTotal)
	if !ok || total.Value != 100 {
		t.Fatalf("seats_total = %+v ok=%v, want 100", total, ok)
	}
	used, ok := find(samples, core.MetricSeatsUsed)
	if !ok || used.Value != 42 {
		t.Fatalf("seats_used = %+v ok=%v, want 42", used, ok)
	}
	exp, ok := find(samples, core.MetricExpiration)
	if !ok || exp.Value != 1801353600 { // 2027-01-31T00:00:00Z
		t.Fatalf("expiration = %+v ok=%v, want 1801353600", exp, ok)
	}
	// label check: vendor/unit/product on seats_total
	got := map[string]string{}
	for _, l := range total.Labels {
		got[l.Key] = l.Value
	}
	if got["vendor"] != "veeam" || got["unit"] != "instances" || got["product"] != "Enterprise Plus" || got["instance"] != "em-a" {
		t.Fatalf("labels = %v, want veeam/instances/Enterprise Plus/em-a", got)
	}
}

// TestParseUnlimitedOmitsTotal: LicensedInstancesNumber <= 0 omits seats_total.
func TestParseUnlimitedOmitsTotal(t *testing.T) {
	raw := []byte(`{"Edition":"Community","ExpirationDate":"2027-01-31T00:00:00Z","LicensedInstancesNumber":0,"UsedInstancesNumber":3}`)
	samples, err := parseLicense(raw, "em-a")
	if err != nil {
		t.Fatalf("parseLicense: %v", err)
	}
	if _, ok := find(samples, core.MetricSeatsTotal); ok {
		t.Fatal("seats_total must be omitted when LicensedInstancesNumber<=0")
	}
	if used, ok := find(samples, core.MetricSeatsUsed); !ok || used.Value != 3 {
		t.Fatalf("seats_used = %+v ok=%v, want 3", used, ok)
	}
}

// TestParsePerpetualOmitsExpiration: empty/absent ExpirationDate omits the
// expiration sample (perpetual), never emits a fake value.
func TestParsePerpetualOmitsExpiration(t *testing.T) {
	raw := []byte(`{"Edition":"Enterprise","LicensedInstancesNumber":50,"UsedInstancesNumber":10}`)
	samples, err := parseLicense(raw, "em-a")
	if err != nil {
		t.Fatalf("parseLicense: %v", err)
	}
	if _, ok := find(samples, core.MetricExpiration); ok {
		t.Fatal("expiration must be omitted when ExpirationDate is absent")
	}
}

// TestParseInvalidJSON: malformed body is an error (source degrades to up=0).
func TestParseInvalidJSON(t *testing.T) {
	if _, err := parseLicense([]byte(`not json`), "em-a"); err == nil {
		t.Fatal("expected error on malformed JSON")
	}
}
