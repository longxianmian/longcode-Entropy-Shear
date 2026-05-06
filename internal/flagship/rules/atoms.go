// Package rules executes atomic validation rules over flagship inputs and
// reports triggered rule IDs, validation issues, hard penalties and the list
// of conflicting items that force the verdict to NO.
package rules

import "entropy-shear/internal/flagship/mapper"

// Reason codes for hard conflicts.
const (
	ReasonHardConstraintUnsatisfied = "FLAGSHIP_HARD_CONSTRAINT_UNSATISFIED"
	ReasonPermissionDenied          = "FLAGSHIP_PERMISSION_DENIED"
	ReasonExplicitForbid            = "FLAGSHIP_EXPLICIT_FORBID"
	ReasonGovernanceReject          = "FLAGSHIP_GOVERNANCE_REJECT"
)

// Atomic rule IDs (stable strings — referenced by tests and audit records).
const (
	RGoalPresent       = "R-GOAL-001"
	RFactKeyValue      = "R-FACT-001"
	REvidenceConf      = "R-EVIDENCE-001"
	REvidenceVerified  = "R-EVIDENCE-002"
	RConstraintSeverityLegal = "R-CONSTRAINT-001"
	RConstraintHardSatisfied = "R-CONSTRAINT-002"
	RActionName        = "R-ACTION-001"
)

// Result aggregates atomic rule outcomes.
type Result struct {
	TriggeredRuleIDs []string `json:"triggered_rule_ids"`
	ValidationIssues []string `json:"validation_issues,omitempty"`
	HardPenalty      float64  `json:"hard_penalty"`
	HasHardConflict  bool     `json:"has_hard_conflict"`
	ConflictingItems []string `json:"conflicting_items,omitempty"`
	HardReasonCode   string   `json:"hard_reason_code,omitempty"`
}

// Apply runs the atomic rules over the inputs.
//
// Hard conflicts (any one is sufficient to force NO):
//   - A constraint with severity == "hard" and Satisfied != true.
//   - A constraint whose kind is "permission" / "forbid" / "governance" with
//     Satisfied != true, regardless of stated severity.
func Apply(g *mapper.Goal, facts []mapper.Fact, ev []mapper.Evidence, cs []mapper.Constraint, acts []mapper.Action) Result {
	res := Result{}

	if g == nil || g.ID == "" {
		res.TriggeredRuleIDs = append(res.TriggeredRuleIDs, RGoalPresent)
		res.ValidationIssues = append(res.ValidationIssues, "goal missing or has empty id")
	}

	for _, f := range facts {
		if f.Key == "" || f.Value == nil {
			res.TriggeredRuleIDs = append(res.TriggeredRuleIDs, RFactKeyValue)
			res.ValidationIssues = append(res.ValidationIssues, "fact "+f.ID+" has empty key or nil value")
		}
	}

	var anyVerified bool
	for _, e := range ev {
		if e.Confidence < 0 || e.Confidence > 1 {
			res.TriggeredRuleIDs = append(res.TriggeredRuleIDs, REvidenceConf)
			res.ValidationIssues = append(res.ValidationIssues, "evidence "+e.ID+" confidence out of [0,1]")
		}
		if e.Verified {
			anyVerified = true
		}
	}
	if len(ev) > 0 && !anyVerified {
		res.TriggeredRuleIDs = append(res.TriggeredRuleIDs, REvidenceVerified)
		res.ValidationIssues = append(res.ValidationIssues, "no verified evidence among supplied evidence")
	}

	for _, c := range cs {
		legal := c.Severity == "soft" || c.Severity == "hard"
		if !legal {
			res.TriggeredRuleIDs = append(res.TriggeredRuleIDs, RConstraintSeverityLegal)
			res.ValidationIssues = append(res.ValidationIssues, "constraint "+c.ID+" has illegal severity: "+c.Severity)
		}
		satisfied := c.Satisfied != nil && *c.Satisfied
		governanceKind := c.Kind == "permission" || c.Kind == "forbid" || c.Kind == "governance"
		if (c.Severity == "hard" || governanceKind) && !satisfied {
			res.TriggeredRuleIDs = append(res.TriggeredRuleIDs, RConstraintHardSatisfied)
			res.HardPenalty++
			res.HasHardConflict = true
			res.ConflictingItems = append(res.ConflictingItems, c.ID)
			if res.HardReasonCode == "" {
				switch c.Kind {
				case "permission":
					res.HardReasonCode = ReasonPermissionDenied
				case "forbid":
					res.HardReasonCode = ReasonExplicitForbid
				case "governance":
					res.HardReasonCode = ReasonGovernanceReject
				default:
					res.HardReasonCode = ReasonHardConstraintUnsatisfied
				}
			}
		}
	}

	for _, a := range acts {
		if a.Name == "" {
			res.TriggeredRuleIDs = append(res.TriggeredRuleIDs, RActionName)
			res.ValidationIssues = append(res.ValidationIssues, "action "+a.ID+" has empty name")
		}
	}

	res.TriggeredRuleIDs = dedup(res.TriggeredRuleIDs)
	return res
}

func dedup(in []string) []string {
	if len(in) == 0 {
		return in
	}
	seen := make(map[string]struct{}, len(in))
	out := in[:0]
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
