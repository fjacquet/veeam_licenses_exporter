package veeam

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	core "github.com/fjacquet/licenses-exporter-core"
	"github.com/sirupsen/logrus"
)

// fakeEM stands up an Enterprise Manager the client can talk to: a session login
// that returns the X-RestSvcSessionId header, a licensing endpoint, and a logout.
func fakeEM(t *testing.T, licenseJSON string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/sessionMngr/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if _, _, ok := r.BasicAuth(); !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("X-RestSvcSessionId", "sess-123")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"SessionId":"sess-123"}`))
	})
	mux.HandleFunc("/api/licensing", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-RestSvcSessionId") != "sess-123" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(licenseJSON))
	})
	mux.HandleFunc("/api/logonSessions/", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	return httptest.NewServer(mux)
}

func TestSourceCollectAgainstFakeEM(t *testing.T) {
	srv := fakeEM(t, `{"Edition":"Enterprise Plus","ExpirationDate":"2027-01-31T00:00:00Z","LicensedInstancesNumber":100,"UsedInstancesNumber":42}`)
	defer srv.Close()

	src := &source{instance: "em-a", host: srv.URL, username: "u", password: "p", insecure: true}
	samples, err := src.Collect(context.Background())
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}
	if _, ok := find(samples, core.MetricSeatsUsed); !ok {
		t.Fatal("expected seats_used from fake EM")
	}
	if total, ok := find(samples, core.MetricSeatsTotal); !ok || total.Value != 100 {
		t.Fatalf("seats_total = %+v ok=%v, want 100", total, ok)
	}
}

// A 401 on licensing (bad session) surfaces as an error so the engine sets up=0.
func TestSourceCollectAuthFailure(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/sessionMngr/", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusUnauthorized) })
	srv := httptest.NewServer(mux)
	defer srv.Close()

	src := &source{instance: "em-a", host: srv.URL, username: "u", password: "p", insecure: true}
	if _, err := src.Collect(context.Background()); err == nil {
		t.Fatal("expected error when session login fails")
	}
}

// NewSources builds one source per enabled server; disabled yields none.
func TestNewSources(t *testing.T) {
	got, err := NewSources(VeeamConfig{Enabled: true, Servers: []ServerConfig{{Instance: "em-a", Host: "https://em:9398", Username: "u", Password: "p"}}}, false)
	if err != nil {
		t.Fatalf("NewSources: %v", err)
	}
	if len(got) != 1 || got[0].Vendor() != "veeam" || got[0].Instance() != "em-a" {
		t.Fatalf("sources = %+v, want one veeam/em-a", got)
	}
	none, err := NewSources(VeeamConfig{Enabled: false}, false)
	if err != nil || none != nil {
		t.Fatalf("disabled NewSources = %v, %v; want nil,nil", none, err)
	}
}

func TestTraceLogsLicensingBodyNotToken(t *testing.T) {
	srv := fakeEM(t, `{"Edition":"Enterprise Plus","ExpirationDate":"2027-01-31T00:00:00Z","LicensedInstancesNumber":100,"UsedInstancesNumber":42}`)
	defer srv.Close()

	var buf bytes.Buffer
	old := logrus.StandardLogger().Out
	logrus.SetOutput(&buf)
	defer logrus.SetOutput(old)

	src := &source{instance: "em-a", host: srv.URL, username: "u", password: "p", insecure: true, trace: true}
	if _, err := src.Collect(context.Background()); err != nil {
		t.Fatalf("Collect: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Enterprise Plus") {
		t.Fatalf("trace did not log the licensing body; got: %s", out)
	}
	if strings.Contains(out, "sess-123") {
		t.Fatal("trace leaked the session token")
	}
}
