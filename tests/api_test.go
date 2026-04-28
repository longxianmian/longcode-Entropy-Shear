package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"entropy-shear/internal/api"
	"entropy-shear/internal/ledger"
)

func newServer(t *testing.T) *api.Server {
	t.Helper()
	dir := t.TempDir()
	l, err := ledger.New(filepath.Join(dir, "shear-chain.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	return &api.Server{Ledger: l, Version: "v1.0.0-test"}
}

func TestHealthEndpoint(t *testing.T) {
	srv := newServer(t)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["service"] != "entropy-shear" {
		t.Errorf("service=%v", body["service"])
	}
	if body["ok"] != true {
		t.Errorf("ok=%v", body["ok"])
	}
}

func TestShearEndpointHappyPath(t *testing.T) {
	srv := newServer(t)
	body := strings.NewReader(`{
	  "policy": {
	    "id": "policy-cityone-v1", "version": "1.0.0",
	    "rules": [{
	      "id": "rule-001", "priority": 1,
	      "condition": {"field": "user.level", "operator": "in", "value": ["member","vip"]},
	      "effect": "Yes", "route": "/x", "reason": "ok"
	    }],
	    "default_effect": "Hold", "default_reason": "fallthrough"
	  },
	  "facts": {"user": {"level": "member"}}
	}`)
	req := httptest.NewRequest(http.MethodPost, "/shear", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp["verdict"] != "Yes" {
		t.Errorf("verdict=%v", resp["verdict"])
	}
	if !strings.HasPrefix(resp["signature"].(string), "sha256:") {
		t.Errorf("signature=%v", resp["signature"])
	}
	if !strings.HasPrefix(resp["shear_id"].(string), "entropy-shear-") {
		t.Errorf("shear_id=%v", resp["shear_id"])
	}

	// Ledger verify should now pass with one record.
	verifyReq := httptest.NewRequest(http.MethodGet, "/ledger/verify", nil)
	verifyRR := httptest.NewRecorder()
	srv.Routes().ServeHTTP(verifyRR, verifyReq)
	if verifyRR.Code != http.StatusOK {
		t.Fatalf("verify status=%d", verifyRR.Code)
	}
	var v map[string]interface{}
	if err := json.NewDecoder(verifyRR.Body).Decode(&v); err != nil {
		t.Fatal(err)
	}
	if v["ok"] != true {
		t.Errorf("verify ok=%v body=%s", v["ok"], verifyRR.Body.String())
	}
	if v["total"].(float64) != 1 {
		t.Errorf("total=%v", v["total"])
	}
}

func TestShearInvalidJSONReturns400(t *testing.T) {
	srv := newServer(t)
	req := httptest.NewRequest(http.MethodPost, "/shear", bytes.NewReader([]byte(`{not json}`)))
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d want 400 body=%s", rr.Code, rr.Body.String())
	}
	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["error"] != "invalid_json" {
		t.Errorf("error=%v", body["error"])
	}
}

func TestShearUnsupportedOperatorReturns422(t *testing.T) {
	srv := newServer(t)
	body := strings.NewReader(`{
	  "policy": {
	    "id": "p", "version": "1.0.0",
	    "rules": [{
	      "id": "r", "priority": 1,
	      "condition": {"field": "x", "operator": "regex", "value": ".*"},
	      "effect": "Yes", "reason": "ok"
	    }],
	    "default_effect": "Hold", "default_reason": "x"
	  },
	  "facts": {}
	}`)
	req := httptest.NewRequest(http.MethodPost, "/shear", body)
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d want 422", rr.Code)
	}
	var b map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&b); err != nil {
		t.Fatal(err)
	}
	if b["error"] != "unsupported_operator" {
		t.Errorf("error=%v", b["error"])
	}
}

func TestShearMissingPolicyFieldReturns422(t *testing.T) {
	srv := newServer(t)
	body := strings.NewReader(`{"policy":{"version":"1"},"facts":{}}`)
	req := httptest.NewRequest(http.MethodPost, "/shear", body)
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d want 422", rr.Code)
	}
	var b map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&b); err != nil {
		t.Fatal(err)
	}
	if b["error"] != "policy_schema_violation" {
		t.Errorf("error=%v", b["error"])
	}
}

func TestLedgerGetRoundtrip(t *testing.T) {
	srv := newServer(t)
	body := strings.NewReader(`{
	  "policy": {"id":"p","version":"1.0.0","rules":[],"default_effect":"Hold","default_reason":"x"},
	  "facts": {}
	}`)
	req := httptest.NewRequest(http.MethodPost, "/shear", body)
	rr := httptest.NewRecorder()
	srv.Routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("post: %d %s", rr.Code, rr.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	id := resp["shear_id"].(string)

	getReq := httptest.NewRequest(http.MethodGet, "/ledger/"+id, nil)
	getRR := httptest.NewRecorder()
	srv.Routes().ServeHTTP(getRR, getReq)
	if getRR.Code != http.StatusOK {
		t.Fatalf("get status=%d body=%s", getRR.Code, getRR.Body.String())
	}
	var rec map[string]interface{}
	if err := json.NewDecoder(getRR.Body).Decode(&rec); err != nil {
		t.Fatal(err)
	}
	if rec["shear_id"] != id {
		t.Errorf("shear_id mismatch: %v vs %s", rec["shear_id"], id)
	}
	if rec["previous_shear_hash"] != "sha256:genesis" {
		t.Errorf("first record previous_shear_hash=%v", rec["previous_shear_hash"])
	}
}
