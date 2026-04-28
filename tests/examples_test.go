package tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"entropy-shear/internal/engine"
	"entropy-shear/internal/schema"
)

// repoRoot resolves to the project root regardless of where `go test` is
// invoked, by walking up from the test file's working directory.
func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// tests are in /tests, so go up one level.
	return filepath.Dir(wd)
}

func loadExample(t *testing.T, name string) schema.ShearRequest {
	t.Helper()
	path := filepath.Join(repoRoot(t), "examples", name)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var req schema.ShearRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}
	if err := schema.ValidatePolicy(&req.Policy); err != nil {
		t.Fatalf("validate %s: %v", name, err)
	}
	if err := schema.ValidateFacts(req.Facts); err != nil {
		t.Fatalf("facts %s: %v", name, err)
	}
	return req
}

// TestCanonicalExamples checks that every shipped example file evaluates
// to the verdict its filename promises in the README / OpenAPI table.
func TestCanonicalExamples(t *testing.T) {
	cases := []struct {
		file    string
		verdict schema.Verdict
		ruleID  string // optional — empty means don't assert
	}{
		{"cityone-request.json", schema.Yes, "rule-001"},
		{"agent-action-request.json", schema.Yes, "allow-readonly-action"},
		{"agent-action-yes-request.json", schema.Yes, "allow-readonly-action"},
		{"agent-action-no-request.json", schema.No, "forbid-delete-production-data"},
		{"agent-action-hold-request.json", schema.Hold, "hold-payment-action"},
		{"ai-customer-service-request.json", schema.Hold, "need-order-id-before-human"},
		{"bid-risk-request.json", schema.No, "missing-required-file"},
		{"permission-gate-request.json", schema.Yes, "vip-or-member-allow"},
	}
	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			req := loadExample(t, tc.file)
			d := engine.Shear(req.Policy, req.Facts)
			if d.Verdict != tc.verdict {
				t.Fatalf("%s verdict=%s want %s; trace=%+v", tc.file, d.Verdict, tc.verdict, d.Trace)
			}
			if tc.ruleID != "" {
				if d.AppliedRuleID == nil || *d.AppliedRuleID != tc.ruleID {
					t.Errorf("%s applied=%v want %s", tc.file, d.AppliedRuleID, tc.ruleID)
				}
			}
		})
	}
}

// TestAgentActionAllVerdicts exercises the agent-action policy across the
// three verdicts using alternate facts. Proves the policy itself emits
// Yes / No / Hold, not just one canonical verdict.
func TestAgentActionAllVerdicts(t *testing.T) {
	req := loadExample(t, "agent-action-request.json")

	yesFacts := schema.Facts{"action": map[string]interface{}{"name": "list_users", "category": "data", "mode": "readonly"}}
	noFacts := schema.Facts{"action": map[string]interface{}{"name": "delete_production_data", "category": "data", "mode": "write"}}
	holdPay := schema.Facts{"action": map[string]interface{}{"name": "transfer", "category": "payment", "mode": "write"}}
	holdUnk := schema.Facts{"action": map[string]interface{}{"name": "noop", "category": "misc", "mode": "write"}}

	mustVerdict(t, "yes-readonly", req.Policy, yesFacts, schema.Yes, "allow-readonly-action")
	mustVerdict(t, "no-delete-prod", req.Policy, noFacts, schema.No, "forbid-delete-production-data")
	mustVerdict(t, "hold-payment", req.Policy, holdPay, schema.Hold, "hold-payment-action")
	mustVerdict(t, "hold-default", req.Policy, holdUnk, schema.Hold, "")
}

