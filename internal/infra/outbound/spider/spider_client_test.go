package spider

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// newTestClient constructs a SpiderClient against the given httptest server.
// Tests in the same package can set unexported fields directly.
//
// withUrl uses only the domain field, so we set domain to the full httptest
// URL (scheme included). The host/port fields mirror the parsed URL for
// completeness and Endpoint() reporting.
func newTestClient(t *testing.T, srv *httptest.Server) *SpiderClient {
	t.Helper()
	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse httptest URL: %v", err)
	}
	creds := base64.StdEncoding.EncodeToString([]byte("default:default"))
	return &SpiderClient{
		client:     srv.Client(),
		host:       u.Hostname(),
		port:       u.Port(),
		username:   "default",
		password:   "default",
		domain:     srv.URL,
		authHeader: "Basic " + creds,
	}
}

// TestReadyzWithContext_OK verifies cb-spider /readyz happy path.
func TestReadyzWithContext_OK(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.URL.Path != "/spider/readyz" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message":"CB-Spider is ready"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if err := c.ReadyzWithContext(context.Background()); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if !called {
		t.Fatalf("handler not called")
	}
}

// TestAuthCheck_OK verifies the auth-enforced GET /spider/cloudos succeeds
// with a valid Basic Auth header.
func TestAuthCheck_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/spider/cloudos" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Basic ") {
			t.Errorf("missing Basic Auth header")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`["aws","azure","gcp"]`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if err := c.AuthCheckWithContext(context.Background()); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

// TestAuthCheck_Unauthorized verifies 401 surfaces as ErrUnauthorized so that
// callers can distinguish authentication failure from other errors and emit a
// targeted operator message (per STANDARD-READYZ §6.3).
func TestAuthCheck_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"Unauthorized"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.AuthCheckWithContext(context.Background())
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

// TestAuthCheck_InternalServerError verifies 500 maps to ErrInternalServerError.
func TestAuthCheck_InternalServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.AuthCheckWithContext(context.Background())
	if !errors.Is(err, ErrInternalServerError) {
		t.Fatalf("expected ErrInternalServerError, got %v", err)
	}
}

// TestAuthCheck_NetworkFailure verifies a closed server surfaces a non-nil
// error (Reachable=false case in dependency_check.go).
func TestAuthCheck_NetworkFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	srv.Close() // close immediately so requests fail

	c := newTestClient(t, srv)
	err := c.AuthCheckWithContext(context.Background())
	if err == nil {
		t.Fatalf("expected non-nil error on closed server")
	}
	if errors.Is(err, ErrUnauthorized) {
		t.Fatalf("network failure should not surface as ErrUnauthorized: %v", err)
	}
}
