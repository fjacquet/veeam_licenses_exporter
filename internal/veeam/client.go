package veeam

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

// emClient is a hand-rolled Veeam Backup Enterprise Manager REST client. Session
// auth is stateless per cycle: login returns the X-RestSvcSessionId token, which
// is sent on the licensing GET, and the session is deleted on logout.
type emClient struct {
	rc    *resty.Client
	token string
}

func newEMClient(host string, insecure bool) *emClient {
	rc := resty.New().
		SetBaseURL(host).
		SetTimeout(30*time.Second).
		// MinVersion pinned to TLS 1.2: Enterprise Manager is commonly IIS-hosted on
		// Windows Server and may not yet offer TLS 1.3.
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: insecure, MinVersion: tls.VersionTLS12}).
		SetHeader("Accept", "application/json").
		// Retry transport/5xx only; 4xx (auth) is never retried.
		SetRetryCount(2).
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return err != nil || r.StatusCode() >= 500
		})
	return &emClient{rc: rc}
}

func (c *emClient) login(ctx context.Context, username, password string) error {
	resp, err := c.rc.R().SetContext(ctx).SetBasicAuth(username, password).
		Post("/api/sessionMngr/?v=latest")
	if err != nil {
		return fmt.Errorf("em login: %w", err)
	}
	if resp.StatusCode() >= 400 {
		return fmt.Errorf("em login: status %d", resp.StatusCode())
	}
	c.token = resp.Header().Get("X-RestSvcSessionId")
	if c.token == "" {
		return fmt.Errorf("em login: no X-RestSvcSessionId in response")
	}
	return nil
}

func (c *emClient) licensing(ctx context.Context) ([]byte, error) {
	resp, err := c.rc.R().SetContext(ctx).SetHeader("X-RestSvcSessionId", c.token).
		Get("/api/licensing")
	if err != nil {
		return nil, fmt.Errorf("em licensing: %w", err)
	}
	if resp.StatusCode() >= 400 {
		return nil, fmt.Errorf("em licensing: status %d", resp.StatusCode())
	}
	return resp.Body(), nil
}

// logout is best-effort on a fresh bounded context so it runs even if the cycle
// context was cancelled. The session id form is EM-version dependent; a failure is
// the caller's to log (potential session leak), never fatal.
func (c *emClient) logout() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := c.rc.R().SetContext(ctx).SetHeader("X-RestSvcSessionId", c.token).
		Delete("/api/logonSessions/current")
	return err
}
