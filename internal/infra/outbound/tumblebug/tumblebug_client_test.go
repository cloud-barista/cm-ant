package tumblebug

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

// newTestClient constructs a TumblebugClient against the given httptest server.
// withUrl uses only the domain field, so we set domain to the full httptest
// URL (scheme included). The host/port fields mirror the parsed URL.
func newTestClient(t *testing.T, srv *httptest.Server) *TumblebugClient {
	t.Helper()
	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse httptest URL: %v", err)
	}
	creds := base64.StdEncoding.EncodeToString([]byte("default:default"))
	return &TumblebugClient{
		client:     srv.Client(),
		host:       u.Hostname(),
		port:       u.Port(),
		username:   "default",
		password:   "default",
		domain:     srv.URL,
		authHeader: "Basic " + creds,
	}
}

// TestReadyz_FullyReady covers the happy path: HTTP 200 + Ready=true + Initialized=true.
func TestReadyz_FullyReady(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/tumblebug/readyz") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ready":true,"initialized":true,"message":"CB-Tumblebug is ready and initialized"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if err := c.ReadyzWithContext(context.Background()); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

// TestReadyz_NotInitialized covers the cb-tumblebug pattern B partial-failure
// case: HTTP 200 but Initialized=false. Per STANDARD-READYZ §5/§6, this must
// surface as a distinct error so cm-ant can mark the dependency as not fully
// healthy even though the HTTP status was 200.
func TestReadyz_NotInitialized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ready":true,"initialized":false,"message":"CB-Tumblebug is ready but not initialized"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.ReadyzWithContext(context.Background())
	if !errors.Is(err, ErrNotInitialized) {
		t.Fatalf("expected ErrNotInitialized, got %v", err)
	}
}

// TestReadyz_NotReady covers HTTP 503 + Ready=false.
func TestReadyz_NotReady(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"ready":false,"initialized":false,"message":"CB-Tumblebug is NOT ready"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.ReadyzWithContext(context.Background())
	// HTTP 503 maps to a generic unexpected status code error (not 401/404/500
	// special-cased in requestWithContext). We accept any non-nil error here.
	if err == nil {
		t.Fatalf("expected non-nil error for 503")
	}
}

// TestReadyz_BadJSON covers malformed body.
func TestReadyz_BadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not-json`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	err := c.ReadyzWithContext(context.Background())
	if err == nil {
		t.Fatalf("expected parse error")
	}
}

// TestAuthCheck_OK verifies the auth-enforced GET /tumblebug/cloudInfo path.
func TestAuthCheck_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/tumblebug/cloudInfo") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Basic ") {
			t.Errorf("missing Basic Auth header")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"cspNames":["aws","azure","gcp"]}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	if err := c.AuthCheckWithContext(context.Background()); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

// TestAuthCheck_Unauthorized verifies 401 surfaces as ErrUnauthorized.
func TestAuthCheck_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
