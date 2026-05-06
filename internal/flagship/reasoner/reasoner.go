package reasoner

import (
	"time"

	"entropy-shear/internal/flagship/hold"
	"entropy-shear/internal/flagship/mapper"
	"entropy-shear/internal/flagship/output"
	"entropy-shear/internal/flagship/rules"
	"entropy-shear/internal/flagship/state"
)

// Reason runs the deterministic three-state reasoner over a single input.
// It performs (in order): atomic-rule checks, multi-source mapping into the
// five-element state vector, weight resolution with risk modulation,
// 5x5-matrix interference scoring, the YES/HOLD/NO FSM, and emission of the
// matching artifact (PermitToken / AlignmentTask list / RejectInstruction).
//
// Reason has no I/O side effects: no logs, no ledger writes, no LLM calls.
func Reason(in Input) Output {
	return reasonAt(in, time.Now())
}

// reasonAt is Reason with an injectable clock for deterministic tests.
func reasonAt(in Input, now time.Time) Output {
	requestID := in.RequestID
	if requestID == "" {
		requestID = "req-anon"
	}

	var trace []string

	// 1. Atomic rules (also detects hard conflicts that force NO).
	ruleRes := rules.Apply(in.Goal, in.Facts, in.Evidence, in.Constraints, in.Actions)

	// 2. Multi-source → five-element states.
	states := mapper.Map(in.Goal, in.Facts, in.Evidence, in.Constraints, in.Actions)

	// 3. Resolve and risk-modulate weights.
	weights, fellBack, reason := state.ResolveWeights(in.Weights)
	if fellBack {
		trace = append(trace, "weights override rejected: "+reason+"; falling back to defaults")
	}
	state.ApplyRiskBoost(weights, in.Risk)

	// 4. Compute score and FSM verdict.
	comp := state.Compute(states, weights, ruleRes.HardPenalty, ruleRes.HasHardConflict)

	// 5. Build audit record. Canonical JSON inputs are computed on the
	//    structured forms so digest is reproducible.
	inputBytes, _ := output.CanonicalJSON(in)
	statesBytes, _ := output.CanonicalJSON(struct {
		States  mapper.ElementStates `json:"element_states"`
		Weights map[string]float64   `json:"normalized_weights"`
	}{States: states, Weights: weights})
	matrixBytes, _ := output.CanonicalJSON(state.DefaultMatrix)
	audit := output.NewAuditRecord(requestID, inputBytes, statesBytes, matrixBytes,
		ruleRes.TriggeredRuleIDs, string(comp.Verdict), now)

	out := Output{
		RequestID:         requestID,
		Verdict:           comp.Verdict,
		Score:             comp.Score,
		ElementStates:     states,
		NormalizedWeights: weights,
		AuditRecord:       audit,
		Trace:             trace,
	}

	// 6. Verdict-specific artifact.
	switch comp.Verdict {
	case state.VerdictYes:
		scope := actionScope(in.Actions)
		token := output.NewPermitToken(requestID, audit.AuditID, "FLAGSHIP_REASONER_YES", scope, now, output.DefaultTokenValidity)
		out.PermitToken = &token
		out.AuditRecord.PermitTokenID = token.ID
	case state.VerdictHold:
		tasks := hold.Generate(requestID, states, ruleRes,
			in.Goal, in.Facts, in.Evidence, in.Constraints, in.Actions)
		out.AlignmentTasks = tasks
	case state.VerdictNo:
		reasonCode := ruleRes.HardReasonCode
		if reasonCode == "" {
			reasonCode = "FLAGSHIP_REASONER_NO_LOW_SCORE"
		}
		remediation := defaultRemediation(reasonCode, ruleRes.ConflictingItems)
		rej := output.NewRejectInstruction(requestID, audit.AuditID, reasonCode,
			ruleRes.ConflictingItems, remediation)
		out.RejectInstruction = &rej
		out.AuditRecord.RejectInstructionID = rej.ID
	}

	return out
}

func actionScope(acts []Action) []string {
	out := make([]string, 0, len(acts))
	for _, a := range acts {
		if a.ID != "" {
			out = append(out, a.ID)
		}
	}
	return out
}

func defaultRemediation(reasonCode string, conflicts []string) []string {
	switch reasonCode {
	case rules.ReasonPermissionDenied:
		return []string{
			"obtain or escalate the missing permission for the requested scope",
			"re-submit with an updated permission constraint marked satisfied=true",
		}
	case rules.ReasonExplicitForbid:
		return []string{
			"do not retry under the current scope",
			"narrow the goal so the forbidden action is no longer required",
		}
	case rules.ReasonGovernanceReject:
		return []string{
			"raise a governance review for the rejecting party",
			"supply governance approval and re-submit with constraint satisfied=true",
		}
	case rules.ReasonHardConstraintUnsatisfied:
		out := []string{"satisfy the hard constraints listed in conflicting_items"}
		for _, id := range conflicts {
			out = append(out, "address constraint id="+id)
		}
		return out
	default:
		return []string{
			"increase evidence quality or relax weights, then re-submit",
			"split the goal into smaller verifiable sub-goals",
		}
	}
}
