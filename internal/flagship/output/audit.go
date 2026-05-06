package output

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

// AuditRecord is the in-memory audit structure (R6 P0 — no JSONL, no Core
// ledger linkage).
type AuditRecord struct {
	AuditID             string   `json:"audit_id"`
	RequestID           string   `json:"request_id"`
	InputDigest         string   `json:"input_digest"`
	FiveElementDigest   string   `json:"five_element_digest"`
	TriggeredRuleIDs    []string `json:"triggered_rule_ids"`
	MatrixDigest        string   `json:"matrix_digest"`
	StateMachineResult  string   `json:"state_machine_result"`
	PermitTokenID       string   `json:"permit_token_id,omitempty"`
	RejectInstructionID string   `json:"reject_instruction_id,omitempty"`
	Timestamp           string   `json:"timestamp"`
}

// SHA256Hex returns the SHA-256 hex digest of bytes; nil/empty input maps to
// the canonical hash of an empty payload.
func SHA256Hex(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// CanonicalJSON is a thin wrapper around json.Marshal. It is "canonical
// enough" for P0: Go's encoding/json sorts map keys and renders structs in
// declared field order, which is deterministic across runs.
func CanonicalJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// NewAuditRecord builds an AuditRecord. inputBytes / fiveElementBytes /
// matrixBytes must already be canonical-JSON encoded by the caller; this
// keeps the digest input explicit and auditable.
func NewAuditRecord(requestID string, inputBytes, fiveElementBytes, matrixBytes []byte,
	triggeredRuleIDs []string, verdict string, now time.Time) AuditRecord {
	if triggeredRuleIDs == nil {
		triggeredRuleIDs = []string{}
	}
	inputDigest := SHA256Hex(inputBytes)
	fiveDigest := SHA256Hex(fiveElementBytes)
	matrixDigest := SHA256Hex(matrixBytes)
	auditID := "audit-" + shortHash(requestID+"|"+inputDigest+"|"+fiveDigest+"|"+verdict)
	return AuditRecord{
		AuditID:            auditID,
		RequestID:          requestID,
		InputDigest:        inputDigest,
		FiveElementDigest:  fiveDigest,
		TriggeredRuleIDs:   triggeredRuleIDs,
		MatrixDigest:       matrixDigest,
		StateMachineResult: verdict,
		Timestamp:          now.UTC().Format(time.RFC3339),
	}
}
