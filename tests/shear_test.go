package tests

import (
	"testing"

	"entropy-shear/internal/engine"
	"entropy-shear/internal/schema"
)

func cityOnePolicy() schema.Policy {
	return schema.Policy{
		ID:      "policy-cityone-v1",
		Version: "1.0.0",
		Rules: []schema.Rule{
			{
				ID: "rule-001", Priority: 1, Effect: schema.Yes,
				Route: "/campaign/member-landing",
				Condition: schema.Condition{
					Field: "user.level", Operator: schema.OpIn,
					Value: []interface{}{"member", "vip"},
				},
				Reason: "会员或 VIP 用户允许进入会员活动页",
			},
			{
				ID: "rule-002", Priority: 2, Effect: schema.No,
				Condition: schema.Condition{
					Field: "user.tags", Operator: schema.OpContains,
					Value: "blocked",
				},
				Reason: "用户命中黑名单标签，拒绝访问",
			},
		},
		DefaultEffect: schema.Hold,
		DefaultReason: "未命中任何规则，需人工复核",
	}
}

func TestFirstRuleHitYes(t *testing.T) {
	got := engine.Shear(cityOnePolicy(), schema.Facts{
		"user": map[string]interface{}{"level": "member", "tags": []interface{}{}},
	})
	if got.Verdict != schema.Yes {
		t.Fatalf("verdict=%s want Yes", got.Verdict)
	}
	if got.AppliedRuleID == nil || *got.AppliedRuleID != "rule-001" {
		t.Fatalf("applied_rule_id mismatch: %v", got.AppliedRuleID)
	}
	if got.Route != "/campaign/member-landing" {
		t.Fatalf("route mismatch: %q", got.Route)
	}
	if len(got.Trace) != 1 {
		t.Fatalf("trace must short-circuit: got %d items", len(got.Trace))
	}
}

func TestSecondRuleHitNo(t *testing.T) {
	got := engine.Shear(cityOnePolicy(), schema.Facts{
		"user": map[string]interface{}{"level": "trial", "tags": []interface{}{"blocked"}},
	})
	if got.Verdict != schema.No {
		t.Fatalf("verdict=%s want No", got.Verdict)
	}
	if got.AppliedRuleID == nil || *got.AppliedRuleID != "rule-002" {
		t.Fatalf("applied_rule_id=%v want rule-002", got.AppliedRuleID)
	}
	if len(got.Trace) != 2 {
		t.Fatalf("trace must include rule-001 (miss) and rule-002 (hit): got %d", len(got.Trace))
	}
	if got.Trace[0].Matched {
		t.Errorf("rule-001 should not match for trial user")
	}
	if !got.Trace[1].Matched {
		t.Errorf("rule-002 should match for blocked tag")
	}
}

func TestAllMissDefaultHold(t *testing.T) {
	got := engine.Shear(cityOnePolicy(), schema.Facts{
		"user": map[string]interface{}{"level": "trial", "tags": []interface{}{}},
	})
	if got.Verdict != schema.Hold {
		t.Fatalf("verdict=%s want Hold", got.Verdict)
	}
	if got.AppliedRuleID != nil {
		t.Fatalf("applied_rule_id must be nil on default fall-through, got %v", *got.AppliedRuleID)
	}
	if got.Reason != "未命中任何规则，需人工复核" {
		t.Fatalf("default_reason mismatch: %q", got.Reason)
	}
}

func TestPriorityOrderingRespected(t *testing.T) {
	policy := schema.Policy{
		ID: "p", Version: "1.0.0",
		Rules: []schema.Rule{
			// declared in reverse priority — engine must sort.
			{ID: "high-prio-no", Priority: 1, Effect: schema.No,
				Condition: schema.Condition{Field: "x", Operator: schema.OpEq, Value: "a"},
				Reason:    "blocked first"},
			{ID: "low-prio-yes", Priority: 10, Effect: schema.Yes,
				Condition: schema.Condition{Field: "x", Operator: schema.OpEq, Value: "a"},
				Reason:    "would allow if reached"},
		},
		DefaultEffect: schema.Hold, DefaultReason: "fallthrough",
	}
	got := engine.Shear(policy, schema.Facts{"x": "a"})
	if got.Verdict != schema.No {
		t.Fatalf("priority 1 must win, got %s", got.Verdict)
	}
	if *got.AppliedRuleID != "high-prio-no" {
		t.Fatalf("applied=%s want high-prio-no", *got.AppliedRuleID)
	}
}

func TestMissingFieldDoesNotCrash(t *testing.T) {
	got := engine.Shear(cityOnePolicy(), schema.Facts{
		"user": map[string]interface{}{},
	})
	if got.Verdict != schema.Hold {
		t.Fatalf("missing fields should fall through to default Hold, got %s", got.Verdict)
	}
	if len(got.Trace) != 2 {
		t.Fatalf("each rule must still be evaluated and traced: got %d", len(got.Trace))
	}
}

func TestRuleEffectHoldAllowed(t *testing.T) {
	// §14.1 example uses Hold as a rule effect — it must be a legal verdict.
	policy := schema.Policy{
		ID: "p", Version: "1.0.0",
		Rules: []schema.Rule{
			{ID: "r1", Priority: 1, Effect: schema.Hold,
				Condition: schema.Condition{Field: "intent.type", Operator: schema.OpEq, Value: "refund"},
				Reason:    "needs order id"},
		},
		DefaultEffect: schema.Yes, DefaultReason: "ok",
	}
	got := engine.Shear(policy, schema.Facts{"intent": map[string]interface{}{"type": "refund"}})
	if got.Verdict != schema.Hold {
		t.Fatalf("rule with Hold effect must produce Hold, got %s", got.Verdict)
	}
}
