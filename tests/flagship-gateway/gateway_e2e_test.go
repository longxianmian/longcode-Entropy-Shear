package flagship_gateway_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"entropy-shear/internal/flagship/gateway"
	"entropy-shear/internal/flagship/provider"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func loadExample(t *testing.T, name string) []byte {
	t.Helper()
	p := filepath.Join(repoRoot(t), "examples", "flagship-gateway", name)
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read %s: %v", p, err)
	}
	return b
}

func newServer() *httptest.Server {
	return httptest.NewServer(gateway.Mux(provider.NewMockProvider()))
}

// respShape mirrors the gateway response just enough for assertions.
type respShape struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Role       string `json:"role"`
	Content    []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence *string `json:"stop_sequence"`
	Usage        struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Verdict      string `json:"verdict"`
	GatewayAudit *struct {
		GatewayID string `json:"gateway_id"`
		RequestID string `json:"request_id"`
		PreReasonerAudit struct {
			AuditID string `json:"audit_id"`
		} `json:"pre_reasoner_audit"`
		PostReasonerAudit *struct {
			AuditID string `json:"audit_id"`
		} `json:"post_reasoner_audit,omitempty"`
		ProviderName string `json:"provider_name"`
		Verdict      string `json:"verdict"`
	} `json:"gateway_audit,omitempty"`
}

func postJSON(t *testing.T, url string, body []byte) *http.Response {
	t.Helper()
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	return resp
}

func decodeResp(t *testing.T, body interface{ Read(p []byte) (int, error) }) respShape {
	t.Helper()
	var r respShape
	if err := json.NewDecoder(body).Decode(&r); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return r
}

func TestMessagesYesEndToEnd(t *testing.T) {
	srv := newServer()
	defer srv.Close()
	resp := postJSON(t, srv.URL+"/v1/messages", loadExample(t, "messages-yes-request.json"))
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
	if r.StopSequence != nil {
		t.Fatalf("stop_sequence must be null on YES; got %v", *r.StopSequence)
	}
	if len(r.Content) == 0 || r.Content[0].Type != "text" {
		t.Fatalf("content: %+v", r.Content)
	}
	if !strings.Contains(r.Content[0].Text, "[mock-candidate sha:") {
		t.Fatalf("YES content should embed mock candidate; got %q", r.Content[0].Text)
	}
	if r.GatewayAudit == nil {
		t.Fatal("missing gateway_audit")
	}
	if r.GatewayAudit.PreReasonerAudit.AuditID == "" {
		t.Fatal("missing pre audit id")
	}
	if r.GatewayAudit.PostReasonerAudit == nil || r.GatewayAudit.PostReasonerAudit.AuditID == "" {
		t.Fatal("YES path must run post-governance and surface its audit")
	}
	if r.GatewayAudit.PreReasonerAudit.AuditID == r.GatewayAudit.PostReasonerAudit.AuditID {
		t.Fatal("pre and post audit ids must differ on YES")
	}
	if r.GatewayAudit.ProviderName != "mock" {
		t.Fatalf("provider_name: got %s want mock", r.GatewayAudit.ProviderName)
	}
	if r.Usage.InputTokens <= 0 || r.Usage.OutputTokens <= 0 {
		t.Fatalf("usage: %+v want positive token counts", r.Usage)
	}
}

func TestMessagesHoldEndToEnd(t *testing.T) {
	srv := newServer()
	defer srv.Close()
	resp := postJSON(t, srv.URL+"/v1/messages", loadExample(t, "messages-hold-request.json"))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want 200", resp.StatusCode)
	}
	r := decodeResp(t, resp.Body)
	if r.Verdict != "HOLD" {
		t.Fatalf("verdict: got %s want HOLD (score must land in [T2, T1))", r.Verdict)
	}
	if r.StopReason != "end_turn" {
		t.Fatalf("stop_reason: got %s want end_turn (HOLD reuses end_turn per GD-3)", r.StopReason)
	}
	if !strings.Contains(strings.ToUpper(r.Content[0].Text), "HOLD") {
		t.Fatalf("HOLD content should mention HOLD; got %q", r.Content[0].Text)
	}
	if r.GatewayAudit == nil {
		t.Fatal("missing gateway_audit")
	}
	if r.GatewayAudit.PostReasonerAudit != nil {
		t.Fatal("pre-governance HOLD must short-circuit and not call provider")
	}
}

func TestMessagesNoEndToEnd(t *testing.T) {
	srv := newServer()
	defer srv.Close()
	resp := postJSON(t, srv.URL+"/v1/messages", loadExample(t, "messages-no-request.json"))
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want 200 (governance NO returns 200 per GD-10)", resp.StatusCode)
	}
	r := decodeResp(t, resp.Body)
	if r.Verdict != "NO" {
		t.Fatalf("verdict: got %s want NO", r.Verdict)
	}
	if r.StopReason != "refusal" {
		t.Fatalf("stop_reason: got %s want refusal", r.StopReason)
	}
	if !strings.Contains(r.Content[0].Text, "FLAGSHIP_PERMISSION_DENIED") {
		t.Fatalf("NO content should embed the reason code; got %q", r.Content[0].Text)
	}
	if r.GatewayAudit == nil || r.GatewayAudit.PostReasonerAudit != nil {
		t.Fatalf("pre-governance NO must short-circuit; got post audit %+v", r.GatewayAudit)
	}
}

func TestModelsListShape(t *testing.T) {
	srv := newServer()
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/v1/models")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d", resp.StatusCode)
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
	if len(body.Data) == 0 {
		t.Fatal("models list is empty")
	}
	if body.Data[0].ID != "flagship-coder-mock-1" {
		t.Fatalf("id: got %s want flagship-coder-mock-1", body.Data[0].ID)
	}
	if body.Data[0].Type != "model" {
		t.Fatalf("type: got %s want model", body.Data[0].Type)
	}
}

func TestHealthReachable(t *testing.T) {
	srv := newServer()
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["module"] != "flagship-coder-gateway" {
		t.Fatalf("module: got %s", body["module"])
	}
}

func TestGatewayAuditPreContainsRequestID(t *testing.T) {
	srv := newServer()
	defer srv.Close()
	resp := postJSON(t, srv.URL+"/v1/messages", loadExample(t, "messages-yes-request.json"))
	defer resp.Body.Close()
	r := decodeResp(t, resp.Body)
	if r.GatewayAudit == nil {
		t.Fatal("missing gateway_audit")
	}
	if !strings.HasPrefix(r.GatewayAudit.RequestID, "gw-req-") {
		t.Fatalf("gateway request_id should be gw-req-...; got %s", r.GatewayAudit.RequestID)
	}
	if r.ID == "" || !strings.HasPrefix(r.ID, "msg_") {
		t.Fatalf("response id should be msg_*; got %s", r.ID)
	}
}