// TestAICustomerServiceAllVerdicts confirms the AI policy emits all three
// verdicts.
func TestAICustomerServiceAllVerdicts(t *testing.T) {
	req := loadExample(t, "ai-customer-service-request.json")

	holdRefund := schema.Facts{
		"intent":    map[string]interface{}{"type": "refund"},
		"risk":      map[string]interface{}{"category": "general"},
		"knowledge": map[string]interface{}{"hit_confidence": 0.5},
	}
	noMedical := schema.Facts{
		"intent":    map[string]interface{}{"type": "general"},
		"risk":      map[string]interface{}{"category": "medical_advice"},
		"knowledge": map[string]interface{}{"hit_confidence": 0.5},
	}
	yesFAQ := schema.Facts{
		"intent":    map[string]interface{}{"type": "faq"},
		"risk":      map[string]interface{}{"category": "general"},
		"knowledge": map[string]interface{}{"hit_confidence": 0.95},
	}
	holdLowConf := schema.Facts{
		"intent":    map[string]interface{}{"type": "faq"},
		"risk":      map[string]interface{}{"category": "general"},
		"knowledge": map[string]interface{}{"hit_confidence": 0.3},
	}

	mustVerdict(t, "hold-refund", req.Policy, holdRefund, schema.Hold, "need-order-id-before-human")
	mustVerdict(t, "no-medical", req.Policy, noMedical, schema.No, "forbidden-medical-advice")
	mustVerdict(t, "yes-faq", req.Policy, yesFAQ, schema.Yes, "faq-answerable")
	mustVerdict(t, "hold-default", req.Policy, holdLowConf, schema.Hold, "")
}

// TestBidRiskAllVerdicts confirms the bid-risk policy emits all three.
func TestBidRiskAllVerdicts(t *testing.T) {
	req := loadExample(t, "bid-risk-request.json")

	noMissing := schema.Facts{
		"bid":           map[string]interface{}{"required_files_missing_count": 2},
		"qualification": map[string]interface{}{"evidence_complete": true},
		"risk":          map[string]interface{}{"critical_issue_count": 0},
	}
	holdQual := schema.Facts{
		"bid":           map[string]interface{}{"required_files_missing_count": 0},
		"qualification": map[string]interface{}{"evidence_complete": false},
		"risk":          map[string]interface{}{"critical_issue_count": 0},
	}
	yesAllPass := schema.Facts{
		"bid":           map[string]interface{}{"required_files_missing_count": 0},
		"qualification": map[string]interface{}{"evidence_complete": true},
		"risk":          map[string]interface{}{"critical_issue_count": 0},
	}
	holdResidual := schema.Facts{
		"bid":           map[string]interface{}{"required_files_missing_count": 0},
		"qualification": map[string]interface{}{"evidence_complete": true},
		"risk":          map[string]interface{}{"critical_issue_count": 3},
	}

	mustVerdict(t, "no-missing-file", req.Policy, noMissing, schema.No, "missing-required-file")
	mustVerdict(t, "hold-incomplete-qual", req.Policy, holdQual, schema.Hold, "uncertain-qualification")
	mustVerdict(t, "yes-all-pass", req.Policy, yesAllPass, schema.Yes, "all-critical-checks-pass")
	mustVerdict(t, "hold-default", req.Policy, holdResidual, schema.Hold, "")
}

// TestPermissionGateAllVerdicts confirms the gate emits all three.
func TestPermissionGateAllVerdicts(t *testing.T) {
	req := loadExample(t, "permission-gate-request.json")

	yesVIP := schema.Facts{"user": map[string]interface{}{"id": "U-1", "level": "vip", "tags": []interface{}{}}}
	noBlocked := schema.Facts{"user": map[string]interface{}{"id": "U-2", "level": "trial", "tags": []interface{}{"blocked"}}}
	holdUnknown := schema.Facts{"user": map[string]interface{}{"id": "U-3", "level": "trial", "tags": []interface{}{}}}

	mustVerdict(t, "yes-vip", req.Policy, yesVIP, schema.Yes, "vip-or-member-allow")
	mustVerdict(t, "no-blocked", req.Policy, noBlocked, schema.No, "blocked-tag-deny")
	mustVerdict(t, "hold-unknown", req.Policy, holdUnknown, schema.Hold, "")
}

func mustVerdict(t *testing.T, name string, p schema.Policy, f schema.Facts, want schema.Verdict, ruleID string) {
	t.Helper()
	d := engine.Shear(p, f)
	if d.Verdict != want {
		t.Errorf("%s verdict=%s want %s detail=%+v", name, d.Verdict, want, d.Trace)
		return
	}
	if ruleID == "" {
		if d.AppliedRuleID != nil {
			t.Errorf("%s applied=%v want nil (default fall-through)", name, *d.AppliedRuleID)
		}
		return
	}
	if d.AppliedRuleID == nil || *d.AppliedRuleID != ruleID {
		t.Errorf("%s applied=%v want %s", name, d.AppliedRuleID, ruleID)
	}
}
