// Package policy provides shared loading, validation, and hashing
// helpers used by the cmd/validate-policy and cmd/hash-policy CLIs as
// well as the policy-pack tests. It is purely additive — none of the
// P0/P1 packages depend on it.
package policy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	apperr "entropy-shear/internal/errors"
	"entropy-shear/internal/schema"
	"entropy-shear/internal/signature"
)

// Load reads a policy file and decodes it. Unknown fields are rejected
// so policy packs can't quietly accumulate orphaned keys.
func Load(path string) (schema.Policy, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return schema.Policy{}, fmt.Errorf("read %s: %w", path, err)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var p schema.Policy
	if err := dec.Decode(&p); err != nil {
		return schema.Policy{}, fmt.Errorf("decode %s: %w", path, err)
	}
	return p, nil
}

// LoadAndValidate runs Load + schema.ValidatePolicy — the same gate the
// running /shear endpoint applies on every request.
func LoadAndValidate(path string) (schema.Policy, error) {
	p, err := Load(path)
	if err != nil {
		return schema.Policy{}, err
	}
	if err := schema.ValidatePolicy(&p); err != nil {
		return schema.Policy{}, err
	}
	return p, nil
}

// Hash returns the canonical SHA-256 of a policy. The pre-image is the
// stable JSON encoding of the policy struct (sorted keys at every depth,
// HTML escaping disabled). It is independent of source-file formatting.
func Hash(p schema.Policy) (string, error) {
	return signature.Sum(p)
}

// HashFile is a convenience: load + hash, without validation. Useful
// for archival / change-detection of unverified inputs.
func HashFile(path string) (string, error) {
	p, err := Load(path)
	if err != nil {
		return "", err
	}
	return Hash(p)
}

// ValidateReport is the structured result emitted by cmd/validate-policy.
type ValidateReport struct {
	OK            bool   `json:"ok"`
	PolicyID      string `json:"policy_id,omitempty"`
	PolicyVersion string `json:"version,omitempty"`
	RuleCount     int    `json:"rule_count,omitempty"`
	Error         string `json:"error,omitempty"`
	Detail        string `json:"detail,omitempty"`
}

// HashReport is the structured result emitted by cmd/hash-policy.
type HashReport struct {
	PolicyID      string `json:"policy_id"`
	PolicyVersion string `json:"version"`
	Hash          string `json:"hash"`
}

// ValidateFile is the function the CLI shells over. Returns a report
// suitable for direct JSON marshalling. err is reserved for I/O
// failures; schema violations live inside the report (ok=false).
func ValidateFile(path string) (ValidateReport, error) {
	p, err := LoadAndValidate(path)
	if err != nil {
		var ae *apperr.APIError
		if errors.As(err, &ae) {
			return ValidateReport{OK: false, Error: ae.Code, Detail: ae.Detail}, nil
		}
		return ValidateReport{
			OK:     false,
			Error:  "policy_validation_error",
			Detail: err.Error(),
		}, nil
	}
	return ValidateReport{
		OK:            true,
		PolicyID:      p.ID,
		PolicyVersion: p.Version,
		RuleCount:     len(p.Rules),
	}, nil
}

// HashFileReport loads + validates + hashes. Hashing an invalid policy
// is intentionally rejected — an unverified pre-image would not be
// trustworthy as a signed artifact.
func HashFileReport(path string) (HashReport, error) {
	p, err := LoadAndValidate(path)
	if err != nil {
		return HashReport{}, err
	}
	h, err := Hash(p)
	if err != nil {
		return HashReport{}, err
	}
	return HashReport{
		PolicyID:      p.ID,
		PolicyVersion: p.Version,
		Hash:          h,
	}, nil
}
