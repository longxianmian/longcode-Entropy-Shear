package flagship_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"entropy-shear/internal/flagship/reasoner"
	"entropy-shear/internal/flagship/rules"
)

// repoRoot walks up from this test file to the repository root so that the
// fixture-loading tests work regardless of where `go test` is invoked from.
func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// .../tests/flagship/reasoner_e2e_test.go → up two levels → repo root.
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func loadFixture(t *testing.T, name string) reasoner.Input {
	t.Helper()
	path := filepath.Join(repoRoot(t), "examples", "flagship", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}
	var in reasoner.Input
	if err := json.Unmarshal(data, &in); err != nil {
		t.Fatalf("decode fixture %s: %v", path, err)
	}
	return in
}

func TestReasonExampleYes(t *testing.T) {
	out := reasoner.Reason(loadFixture(t, "reason-yes-request.json"))
	if out.Verdict != reasoner.VerdictYes {
		t.Fatalf("expected YES, got %v (score=%v)", out.Verdict, out.Score)
	}
	if out.PermitToken == nil {
		t.Fatal("YES verdict must produce a permit token")
	}
	if out.RejectInstruction != nil {
		t.Fatal("YES verdict must not produce a reject instruction")
	}
	if len(out.AlignmentTasks) != 0 {
		t.Fatalf("YES verdict must not produce alignment tasks, got %d", len(out.AlignmentTasks))
	}
	if out.PermitToken.AuditID != out.AuditRecord.AuditID {
		t.Fatalf("permit token audit_id mismatch: %s vs %s", out.PermitToken.AuditID, out.AuditRecord.AuditID)
	}
	if out.AuditRecord.PermitTokenID != out.PermitToken.ID {
		t.Fatalf("audit record must reference permit token id")
	}
	if len(out.PermitToken.Scope) == 0 {
		t.Fatal("permit token scope should include action ids when actions exist")
	}
}

func TestReasonExampleHold(t *testing.T) {
	out := reasoner.Reason(loadFixture(t, "reason-hold-request.json"))
	if out.Verdict != reasoner.VerdictHold {
		t.Fatalf("expected HOLD, got %v (score=%v)", out.Verdict, out.Score)
	}
	if out.PermitToken != nil || out.RejectInstruction != nil {
		t.Fatal("HOLD verdict must not produce token/instruction")
	}
	if len(out.AlignmentTasks) == 0 {
		t.Fatal("HOLD verdict must produce at least one alignment task")
	}
	for _, tk := range out.AlignmentTasks {
		switch tk.Priority {
		case "low", "medium", "high":
		default:
			t.Fatalf("alignment task priority must be low/medium/high, got %q", tk.Priority)
		}
		if tk.ID == "" || tk.TargetElement == "" || tk.ReasonCode == "" || tk.Prompt == "" {
			t.Fatalf("alignment task missing required field: %+v", tk)
		}
	}
}

func TestReasonExampleNoHardConstraint(t *testing.T) {
	out := reasoner.Reason(loadFixture(t, "reason-no-request.json"))
	if out.Verdict != reasoner.VerdictNo {
		t.Fatalf("expected NO, got %v (score=%v)", out.Verdict, out.Score)
	}
	if out.RejectInstruction == nil {
		t.Fatal("NO verdict must produce reject instruction")
	}
	if out.PermitToken != nil {
		t.Fatal("NO verdict must not produce permit token")
	}
	if out.RejectInstruction.ReasonCode != rules.ReasonPermissionDenied {
		t.Fatalf("expected permission-denied reason code, got %q", out.RejectInstruction.ReasonCode)
	}
	if len(out.RejectInstruction.ConflictingItems) == 0 {
		t.Fatal("reject instruction must list conflicting items")
	}
	if len(out.RejectInstruction.RemediationSteps) == 0 {
		t.Fatal("reject instruction must include remediation steps")
	}
	if out.AuditRecord.RejectInstructionID != out.RejectInstruction.ID {
		t.Fatalf("audit record must reference reject instruction id")
	}
}

func TestReasonAuditDeterminism(t *testing.T) {
	in := loadFixture(t, "reason-yes-request.json")
	a := reasoner.Reason(in)
	b := reasoner.Reason(in)
	if a.AuditRecord.InputDigest != b.AuditRecord.InputDigest {
		t.Fatalf("input_digest must be deterministic for identical input")
	}
	if a.AuditRecord.MatrixDigest != b.AuditRecord.MatrixDigest {
		t.Fatalf("matrix_digest must be deterministic")
	}
	if a.AuditRecord.FiveElementDigest != b.AuditRecord.FiveElementDigest {
		t.Fatalf("five_element_digest must be deterministic")
	}
}

func TestReasonInvalidWeightsFallback(t *testing.T) {
	in := loadFixture(t, "reason-yes-request.json")
	in.Weights = map[string]float64{"Goal": 1.0} // missing keys
	out := reasoner.Reason(in)
	if out.Verdict != reasoner.VerdictYes {
		t.Fatalf("expected YES with default-weights fallback, got %v", out.Verdict)
	}
	found := false
	for _, line := range out.Trace {
		if bytes.Contains([]byte(line), []byte("weights override rejected")) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("trace should mention weights override fallback; got %v", out.Trace)
	}
}

// HTTP boundary: re-implement the handler logic in the test to avoid spinning
// up the binary; we hit the same Reason() path the cmd/flagship-server does.
func TestHTTPRoundTrip(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/flagship/reason", func(w http.ResponseWriter, r *http.Request) {
		var in reasoner.Input
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		out := reasoner.Reason(in)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	body, err := os.ReadFile(filepath.Join(repoRoot(t), "examples", "flagship", "reason-yes-request.json"))
	if err != nil {
		t.Fatalf("read example: %v", err)
	}
	resp, err := http.Post(srv.URL+"/flagship/reason", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want 200", resp.StatusCode)
	}
	var out reasoner.Output
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if out.Verdict != reasoner.VerdictYes {
		t.Fatalf("verdict: got %v want YES", out.Verdict)
	}
	if out.AuditRecord.AuditID == "" {
		t.Fatal("audit_id missing in response")
	}
}
