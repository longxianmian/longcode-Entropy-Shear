// Package output builds the verdict-attached artifacts: PermitToken (YES),
// RejectInstruction (NO) and AuditRecord. P0 deliberately uses no
// cryptographic signing — the Core signature / ledger paths are out of scope
// per LONGMA_TASK_ANCHOR.json round-1 boundary.
package output

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"
)

// PermitToken is the YES-side artifact (R5 P0 fields, no signing, no ledger).
type PermitToken struct {
	ID         string   `json:"id"`
	Verdict    string   `json:"verdict"`
	Scope      []string `json:"scope"`
	ValidUntil string   `json:"valid_until"`
	ReasonCode string   `json:"reason_code"`
	AuditID    string   `json:"audit_id"`
}

// RejectInstruction is the NO-side artifact (R5 P0 fields).
type RejectInstruction struct {
	ID                string   `json:"id"`
	Verdict           string   `json:"verdict"`
	ReasonCode        string   `json:"reason_code"`
	ConflictingItems  []string `json:"conflicting_items"`
	RemediationSteps  []string `json:"remediation_steps"`
	AuditID           string   `json:"audit_id"`
}

// DefaultTokenValidity is the lifespan applied to a PermitToken when the
// caller does not override it.
const DefaultTokenValidity = time.Hour

// NewPermitToken builds a PermitToken bound to the given audit record.
// The id is derived from the SHA-256 of (requestID + auditID) so the same
// pair is reproducible across runs.
func NewPermitToken(requestID, auditID, reasonCode string, scope []string, now time.Time, validity time.Duration) PermitToken {
	if validity <= 0 {
		validity = DefaultTokenValidity
	}
	if scope == nil {
		scope = []string{}
	}
	return PermitToken{
		ID:         "ptok-" + shortHash(requestID+"|"+auditID),
		Verdict:    "YES",
		Scope:      scope,
		ValidUntil: now.UTC().Add(validity).Format(time.RFC3339),
		ReasonCode: reasonCode,
		AuditID:    auditID,
	}
}

// NewRejectInstruction builds a RejectInstruction bound to the given audit
// record. remediationSteps must be a non-nil string slice (R5).
func NewRejectInstruction(requestID, auditID, reasonCode string, conflicting, remediation []string) RejectInstruction {
	if conflicting == nil {
		conflicting = []string{}
	}
	if remediation == nil {
		remediation = []string{}
	}
	return RejectInstruction{
		ID:               "rej-" + shortHash(requestID+"|"+auditID+"|"+reasonCode+"|"+strconv.Itoa(len(conflicting))),
		Verdict:          "NO",
		ReasonCode:       reasonCode,
		ConflictingItems: conflicting,
		RemediationSteps: remediation,
		AuditID:          auditID,
	}
}

func shortHash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])[:16]
}
