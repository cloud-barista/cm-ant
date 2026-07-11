package spider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestGetPriceInfo_GET verifies GetPriceInfoWithContext uses the cb-spider
// v0.12.3 BREAKING contract: GET /spider/priceinfo/vm/{region} with
// ConnectionName as a query parameter and a Basic Auth header (v0.12.6+ forced
// auth), replacing the old POST + FilterList body.
func TestGetPriceInfo_GET(t *testing.T) {
	const region, conn = "ap-northeast-2", "aws-conn"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET (POST→GET transition), got %s", r.Method)
		}
		wantPath := "/spider/priceinfo/vm/" + region
		if r.URL.Path != wantPath {
			t.Errorf("unexpected path: got %s, want %s", r.URL.Path, wantPath)
		}
		if got := r.URL.Query().Get("ConnectionName"); got != conn {
			t.Errorf("expected ConnectionName query %q, got %q", conn, got)
		}
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Basic ") {
			t.Errorf("missing Basic Auth header")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"meta":{},"cloudPriceList":[{}]}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	res, err := c.GetPriceInfoWithContext(context.Background(), "vm", region, PriceInfoReq{ConnectionName: conn})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if len(res.CloudPriceList) != 1 {
		t.Errorf("expected cloudPriceList parsed, got %+v", res.CloudPriceList)
	}
}

// TestGetCostWithResource_AnycallAuthHeader verifies GetCostWithResourceWithContext
// POSTs to /spider/anycall with the Basic Auth header (cb-spider v0.12.6+ forced
// auth, BAR-685) and unwraps the anycall OKeyValueList result payload.
func TestGetCostWithResource_AnycallAuthHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/spider/anycall") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Basic ") {
			t.Errorf("missing Basic Auth header on anycall")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"FID":"getcostwithresource","IKeyValueList":[{"Key":"in","Value":"x"}],"OKeyValueList":[{"Key":"result","Value":"{}"}]}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	res, err := c.GetCostWithResourceWithContext(context.Background(), AnycallReq{ConnectionName: "aws-conn"})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if res == nil {
		t.Fatalf("expected non-nil cost result")
	}
}

// TestGetCostWithResource_EmptyResult verifies the empty-result sentinel is
// returned when the anycall response carries no output values.
func TestGetCostWithResource_EmptyResult(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"FID":"getcostwithresource","IKeyValueList":[],"OKeyValueList":null}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetCostWithResourceWithContext(context.Background(), AnycallReq{ConnectionName: "aws-conn"})
	if err == nil {
		t.Fatalf("expected ErrSpiderCostResultEmpty, got nil")
	}
}
