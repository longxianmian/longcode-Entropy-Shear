package flagship_gateway_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"entropy-shear/internal/flagship/gateway"
	"entropy-shear/internal/flagship/provider"
)

// claudeHeaders are the headers Claude Code-class clients commonly attach.
// Per GD-1 (P1-E5) the gateway must accept and ignore them all without
// producing 4xx, including the api-key headers — P1 does not validate the
// key. Real key handling is a Gateway P2 decision (GH-2).
var claudeHeaders = map[string]string{
	"anthropic-version": "2023-06-01",
	"anthropic-beta":    "tools-2024-04-04,prompt-caching-2024-07-31",
	"authorization":     "Bearer fake-not-validated",
	"x-api-key":         "sk-fake-not-validated",
	"content-type":      "application/json",
	"user-agent":        "claude-cli/0.x (entropy-shear-flagship-coder-gateway-p1-e2e)",
	"x-app":             "claude-code",
}

func newClaudeServer() *httptest.Server {
	return httptest.NewServer(gateway.Mux(provider.NewMockProvider()))
}

// doClaudeStyle sends body to url with the full claudeHeaders set, mimicking
// what a Claude Code-class client would attach.
func doClaudeStyle(t *testing.T, method, url string, body []byte) *http.Response {
	t.Helper()
	var rdr *bytes.Reader
	if body != nil {
		rdr = bytes.NewReader(body)
	}
	var req *http.Request
	var err error
	if rdr != nil {
		req, err = http.NewRequest(method, url, rdr)
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	for k, v := range claudeHeaders {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	return resp
}

// E2: GET /v1/models returns Anthropic-compatible data array containing the
// flagship-coder-mock-1 entry, even when called with a full Claude header
// set.
func TestClaudeE2EModelsListShape(t *testing.T) {
	srv := newClaudeServer()
	defer srv.Close()
	resp := doClaudeStyle(t, http.MethodGet, srv.URL+"/v1/models", nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want 200", resp.StatusCode)
	}
	var body struct {
		Data []struct {
			ID          string `json:"id"`
			Type        string `json:"type"`
			DisplayName string `json:"display_name"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	var hasMock bool
	for _, m := range body.Data {
		if m.ID == "flagship-coder-mock-1" && m.Type == "model" {
			hasMock = true
		}
	}
	if !hasMock {
		t.Fatalf("data must contain flagship-coder-mock-1 model: %+v", body.Data)
	}
}

// E3: POST /v1/messages/count_tokens accepts the full Claude header set and
// returns the input_tokens shape.
func TestClaudeE2ECountTokensWithHeaders(t *testing.T) {
	srv := newClaudeServer()
	defer srv.Close()
	body := loadExample(t, "messages-claude-code-style-request.json")
	// count_tokens does not accept max_tokens; strip it by re-encoding.
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	delete(raw, "max_tokens")
	stripped, err := json.Marshal(raw)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	resp := doClaudeStyle(t, http.MethodPost, srv.URL+"/v1/messages/count_tokens", stripped)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want 200", resp.StatusCode)
	}
	var out struct {
		InputTokens int `json:"input_tokens"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.InputTokens <= 0 {
		t.Fatalf("input_tokens must be > 0 for non-trivial Claude-style request; got %d", out.InputTokens)
	}
}

// E4 part 1: POST /v1/messages with a Claude Code-style request (system,
// multi-turn messages, metadata) reaches YES.
func TestClaudeE2EMessagesYes(t *testing.T) {
	srv := newClaudeServer()
	defer srv.Close()
	resp := doClaudeStyle(t, http.MethodPost, srv.URL+"/v1/messages",
		loadExample(t, "messages-claude-code-style-request.json"))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want 200", resp.StatusCode)
	}
	r := decodeResp(t, resp.Body)
	if r.Type != "message" || r.Role != "assistant" {
		t.Fatalf("type=%s role=%s want message/assistant", r.Type, r.Role)
	}
	if r.Verdict != "YES" {
		t.Fatalf("verdict: got %s want YES", r.Verdict)
	}
	if r.StopReason != "end_turn" {
		t.Fatalf("stop_reason: got %s want end_turn", r.StopReason)
	}
	if r.GatewayAudit == nil {
		t.Fatal("missing gateway_audit")
	}
	if r.GatewayAudit.PostReasonerAudit == nil {
		t.Fatal("YES path must include post-governance audit")
	}
	// Claude-style request is an Anthropic model id; the gateway must echo
	// it back unchanged in the response.
	if r.Model != "claude-sonnet-4-5" {
		t.Fatalf("model echo: got %s want claude-sonnet-4-5", r.Model)
	}
}

// E4 part 2 + E6: HOLD path remains 200 with body.verdict=HOLD and
// stop_reason=end_turn even when invoked with full Claude headers.
func TestClaudeE2EMessagesHold(t *testing.T) {
	srv := newClaudeServer()
	defer srv.Close()
	resp := doClaudeStyle(t, http.MethodPost, srv.URL+"/v1/messages",
		loadExample(t, "messages-hold-request.json"))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("HOLD must return 200 (GD-10); got %d", resp.StatusCode)
	}
	r := decodeResp(t, resp.Body)
	if r.Verdict != "HOLD" {
		t.Fatalf("verdict: got %s want HOLD", r.Verdict)
	}
	if r.StopReason != "end_turn" {
		t.Fatalf("stop_reason: got %s want end_turn (GD-3)", r.StopReason)
	}
	if r.GatewayAudit == nil || r.GatewayAudit.PostReasonerAudit != nil {
		t.Fatal("pre-governance HOLD must short-circuit; post audit must be nil")
	}
}

// E4 part 3 + E6: NO path remains 200 with body.verdict=NO and
// stop_reason=refusal even when invoked with full Claude headers.
func TestClaudeE2EMessagesNo(t *testing.T) {
	srv := newClaudeServer()
	defer srv.Close()
	resp := doClaudeStyle(t, http.MethodPost, srv.URL+"/v1/messages",
		loadExample(t, "messages-no-request.json"))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("NO must return 200 (GD-10); got %d", resp.StatusCode)
	}
	r := decodeResp(t, resp.Body)
	if r.Verdict != "NO" {
		t.Fatalf("verdict: got %s want NO", r.Verdict)
	}
	if r.StopReason != "refusal" {
		t.Fatalf("stop_reason: got %s want refusal (GD-3)", r.StopReason)
	}
	if r.GatewayAudit == nil || r.GatewayAudit.PostReasonerAudit != nil {
		t.Fatal("pre-governance NO must short-circuit; post audit must be nil")
	}
}

// E5: Each individual Claude header alone must not produce a 4xx on
// /v1/messages. Catches accidental future header validation that would
// break Claude Code compatibility.
func TestClaudeE2EEachHeaderAccepted(t *testing.T) {
	srv := newClaudeServer()
	defer srv.Close()
	body := loadExample(t, "messages-claude-code-style-request.json")
	for name, value := range claudeHeaders {
		t.Run(name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, srv.URL+"/v1/messages", bytes.NewReader(body))
			if err != nil {
				t.Fatalf("new request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set(name, value)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("do: %v", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("header %q triggered status %d; P1-E5 says headers must be accepted and ignored", name, resp.StatusCode)
			}
		})
	}
}

