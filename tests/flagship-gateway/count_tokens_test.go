package flagship_gateway_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestCountTokensEmptyMessages(t *testing.T) {
	srv := newServer()
	defer srv.Close()
	body := []byte(`{"model":"flagship-coder-mock-1","messages":[]}`)
	resp, err := http.Post(srv.URL+"/v1/messages/count_tokens", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
	var out struct {
		InputTokens int `json:"input_tokens"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.InputTokens < 0 {
		t.Fatalf("input_tokens must be >= 0; got %d", out.InputTokens)
	}
}

func TestCountTokensNonZeroForRealishMessage(t *testing.T) {
	srv := newServer()
	defer srv.Close()
	body := loadExample(t, "count-tokens-request.json")
	resp, err := http.Post(srv.URL+"/v1/messages/count_tokens", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
	var out struct {
		InputTokens int `json:"input_tokens"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.InputTokens <= 0 {
		t.Fatalf("input_tokens should be > 0 for non-empty input; got %d", out.InputTokens)
	}
}

func TestCountTokensLongInputScalesUp(t *testing.T) {
	srv := newServer()
	defer srv.Close()
	long := strings.Repeat("x", 4000)
	body := []byte(`{"model":"flagship-coder-mock-1","messages":[{"role":"user","content":[{"type":"text","text":"` + long + `"}]}]}`)
	resp, err := http.Post(srv.URL+"/v1/messages/count_tokens", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()
	var out struct {
		InputTokens int `json:"input_tokens"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// GD-6: ceil(JSON byte length / 4); a 4000-byte payload must exceed 500 tokens.
	if out.InputTokens < 500 {
		t.Fatalf("expected >=500 tokens for 4000-byte text; got %d", out.InputTokens)
	}
}

func TestCountTokensRejectsBadJSON(t *testing.T) {
	srv := newServer()
	defer srv.Close()
	resp, err := http.Post(srv.URL+"/v1/messages/count_tokens", "application/json", strings.NewReader("{not json"))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want 400", resp.StatusCode)
	}
}
