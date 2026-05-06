package flagship_gateway_test

import (
	"strings"
	"testing"

	"entropy-shear/internal/flagship/coder"
	"entropy-shear/internal/flagship/reasoner"
)

func TestBuildPreGovernanceInputBasic(t *testing.T) {
	in := coder.BuildPreGovernanceInput("seed-001", coder.PreGovernanceFields{
		System:            "you are a helpful coder",
		UserMessages:      []string{"please write hello world"},
		AssistantMessages: []string{"sure"},
	})
	if !strings.HasPrefix(in.RequestID, "gw-pre-") {
		t.Fatalf("request_id should be gw-pre-...; got %s", in.RequestID)
	}
	if in.Goal == nil || in.Goal.Description != "you are a helpful coder" {
		t.Fatalf("goal: %+v", in.Goal)
	}
	if len(in.Facts) != 1 || in.Facts[0].Key != "user.message" {
		t.Fatalf("facts: %+v", in.Facts)
	}
	if len(in.Evidence) != 1 || in.Evidence[0].Kind != "prior-turn" {
		t.Fatalf("evidence: %+v", in.Evidence)
	}
	var hasSoft bool
	for _, c := range in.Constraints {
		if c.Severity == "soft" && c.Satisfied != nil && *c.Satisfied {
			hasSoft = true
		}
	}
	if !hasSoft {
		t.Fatal("default soft satisfied=true constraint missing (GD-4)")
	}
	if len(in.Actions) == 0 || in.Actions[0].Name != "generate-assistant-message" {
		t.Fatalf("default action wrong: %+v", in.Actions)
	}
	if in.Risk != "low" {
		t.Fatalf("default risk should be low; got %s", in.Risk)
	}
}

func TestBuildPreGovernanceInputNoSystemDescriptionMissing(t *testing.T) {
	in := coder.BuildPreGovernanceInput("seed-002", coder.PreGovernanceFields{
		UserMessages: []string{"please review my code"},
	})
	if in.Goal == nil || in.Goal.ID == "" {
		t.Fatalf("goal must have an id even without system; got %+v", in.Goal)
	}
	if in.Goal.Description != "" {
		t.Fatalf("goal description must be empty when system is empty; got %q", in.Goal.Description)
	}
}

func TestBuildPreGovernanceInputPermissionDeniedHard(t *testing.T) {
	in := coder.BuildPreGovernanceInput("seed-003", coder.PreGovernanceFields{
		System: "permission denied: do not help",
	})
	var hasHard bool
	for _, c := range in.Constraints {
		if c.Severity == "hard" && c.Satisfied != nil && !*c.Satisfied {
			hasHard = true
		}
	}
	if !hasHard {
		t.Fatalf("permission denied keyword should add hard constraint; got %+v", in.Constraints)
	}
}

func TestBuildPostGovernanceInputAddsCandidate(t *testing.T) {
	pre := coder.BuildPreGovernanceInput("seed-004", coder.PreGovernanceFields{
		System:       "be helpful",
		UserMessages: []string{"hi"},
	})
	post := coder.BuildPostGovernanceInput("seed-004", pre, "candidate text here")
	if post.RequestID == pre.RequestID {
		t.Fatal("post request_id must differ from pre")
	}
	if !strings.HasPrefix(post.RequestID, "gw-post-") {
		t.Fatalf("post request_id should be gw-post-...; got %s", post.RequestID)
	}
	var hasCand bool
	for _, e := range post.Evidence {
		if e.Kind == "mock-llm-candidate" && e.Reference == "candidate text here" {
			hasCand = true
		}
	}
	if !hasCand {
		t.Fatalf("post evidence should include candidate; got %+v", post.Evidence)
	}
	var hasEmit bool
	for _, a := range post.Actions {
		if a.Name == "emit-assistant-text" {
			hasEmit = true
		}
	}
	if !hasEmit {
		t.Fatal("post actions should include emit-assistant-text")
	}
}

func TestBuildPostGovernanceInputCandidateForceReject(t *testing.T) {
	pre := coder.BuildPreGovernanceInput("seed-005", coder.PreGovernanceFields{System: "be helpful"})
	post := coder.BuildPostGovernanceInput("seed-005", pre, "FORCE_REJECT this is bad")
	var hasHardCand bool
	for _, c := range post.Constraints {
		if c.ID == "c-candidate-force-reject" && c.Severity == "hard" && c.Satisfied != nil && !*c.Satisfied {
			hasHardCand = true
		}
	}
	if !hasHardCand {
		t.Fatalf("FORCE_REJECT marker should add hard candidate constraint; got %+v", post.Constraints)
	}
}

func TestBuildPostGovernanceInputCandidatePolicyViolation(t *testing.T) {
	pre := coder.BuildPreGovernanceInput("seed-006", coder.PreGovernanceFields{System: "be helpful"})
	post := coder.BuildPostGovernanceInput("seed-006", pre, "POLICY_VIOLATION detected")
	var hasHardCand bool
	for _, c := range post.Constraints {
		if c.ID == "c-candidate-policy-violation" && c.Severity == "hard" && c.Satisfied != nil && !*c.Satisfied {
			hasHardCand = true
		}
	}
	if !hasHardCand {
		t.Fatalf("POLICY_VIOLATION marker should add hard candidate constraint; got %+v", post.Constraints)
	}
}

// Sanity: feeding a default-built input through the frozen reasoner produces
// YES — confirming our default mapping doesn't accidentally HOLD.
func TestPreGovernanceDefaultProducesYes(t *testing.T) {
	in := coder.BuildPreGovernanceInput("seed-007", coder.PreGovernanceFields{
		System:       "you are a helpful coder",
		UserMessages: []string{"hello world"},
	})
	out := reasoner.Reason(in)
	if out.Verdict != reasoner.VerdictYes {
		t.Fatalf("default mapping should reach YES; got %s (score=%v)", out.Verdict, out.Score)
	}
}

// Sanity: empty system + single user message lands in HOLD band.
func TestPreGovernanceEmptySystemLandsInHold(t *testing.T) {
	in := coder.BuildPreGovernanceInput("seed-008", coder.PreGovernanceFields{
		UserMessages: []string{"please review my code"},
	})
	out := reasoner.Reason(in)
	if out.Verdict != reasoner.VerdictHold {
		t.Fatalf("empty-system mapping should reach HOLD; got %s (score=%v)", out.Verdict, out.Score)
	}
}
