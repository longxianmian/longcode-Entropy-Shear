package engine

import (
	"sort"

	"entropy-shear/internal/schema"
)

// Decision is the engine's pre-signature output. The API/handler is
// responsible for hashing, signing, and persisting it.
type Decision struct {
	Verdict       schema.Verdict
	AppliedRuleID *string
	Route         string
	Reason        string
	Trace         []schema.TraceItem
}

// Shear runs the chained-rule decision (§8/§9). Caller must validate
// inputs first; this function is pure and deterministic.
func Shear(policy schema.Policy, facts schema.Facts) Decision {
	rules := append([]schema.Rule(nil), policy.Rules...)
	sort.SliceStable(rules, func(i, j int) bool {
		return rules[i].Priority < rules[j].Priority
	})

	trace := make([]schema.TraceItem, 0, len(rules))
	for _, r := range rules {
		matched, detail := Evaluate(r.Condition, facts)
		trace = append(trace, schema.TraceItem{
			RuleID:    r.ID,
			Evaluated: true,
			Matched:   matched,
			Detail:    detail,
		})
		if matched {
			id := r.ID
			return Decision{
				Verdict:       r.Effect,
				AppliedRuleID: &id,
				Route:         r.Route,
				Reason:        r.Reason,
				Trace:         trace,
			}
		}
	}

	verdict := policy.DefaultEffect
	if verdict == "" {
		verdict = schema.Hold
	}
	reason := policy.DefaultReason
	if reason == "" {
		reason = "未命中规则，需人工复核"
	}
	return Decision{
		Verdict:       verdict,
		AppliedRuleID: nil,
		Reason:        reason,
		Trace:         trace,
	}
}
