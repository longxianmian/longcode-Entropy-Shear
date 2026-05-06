package coder

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"entropy-shear/internal/flagship/reasoner"
)

// PreGovernanceFields is the lowered request shape that the gateway hands to
// the adapter. Keeping the gateway-side Anthropic types out of this signature
// avoids a coder→gateway import cycle.
type PreGovernanceFields struct {
	System            string
	UserMessages      []string
	AssistantMessages []string
	Metadata          map[string]interface{}
}

// BuildPreGovernanceInput maps the lowered request to a reasoner.Input per
// GD-4. The request_id is "gw-pre-" + a deterministic short hash of seedID
// so identical requests reproduce identical audit ids.
func BuildPreGovernanceInput(seedID string, fields PreGovernanceFields) reasoner.Input {
	goal := buildGoal(fields.System)
	facts := buildUserFacts(fields.UserMessages)
	evidence := buildPriorTurnEvidence(fields.AssistantMessages)
	constraints := requestConstraints(fields.System, fields.UserMessages)
	actions := defaultActions()
	return reasoner.Input{
		RequestID:   "gw-pre-" + shortHash(seedID),
		Goal:        goal,
		Facts:       facts,
		Evidence:    evidence,
		Constraints: constraints,
		Actions:     actions,
		Risk:        "low",
		Metadata:    fields.Metadata,
	}
}

// BuildPostGovernanceInput extends a pre-governance Input with the candidate
// produced by the LLM provider per GD-5. It re-uses the pre-governance Goal
// and Facts so the post check stays anchored on the same intent.
func BuildPostGovernanceInput(seedID string, pre reasoner.Input, candidate string) reasoner.Input {
	facts := append([]reasoner.Fact(nil), pre.Facts...)
	evidence := append([]reasoner.Evidence(nil), pre.Evidence...)
	evidence = append(evidence, reasoner.Evidence{
		ID:         "e-candidate",
		Kind:       "mock-llm-candidate",
		Reference:  candidate,
		Confidence: 0.5,
		Verified:   false,
	})
	actions := append([]reasoner.Action(nil), pre.Actions...)
	actions = append(actions, reasoner.Action{
		ID:         "a-emit-assistant",
		Name:       "emit-assistant-text",
		Reversible: true,
		Cost:       0.3,
	})
	constraints := append([]reasoner.Constraint(nil), pre.Constraints...)
	constraints = append(constraints, candidateConstraints(candidate)...)
	return reasoner.Input{
		RequestID:   "gw-post-" + shortHash(seedID),
		Goal:        pre.Goal,
		Facts:       facts,
		Evidence:    evidence,
		Constraints: constraints,
		Actions:     actions,
		Risk:        pre.Risk,
		Metadata:    pre.Metadata,
	}
}

func buildGoal(system string) *reasoner.Goal {
	if system != "" {
		return &reasoner.Goal{
			ID:          "g-system",
			Description: system,
			Priority:    "p1",
		}
	}
	// No system prompt → goal id present but description empty so Goal state
	// is 0.5; this lets downstream callers detect the under-specified case via
	// the score band without forcing a hard refusal.
	return &reasoner.Goal{
		ID:       "g-default",
		Priority: "p2",
	}
}

func buildUserFacts(userMessages []string) []reasoner.Fact {
	if len(userMessages) == 0 {
		return nil
	}
	facts := make([]reasoner.Fact, 0, len(userMessages))
	for i, m := range userMessages {
		facts = append(facts, reasoner.Fact{
			ID:    fmt.Sprintf("f-u%d", i+1),
			Key:   "user.message",
			Value: m,
		})
	}
	return facts
}

func buildPriorTurnEvidence(assistantMessages []string) []reasoner.Evidence {
	if len(assistantMessages) == 0 {
		return nil
	}
	ev := make([]reasoner.Evidence, 0, len(assistantMessages))
	for i, m := range assistantMessages {
		ev = append(ev, reasoner.Evidence{
			ID:         fmt.Sprintf("e-prior%d", i+1),
			Kind:       "prior-turn",
			Reference:  m,
			Confidence: 0.6,
			Verified:   false,
		})
	}
	return ev
}

func defaultActions() []reasoner.Action {
	return []reasoner.Action{
		{
			ID:         "a-default-generate",
			Name:       "generate-assistant-message",
			Reversible: true,
			Cost:       0.2,
		},
	}
}

// requestConstraints assembles the constraints derived from the request
// itself: the always-present satisfied-true soft default (GD-4) plus any
// hard constraints triggered by GD-5 keywords in system or user content.
func requestConstraints(system string, userMessages []string) []reasoner.Constraint {
	yes := true
	cs := []reasoner.Constraint{{
		ID:        "c-default-soft",
		Kind:      "policy",
		Severity:  "soft",
		Statement: "default soft constraint to keep mapper happy",
		Satisfied: &yes,
	}}
	scan := strings.ToLower(system + " " + strings.Join(userMessages, " "))
	cs = append(cs, scanRequestKeywords(scan)...)
	return cs
}

func scanRequestKeywords(loweredText string) []reasoner.Constraint {
	no := false
	var cs []reasoner.Constraint
	if strings.Contains(loweredText, "permission denied") {
		cs = append(cs, reasoner.Constraint{
			ID:        "c-permission-denied",
			Kind:      "permission",
			Severity:  "hard",
			Statement: "GD-5 keyword: permission denied",
			Satisfied: &no,
		})
	}
	if strings.Contains(loweredText, "forbidden") {
		cs = append(cs, reasoner.Constraint{
			ID:        "c-forbidden",
			Kind:      "forbid",
			Severity:  "hard",
			Statement: "GD-5 keyword: forbidden",
			Satisfied: &no,
		})
	}
	if strings.Contains(loweredText, "do not answer") {
		cs = append(cs, reasoner.Constraint{
			ID:        "c-do-not-answer",
			Kind:      "governance",
			Severity:  "hard",
			Statement: "GD-5 keyword: do not answer",
			Satisfied: &no,
		})
	}
	return cs
}

func candidateConstraints(candidate string) []reasoner.Constraint {
	no := false
	var cs []reasoner.Constraint
	upper := strings.ToUpper(candidate)
	if strings.Contains(upper, "FORCE_REJECT") {
		cs = append(cs, reasoner.Constraint{
			ID:        "c-candidate-force-reject",
			Kind:      "governance",
			Severity:  "hard",
			Statement: "GD-5 candidate keyword: FORCE_REJECT",
			Satisfied: &no,
		})
	}
	if strings.Contains(upper, "POLICY_VIOLATION") {
		cs = append(cs, reasoner.Constraint{
			ID:        "c-candidate-policy-violation",
			Kind:      "policy",
			Severity:  "hard",
			Statement: "GD-5 candidate keyword: POLICY_VIOLATION",
			Satisfied: &no,
		})
	}
	return cs
}

func shortHash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])[:16]
}
