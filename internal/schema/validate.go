package schema

import (
	"fmt"
	"strings"

	apperr "entropy-shear/internal/errors"
)

// ValidatePolicy enforces the structural contract of a policy (§13).
// Returns *errors.APIError with the exact code from the spec on violation.
func ValidatePolicy(p *Policy) error {
	if p == nil {
		return apperr.New(422, apperr.CodePolicySchemaViolation, "policy is required")
	}
	if strings.TrimSpace(p.ID) == "" {
		return apperr.New(422, apperr.CodePolicySchemaViolation, "policy.id is required")
	}
	if strings.TrimSpace(p.Version) == "" {
		return apperr.New(422, apperr.CodePolicySchemaViolation, "policy.version is required")
	}
	if !isValidVerdict(p.DefaultEffect) {
		return apperr.New(422, apperr.CodePolicySchemaViolation,
			fmt.Sprintf("policy.default_effect must be Yes/No/Hold, got %q", p.DefaultEffect))
	}
	if strings.TrimSpace(p.DefaultReason) == "" {
		return apperr.New(422, apperr.CodePolicySchemaViolation, "policy.default_reason is required")
	}
	if p.Rules == nil {
		return apperr.New(422, apperr.CodePolicySchemaViolation, "policy.rules must be an array (use [] for none)")
	}
	seen := make(map[string]struct{}, len(p.Rules))
	for i, r := range p.Rules {
		if err := validateRule(i, &r); err != nil {
			return err
		}
		if _, dup := seen[r.ID]; dup {
			return apperr.New(422, apperr.CodePolicySchemaViolation,
				fmt.Sprintf("policy.rules[%d].id %q is duplicated", i, r.ID))
		}
		seen[r.ID] = struct{}{}
	}
	return nil
}

func validateRule(idx int, r *Rule) error {
	prefix := fmt.Sprintf("policy.rules[%d]", idx)
	if strings.TrimSpace(r.ID) == "" {
		return apperr.New(422, apperr.CodePolicySchemaViolation, prefix+".id is required")
	}
	if !isValidVerdict(r.Effect) {
		return apperr.New(422, apperr.CodePolicySchemaViolation,
			fmt.Sprintf("%s.effect must be Yes/No/Hold, got %q", prefix, r.Effect))
	}
	if strings.TrimSpace(r.Reason) == "" {
		return apperr.New(422, apperr.CodePolicySchemaViolation, prefix+".reason is required")
	}
	if strings.TrimSpace(r.Condition.Field) == "" {
		return apperr.New(422, apperr.CodePolicySchemaViolation, prefix+".condition.field is required")
	}
	if _, ok := SupportedOperators[r.Condition.Operator]; !ok {
		return apperr.New(422, apperr.CodeUnsupportedOperator,
			fmt.Sprintf("%s.condition.operator %q is not supported", prefix, r.Condition.Operator))
	}
	return nil
}

// ValidateFacts requires Facts to be a non-nil object. P0 keeps the contract
// minimal: Facts is free-form JSON, but it must be present and an object.
func ValidateFacts(f Facts) error {
	if f == nil {
		return apperr.New(422, apperr.CodeFactsSchemaViolation, "facts must be an object")
	}
	return nil
}

func isValidVerdict(v Verdict) bool {
	return v == Yes || v == No || v == Hold
}
