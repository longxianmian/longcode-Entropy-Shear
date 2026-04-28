package tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"entropy-shear/internal/engine"
	"entropy-shear/internal/policy"
	"entropy-shear/internal/schema"
)

// TestPolicyPackFilesValidate makes sure every *.v<n>.json file under
// policies/ passes the same schema check the running engine applies.
func TestPolicyPackFilesValidate(t *testing.T) {
	root := repoRoot(t)
	packs := walkPolicyFiles(t, filepath.Join(root, "policies"))
	if len(packs) == 0 {
		t.Fatal("no policy pack files found under policies/")
	}
	for _, p := range packs {
		t.Run(rel(root, p), func(t *testing.T) {
			rep, err := policy.ValidateFile(p)
			if err != nil {
				t.Fatalf("validate %s: %v", p, err)
			}
			if !rep.OK {
				t.Fatalf("%s did not validate: %s — %s", p, rep.Error, rep.Detail)
			}
			if rep.PolicyID == "" || rep.PolicyVersion == "" {
				t.Errorf("%s missing id/version in report: %+v", p, rep)
			}
		})
	}
}

// TestManifestMatchesDisk verifies that every entry in policies/manifest.json
// names a real file and that the recorded hash matches the canonical hash
// re-derived from disk. This is the contract that keeps the manifest
// from silently drifting.
func TestManifestMatchesDisk(t *testing.T) {
	root := repoRoot(t)
	m, err := policy.LoadManifest(filepath.Join(root, "policies", "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if m.ManifestVersion == "" || len(m.Policies) == 0 {
		t.Fatalf("manifest looks empty: %+v", m)
	}

	// Every manifest entry's hash must match the on-disk recomputation.
	res, err := policy.VerifyManifest(root, m)
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK {
		t.Fatalf("manifest mismatch: missing=%v mismatched=%+v", res.Missing, res.Mismatched)
	}

	// Every on-disk policy file must be listed in the manifest.
	listed := map[string]struct{}{}
	for _, e := range m.Policies {
		listed[e.Path] = struct{}{}
	}
	walked := walkPolicyFiles(t, filepath.Join(root, "policies"))
	var orphans []string
	for _, full := range walked {
		r := rel(root, full)
		if _, ok := listed[r]; !ok {
			orphans = append(orphans, r)
		}
	}
	sort.Strings(orphans)
	if len(orphans) > 0 {
		t.Errorf("policy files not listed in manifest.json: %v", orphans)
	}
}

// TestManifestEntryFieldsAreReasonable checks the meta fields aren't blank.
func TestManifestEntryFieldsAreReasonable(t *testing.T) {
	root := repoRoot(t)
	m, err := policy.LoadManifest(filepath.Join(root, "policies", "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range m.Policies {
		t.Run(e.Path, func(t *testing.T) {
			if e.PolicyID == "" || e.PolicyVersion == "" {
				t.Error("policy_id or policy_version missing")
			}
			if !strings.HasPrefix(e.Hash, "sha256:") || len(e.Hash) != len("sha256:")+64 {
				t.Errorf("hash format unexpected: %q", e.Hash)
			}
			if e.Scenario == "" {
				t.Error("scenario missing")
			}
			if e.Maintainer == "" {
				t.Error("maintainer missing")
			}
			if e.CreatedAt == "" {
				t.Error("created_at missing")
			}
		})
	}
}

// TestAICustomerServiceFactsExamples confirms the three companion facts
// payloads under integrations/ai-customer-service-gate/ produce the
// verdicts their README promises when paired with the canonical pack.
func TestAICustomerServiceFactsExamples(t *testing.T) {
	root := repoRoot(t)
	policyPath := filepath.Join(root, "policies", "ai-customer-service",
		"ai-customer-service-policy.v1.json")
	p, err := policy.LoadAndValidate(policyPath)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		file    string
		verdict schema.Verdict
		ruleID  string
	}{
		{"refund-missing-order.json",   schema.Hold, "need-order-id-before-human"},
		{"high-risk-medical.json",      schema.No,   "forbidden-medical-advice"},
		{"faq-high-confidence.json",    schema.Yes,  "faq-answerable"},
	}
	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			path := filepath.Join(root, "integrations", "ai-customer-service-gate",
				"facts-examples", tc.file)
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var f schema.Facts
			if err := json.Unmarshal(raw, &f); err != nil {
				t.Fatalf("parse %s: %v", path, err)
			}
			d := engine.Shear(p, f)
			if d.Verdict != tc.verdict {
				t.Fatalf("%s verdict=%s want %s; trace=%+v", tc.file, d.Verdict, tc.verdict, d.Trace)
			}
			if d.AppliedRuleID == nil || *d.AppliedRuleID != tc.ruleID {
				t.Fatalf("%s applied=%v want %s", tc.file, d.AppliedRuleID, tc.ruleID)
			}
		})
	}
}

