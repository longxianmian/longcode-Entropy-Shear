// Package coder bridges Anthropic-style coder requests to the frozen
// flagship reasoner kernel. It owns request/candidate adaptation, the
// assistant-message wrapping rules and the GatewayAuditRecord shape.
//
// The package only depends on the frozen reasoner package; it does not
// import the gateway HTTP package, so types declared here can be safely
// referenced from gateway/types.go without creating an import cycle.
package coder

import (
	"time"

	"entropy-shear/internal/flagship/reasoner"
)

// GatewayID is the stable identifier the network gateway stamps on every
// audit record so consumers can tell which deployment produced it.
const GatewayID = "flagship-coder-gateway-p0"

// GatewayAuditRecord aggregates the pre-governance and (optionally) the
// post-governance reasoner audits for a single coder request, plus
// provider-side metadata. P0 returns this as a JSON object on the response
// — it is not written to the Core ledger.
type GatewayAuditRecord struct {
	GatewayID         string                `json:"gateway_id"`
	RequestID         string                `json:"request_id"`
	PreReasonerAudit  reasoner.AuditRecord  `json:"pre_reasoner_audit"`
	PostReasonerAudit *reasoner.AuditRecord `json:"post_reasoner_audit,omitempty"`
	ProviderName      string                `json:"provider_name"`
	ProviderTraceID   string                `json:"provider_trace_id,omitempty"`
	Verdict           string                `json:"verdict"`
	Timestamp         string                `json:"timestamp"`
}

// NewGatewayAuditRecord assembles the record from already-built reasoner
// audits. PostReasonerAudit may be nil for HOLD/NO short-circuited at
// pre-governance.
func NewGatewayAuditRecord(requestID, providerName, providerTraceID, verdict string,
	pre reasoner.AuditRecord, post *reasoner.AuditRecord, now time.Time) GatewayAuditRecord {
	return GatewayAuditRecord{
		GatewayID:         GatewayID,
		RequestID:         requestID,
		PreReasonerAudit:  pre,
		PostReasonerAudit: post,
		ProviderName:      providerName,
		ProviderTraceID:   providerTraceID,
		Verdict:           verdict,
		Timestamp:         now.UTC().Format(time.RFC3339),
	}
}
