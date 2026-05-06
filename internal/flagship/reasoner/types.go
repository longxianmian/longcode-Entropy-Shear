// Package reasoner is the public entry point of the flagship P0 龙码三态
// 逻辑推理内核. It re-exports the types used at the JSON boundary so callers
// see a single coherent API surface, and sub-packages stay free of cyclic
// imports.
package reasoner

import (
	"entropy-shear/internal/flagship/hold"
	"entropy-shear/internal/flagship/mapper"
	"entropy-shear/internal/flagship/output"
	"entropy-shear/internal/flagship/state"
)

// Re-exports of leaf-package types. Type aliases keep the JSON shape, the
// docstrings and the methods identical to the originating package.
type (
	Goal       = mapper.Goal
	Fact       = mapper.Fact
	Evidence   = mapper.Evidence
	Constraint = mapper.Constraint
	Action     = mapper.Action

	ElementStates = mapper.ElementStates

	Verdict = state.Verdict

	AlignmentTask     = hold.AlignmentTask
	PermitToken       = output.PermitToken
	RejectInstruction = output.RejectInstruction
	AuditRecord       = output.AuditRecord
)

// Verdict re-exports.
const (
	VerdictYes  = state.VerdictYes
	VerdictHold = state.VerdictHold
	VerdictNo   = state.VerdictNo
)

// Input is the multi-source input accepted by Reason and the
// /flagship/reason HTTP endpoint.
type Input struct {
	RequestID   string                 `json:"request_id"`
	Goal        *Goal                  `json:"goal,omitempty"`
	Facts       []Fact                 `json:"facts,omitempty"`
	Evidence    []Evidence             `json:"evidence,omitempty"`
	Constraints []Constraint           `json:"constraints,omitempty"`
	Actions     []Action               `json:"actions,omitempty"`
	Risk        string                 `json:"risk,omitempty"`
	Weights     map[string]float64     `json:"weights,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Output is the single response payload returned by Reason.
type Output struct {
	RequestID         string             `json:"request_id"`
	Verdict           Verdict            `json:"verdict"`
	Score             float64            `json:"score"`
	ElementStates     ElementStates      `json:"element_states"`
	NormalizedWeights map[string]float64 `json:"normalized_weights"`
	PermitToken       *PermitToken       `json:"permit_token,omitempty"`
	RejectInstruction *RejectInstruction `json:"reject_instruction,omitempty"`
	AlignmentTasks    []AlignmentTask    `json:"alignment_tasks,omitempty"`
	AuditRecord       AuditRecord        `json:"audit_record"`
	Trace             []string           `json:"trace,omitempty"`
}