// TestValidateFileRejectsMalformedPolicy makes sure cmd/validate-policy's
// underlying logic returns ok=false for an obviously invalid policy.
func TestValidateFileRejectsMalformedPolicy(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad-policy.json")
	if err := os.WriteFile(bad, []byte(`{"id":"x","version":"1","rules":[{"id":"r","priority":1,"condition":{"field":"a","operator":"regex","value":".*"},"effect":"Yes","reason":"x"}],"default_effect":"Hold","default_reason":"x"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	rep, err := policy.ValidateFile(bad)
	if err != nil {
		t.Fatalf("validate (i/o): %v", err)
	}
	if rep.OK {
		t.Fatal("regex operator should not validate")
	}
	if rep.Error != "unsupported_operator" {
		t.Errorf("error code=%q want unsupported_operator", rep.Error)
	}
}

// TestHashFileReportRejectsInvalidPolicy makes sure cmd/hash-policy refuses
// to emit a hash for an unverified pre-image.
func TestHashFileReportRejectsInvalidPolicy(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(bad, []byte(`{"id":"","version":"","rules":[],"default_effect":"Yes","default_reason":""}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := policy.HashFileReport(bad); err == nil {
		t.Fatal("expected error for invalid policy")
	}
}

// TestHashFileIsStableAcrossFormatting hashes the same logical policy
// written with different whitespace/key-order and asserts equal hashes.
func TestHashFileIsStableAcrossFormatting(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.json")
	b := filepath.Join(dir, "b.json")
	if err := os.WriteFile(a, []byte(`{"id":"p","version":"1.0.0","rules":[{"id":"r","priority":1,"condition":{"field":"x","operator":"==","value":"a"},"effect":"Yes","reason":"ok"}],"default_effect":"Hold","default_reason":"x"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte(`{
  "rules": [
    {
      "reason": "ok",
      "priority": 1,
      "condition": { "operator": "==", "field": "x", "value": "a" },
      "id": "r",
      "effect": "Yes"
    }
  ],
  "version": "1.0.0",
  "default_reason": "x",
  "id": "p",
  "default_effect": "Hold"
}`), 0o644); err != nil {
		t.Fatal(err)
	}
	ha, err := policy.HashFile(a)
	if err != nil {
		t.Fatal(err)
	}
	hb, err := policy.HashFile(b)
	if err != nil {
		t.Fatal(err)
	}
	if ha != hb {
		t.Fatalf("hash drifted across formatting: %s vs %s", ha, hb)
	}
}

func walkPolicyFiles(t *testing.T, root string) []string {
	t.Helper()
	var out []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		base := info.Name()
		// only *.v<n>.json under each scenario folder; manifest.json excluded.
		if base == "manifest.json" || !strings.HasSuffix(base, ".json") {
			return nil
		}
		out = append(out, path)
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	sort.Strings(out)
	return out
}

func rel(root, full string) string {
	r, err := filepath.Rel(root, full)
	if err != nil {
		return full
	}
	return filepath.ToSlash(r)
}
