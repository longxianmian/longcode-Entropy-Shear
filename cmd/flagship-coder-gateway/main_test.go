package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"entropy-shear/internal/flagship/gateway"
	"entropy-shear/internal/flagship/provider"
)

func newTestServer() *httptest.Server {
	return httptest.NewServer(gateway.Mux(provider.NewMockProvider()))
}

func decodeError(t *testing.T, body io.Reader) map[string]string {
	t.Helper()
	var m map[string]string
	if err := json.NewDecoder(body).Decode(&m); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	return m
}

func TestMessagesHandlerRejectsGET(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/v1/messages")
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

func TestMessagesHandlerBadJSON(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	resp, err := http.Post(srv.URL+"/v1/messages", "application/json", strings.NewReader("{not json"))
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
}

func TestMessagesHandlerDisallowsUnknownFields(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	payload := []byte(`{"model":"flagship-coder-mock-1","max_tokens":256,"messages":[],"stowaway_field":"oops"}`)
	resp, err := http.Post(srv.URL+"/v1/messages", "application/json", bytes.NewReader(payload))
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

func TestMessagesHandlerIgnoresAnthropicHeaders(t *testing.T) {
	// GD-1 tail: P0 must allow these headers to be present and must not
	// reject the request because of them. We do not validate the API key.
	srv := newTestServer()
	defer srv.Close()
	payload := []byte(`{
		"model":"flagship-coder-mock-1",
		"max_tokens":256,
		"messages":[{"role":"user","content":[{"type":"text","text":"hi"}]}]
	}`)
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/v1/messages", bytes.NewReader(payload))
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("anthropic-beta", "tools-2024-04-04")
	req.Header.Set("authorization", "Bearer fake-not-validated")
	req.Header.Set("x-api-key", "sk-not-validated")
	req.Header.Set("content-type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want 200 (extra headers must not error)", resp.StatusCode)
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
	if body["status"] != "ok" || body["module"] != "flagship-coder-gateway" {
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
}

func TestModelsHandlerRejectsPOST(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	resp, err := http.Post(srv.URL+"/v1/models", "application/json", strings.NewReader("{}"))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("status: got %d want 405", resp.StatusCode)
	}
}

func TestCountTokensHandlerRejectsGET(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/v1/messages/count_tokens")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("status: got %d want 405", resp.StatusCode)
	}
}
