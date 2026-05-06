package flagship_test

import (
	"strings"
	"testing"

	"entropy-shear/internal/flagship/mapper"
	"entropy-shear/internal/flagship/output"
	"entropy-shear/internal/flagship/reasoner"
	"entropy-shear/internal/flagship/state"
)

// P1-H2: NO with Score below T2 and no hard conflict must surface
// reason_code = FLAGSHIP_REASONER_NO_LOW_SCORE.
func TestReasonNoLowScoreReasonCode(t *testing.T) {
	in := reasoner.Input{
		RequestID: "req-low-001",
		// Empty goal/facts/evidence/actions; no constraints.
		// Element states resolve to: Goal=0, Fact=0, Evidence=0,
		// Constraint=1.0 (unimpeded), Action=0. With default weights and no
		// risk modulation the score is 0.20 (below T2=0.35), hard conflict
		// is false, so the reasoner must produce NO with the low-score
		// reason code rather than any hard-conflict reason code.
	}
	out := reasoner.Reason(in)
	if out.Verdict != reasoner.VerdictNo {
		t.Fatalf("verdict: got %v (score=%v) want NO", out.Verdict, out.Score)
	}
	if out.Score >= state.T2 {
		t.Fatalf("score %v should be below T2 %v for this input", out.Score, state.T2)
	}
	if out.RejectInstruction == nil {
		t.Fatal("NO verdict must produce a reject instruction")
	}
	if got, want := out.RejectInstruction.ReasonCode, "FLAGSHIP_REASONER_NO_LOW_SCORE"; got != want {
		t.Fatalf("reason_code: got %q want %q", got, want)
	}
	if len(out.RejectInstruction.RemediationSteps) == 0 {
		t.Fatal("low-score reject must include remediation steps")
	}
}

// P1-H3: state.Compute must return a NormalizedWeights map that does not
// share storage with the caller's input map. Mutating the caller's map after
// Compute returns must not leak into the Computation.
func TestComputeReturnsWeightsCopy(t *testing.T) {
	in := map[string]float64{"Goal": 0.25, "Fact": 0.20, "Evidence": 0.25, "Constraint": 0.20, "Action": 0.10}
	got := state.Compute(mapper.ElementStates{Goal: 1, Fact: 1, Evidence: 1, Constraint: 1, Action: 1}, in, 0, false)
	snapshot := map[string]float64{}
	for k, v := range got.NormalizedWeights {
		snapshot[k] = v
	}
	// Mutate the caller's input map.
	for k := range in {
		in[k] = 999.0
	}
	for k, v := range snapshot {
		if got.NormalizedWeights[k] != v {
			t.Fatalf("NormalizedWeights[%q] leaked caller mutation: got %v want %v", k, got.NormalizedWeights[k], v)
		}
	}
}

// P1-H3 companion: the public Output.NormalizedWeights returned by Reason
// must likewise be insulated from the reasoner's internal weights map.
func TestReasonOutputNormalizedWeightsIndependent(t *testing.T) {
	a := reasoner.Reason(reasoner.Input{RequestID: "req-w-1"})
	b := reasoner.Reason(reasoner.Input{RequestID: "req-w-2"})
	// Mutate the first output's weights map; the second output must not see
	// the change.
	for k := range a.NormalizedWeights {
		a.NormalizedWeights[k] = -1
	}
	for k, v := range b.NormalizedWeights {
		if v == -1 {
			t.Fatalf("Output.NormalizedWeights[%q] shared storage across calls", k)
		}
	}
}

// P1-H4: NewRejectInstruction.ID must encode the conflicting_items content,
// not just the count. Two rejects with identical request/audit/reason but
// different conflict contents (same length) must yield different ids.
func TestRejectInstructionIDIncludesConflictContent(t *testing.T) {
	a := output.NewRejectInstruction("req-x", "audit-y", "FLAGSHIP_PERMISSION_DENIED",
		[]string{"c-permission-001"}, []string{"step a"})
	b := output.NewRejectInstruction("req-x", "audit-y", "FLAGSHIP_PERMISSION_DENIED",
		[]string{"c-permission-002"}, []string{"step a"})
	if a.ID == b.ID {
		t.Fatalf("ids must differ when conflict content differs: a=%s b=%s", a.ID, b.ID)
	}
	// Order also matters as a content fingerprint — reordering the same items
	// should change the id (defensive).
	c := output.NewRejectInstruction("req-x", "audit-y", "FLAGSHIP_PERMISSION_DENIED",
		[]string{"c-1", "c-2"}, []string{"step a"})
	d := output.NewRejectInstruction("req-x", "audit-y", "FLAGSHIP_PERMISSION_DENIED",
		[]string{"c-2", "c-1"}, []string{"step a"})
	if c.ID == d.ID {
		t.Fatalf("ids must differ for reordered conflicts: c=%s d=%s", c.ID, d.ID)
	}
	// Identical inputs still produce identical ids (deterministic).
	e := output.NewRejectInstruction("req-x", "audit-y", "FLAGSHIP_PERMISSION_DENIED",
		[]string{"c-permission-001"}, []string{"step a"})
	if e.ID != a.ID {
		t.Fatalf("identical inputs must produce identical ids: a=%s e=%s", a.ID, e.ID)
	}
}

// P1-H5: when CanonicalJSON cannot encode a payload, the reasoner must
// surface the error in trace and use a fallback digest, never silently
// swallow the error.
func TestReasonCanonicalJSONErrorTrace(t *testing.T) {
	// json.Marshal cannot encode a chan; place one in Metadata so the input
	// digest path triggers an error. The HTTP boundary cannot transport a
	// chan, so this is a Go-call-only failure mode — the reasoner must
	// still produce a usable Output with a trace entry.
	in := reasoner.Input{
		RequestID: "req-trace-001",
		Metadata: map[string]interface{}{
			"unmarshalable": make(chan int),
		},
	}
	out := reasoner.Reason(in)
	if out.AuditRecord.AuditID == "" {
		t.Fatal("audit record must still be produced on canonical JSON error")
	}
	var found bool
	for _, line := range out.Trace {
		if strings.Contains(line, "canonical_json_error[input]") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("trace must contain canonical_json_error[input]; got %v", out.Trace)
	}
	// Sanity: the matrix and five_element kinds are still encodable, so
	// only the input kind should have triggered the fallback.
	for _, line := range out.Trace {
		if strings.Contains(line, "canonical_json_error[matrix]") {
			t.Fatalf("matrix encoding must still succeed; got trace %q", line)
		}
		if strings.Contains(line, "canonical_json_error[five_element]") {
			t.Fatalf("five_element encoding must still succeed; got trace %q", line)
		}
	}
}
