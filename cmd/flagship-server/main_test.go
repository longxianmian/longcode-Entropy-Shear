package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestServer wires the real reasonHandler/healthHandler into an httptest
// server so the same code path the binary serves is exercised in tests.
func newTestServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/flagship/reason", reasonHandler)
	mux.HandleFunc("/health", healthHandler)
	return httptest.NewServer(mux)
}

func decodeError(t *testing.T, body io.Reader) map[string]string {
	t.Helper()
	var m map[string]string
	if err := json.NewDecoder(body).Decode(&m); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	return m
}

func TestReasonHandlerRejectsGET(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/flagship/reason")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("status: got %d want 405", resp.StatusCode)
	}
	body := decodeError(t, resp.Body)
	if body["error"] != "method_not_allowed" {
		t.Fatalf("error code: got %q want method_not_allowed", body["error"])
	}
}

func TestReasonHandlerBadJSON(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	resp, err := http.Post(srv.URL+"/flagship/reason", "application/json", strings.NewReader("{not json"))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want 400", resp.StatusCode)
	}
	body := decodeError(t, resp.Body)
	if body["error"] != "bad_json" {
		t.Fatalf("error code: got %q want bad_json", body["error"])
	}
	if body["message"] == "" {
		t.Fatalf("expected non-empty error message; got %+v", body)
	}
}

func TestReasonHandlerDisallowsUnknownFields(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	// "request_id" is valid, "stowaway_field" is not — DisallowUnknownFields
	// must reject the whole request.
	payload := []byte(`{"request_id":"req-x","stowaway_field":"oops"}`)
	resp, err := http.Post(srv.URL+"/flagship/reason", "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want 400", resp.StatusCode)
	}
	body := decodeError(t, resp.Body)
	if body["error"] != "bad_json" {
		t.Fatalf("error code: got %q want bad_json", body["error"])
	}
	if !strings.Contains(strings.ToLower(body["message"]), "unknown field") {
		t.Fatalf("expected message to mention unknown field; got %q", body["message"])
	}
}

func TestReasonHandlerHappyMinimal(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	payload := []byte(`{"request_id":"req-min-001"}`)
	resp, err := http.Post(srv.URL+"/flagship/reason", "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("content-type: got %q want application/json...", ct)
	}
	var out map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	// A minimal request has all-zero element states except Constraint=1.0
	// (no constraints supplied), so the score is below T2 → NO. The handler
	// just needs to round-trip the JSON; verdict semantics are exercised in
	// tests/flagship.
	if out["verdict"] == nil {
		t.Fatalf("response missing verdict; got %+v", out)
	}
}

func TestHealthHandlerGET(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want 200", resp.StatusCode)
	}
	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "ok" || body["module"] != "flagship" {
		t.Fatalf("unexpected health body: %+v", body)
	}
}

func TestHealthHandlerRejectsPOST(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	resp, err := http.Post(srv.URL+"/health", "application/json", strings.NewReader("{}"))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("status: got %d want 405", resp.StatusCode)
	}
	body := decodeError(t, resp.Body)
	if body["error"] != "method_not_allowed" {
		t.Fatalf("error code: got %q want method_not_allowed", body["error"])
	}
}
