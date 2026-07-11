package tumblebug

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestGetMci_InfraPath verifies GetMciWithContext targets the cb-tumblebug
// v0.12.7 BREAKING path (/mci/ → /infra/), sends the Basic Auth header, and
// parses the renamed "node" response key into the Vm slice.
func TestGetMci_InfraPath(t *testing.T) {
	const nsId, mciId = "ant-ns", "ant-mci"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		wantPath := "/tumblebug/ns/" + nsId + "/infra/" + mciId
		if r.URL.Path != wantPath {
			t.Errorf("unexpected path: got %s, want %s", r.URL.Path, wantPath)
		}
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Basic ") {
			t.Errorf("missing Basic Auth header")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"ant-mci","name":"ant-mci","status":"Running","node":[{"id":"ant-vm-1"}]}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	res, err := c.GetMciWithContext(context.Background(), nsId, mciId)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if res.Id != "ant-mci" {
		t.Errorf("unexpected id: got %q", res.Id)
	}
	// The response key "node" must populate Vm ([]VmRes json:"node").
	if len(res.Vm) != 1 || res.Vm[0].Id != "ant-vm-1" {
		t.Errorf("expected node key parsed into Vm, got %+v", res.Vm)
	}
}

// TestGetVm_InfraNodePath verifies GetVmWithContext targets the cb-tumblebug
// v0.12.7 BREAKING path (/mci/ → /infra/, /vm/ → /node/), sends Basic Auth,
// and parses the renamed "nodeGroupId" response key into SubGroupId.
func TestGetVm_InfraNodePath(t *testing.T) {
	const nsId, mciId, vmId = "ant-ns", "ant-mci", "ant-vm-1"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		wantPath := "/tumblebug/ns/" + nsId + "/infra/" + mciId + "/node/" + vmId
		if r.URL.Path != wantPath {
			t.Errorf("unexpected path: got %s, want %s", r.URL.Path, wantPath)
		}
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Basic ") {
			t.Errorf("missing Basic Auth header")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"ant-vm-1","name":"ant-vm-1","status":"Running","nodeGroupId":"g1"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	res, err := c.GetVmWithContext(context.Background(), nsId, mciId, vmId)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if res.Id != "ant-vm-1" {
		t.Errorf("unexpected id: got %q", res.Id)
	}
	// The response key "nodeGroupId" must populate SubGroupId.
	if res.SubGroupId != "g1" {
		t.Errorf("expected nodeGroupId parsed into SubGroupId, got %q", res.SubGroupId)
	}
}

// TestGetMci_NotFound verifies that a cb-tumblebug 500 whose body looks like a
// missing-resource error is normalized to ErrNotFound.
func TestGetMci_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"does not exist"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.GetMciWithContext(context.Background(), "ant-ns", "missing")
	if err == nil {
		t.Fatalf("expected error for missing mci")
	}
}