// E5 part 2: bogus / unknown headers (mimicking future Anthropic versions
// or experimental beta flags) must also be silently ignored.
func TestClaudeE2EUnknownFutureHeadersAccepted(t *testing.T) {
	srv := newClaudeServer()
	defer srv.Close()
	body := loadExample(t, "messages-claude-code-style-request.json")
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/v1/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", "future-version-9999-99-99")
	req.Header.Set("anthropic-beta", "future-beta-flag,another-experimental-flag")
	req.Header.Set("x-anthropic-experimental", "yes")
	req.Header.Set("x-claude-code-session-id", "session_abc123")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unknown headers must not produce 4xx; got %d", resp.StatusCode)
	}
}

// E6 standalone matrix: HOLD / NO / YES all share HTTP 200, the verdict is
// expressed only in body.verdict, stop_reason follows GD-3.
func TestClaudeE2EVerdictMatrixAllReturn200(t *testing.T) {
	srv := newClaudeServer()
	defer srv.Close()
	cases := []struct {
		name       string
		example    string
		wantVerd   string
		wantStop   string
	}{
		{"YES Claude-style", "messages-claude-code-style-request.json", "YES", "end_turn"},
		{"HOLD existing P0", "messages-hold-request.json", "HOLD", "end_turn"},
		{"NO existing P0", "messages-no-request.json", "NO", "refusal"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := doClaudeStyle(t, http.MethodPost, srv.URL+"/v1/messages", loadExample(t, tc.example))
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status: got %d want 200 (GD-10 HOLD/NO must be 200)", resp.StatusCode)
			}
			r := decodeResp(t, resp.Body)
			if r.Verdict != tc.wantVerd {
				t.Fatalf("verdict: got %s want %s", r.Verdict, tc.wantVerd)
			}
			if r.StopReason != tc.wantStop {
				t.Fatalf("stop_reason: got %s want %s (GD-3)", r.StopReason, tc.wantStop)
			}
			if r.Type != "message" || r.Role != "assistant" {
				t.Fatalf("response shape drift: type=%s role=%s", r.Type, r.Role)
			}
		})
	}
}

// E2 + E5: GET /v1/models also accepts the full Claude header set without
// erroring (some clients inspect /v1/models with the same headers as
// /v1/messages).
func TestClaudeE2EModelsAcceptsHeaders(t *testing.T) {
	srv := newClaudeServer()
	defer srv.Close()
	resp := doClaudeStyle(t, http.MethodGet, srv.URL+"/v1/models", nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("content-type: got %q", ct)
	}
}

// E2 sanity: even an empty header set still produces 200; this guards
// against an accidental "require Claude headers" regression.
func TestClaudeE2EModelsWithoutHeaders(t *testing.T) {
	srv := newClaudeServer()
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/v1/models")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status without headers: got %d want 200", resp.StatusCode)
	}
}
