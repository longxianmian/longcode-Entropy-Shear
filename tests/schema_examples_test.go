package tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestSchemaArtifactsExist asserts that every documented artifact ships
// with the repo and parses cleanly. It does not run a full JSON Schema
// validator (kept dep-free) but enforces structural invariants every
// artifact has to satisfy.
func TestSchemaArtifactsExist(t *testing.T) {
	root := repoRoot(t)

	t.Run("openapi.yaml present", func(t *testing.T) {
		path := filepath.Join(root, "openapi.yaml")
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("openapi.yaml missing: %v", err)
		}
		if info.Size() < 200 {
			t.Errorf("openapi.yaml suspiciously small: %d bytes", info.Size())
		}
	})

	schemas := []string{
		"policy.schema.json",
		"shear-request.schema.json",
		"shear-response.schema.json",
		"ledger-record.schema.json",
	}
	for _, name := range schemas {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(root, "schemas", name)
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", name, err)
			}
			var m map[string]interface{}
			if err := json.Unmarshal(raw, &m); err != nil {
				t.Fatalf("parse %s: %v", name, err)
			}
			if m["$schema"] == nil {
				t.Errorf("%s: missing $schema", name)
			}
			if m["title"] == nil {
				t.Errorf("%s: missing title", name)
			}
		})
	}
}

// TestExamplesAreValidJSON verifies every shipped example parses as JSON
// and contains the top-level shape required by shear-request.schema.json.
func TestExamplesAreValidJSON(t *testing.T) {
	root := repoRoot(t)
	files := []string{
		"cityone-request.json",
		"agent-action-request.json",
		"agent-action-yes-request.json",
		"agent-action-no-request.json",
		"agent-action-hold-request.json",
		"ai-customer-service-request.json",
		"bid-risk-request.json",
		"permission-gate-request.json",
	}
	for _, f := range files {
		t.Run(f, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join(root, "examples", f))
			if err != nil {
				t.Fatal(err)
			}
			var m map[string]interface{}
			if err := json.Unmarshal(raw, &m); err != nil {
				t.Fatalf("invalid JSON: %v", err)
			}
			if _, ok := m["policy"]; !ok {
				t.Errorf("%s: missing policy", f)
			}
			if _, ok := m["facts"]; !ok {
				t.Errorf("%s: missing facts", f)
			}
		})
	}
}
