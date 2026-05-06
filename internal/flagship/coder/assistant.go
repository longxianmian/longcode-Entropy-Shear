package coder

import (
	"fmt"
	"strings"

	"entropy-shear/internal/flagship/reasoner"
)

// stop_reason values per GD-3.
const (
	StopReasonEndTurn = "end_turn"
	StopReasonRefusal = "refusal"
)

// BuildAssistantContent returns the (text, stop_reason) pair to put in the
// assistant message for a given verdict. phase is "pre" or "post" so the
// surface text can tell users at which gate the verdict landed.
//
// Per GD-2 and GD-3:
//   YES  → candidate text, stop_reason=end_turn
//   HOLD → human-readable HOLD summary plus alignment-task list, stop_reason=end_turn
//   NO   → refusal text plus reject-instruction summary, stop_reason=refusal
func BuildAssistantContent(verdict string, candidate string, tasks []reasoner.AlignmentTask, reject *reasoner.RejectInstruction, phase string) (text string, stopReason string) {
	switch verdict {
	case string(reasoner.VerdictYes):
		return candidate, StopReasonEndTurn
	case string(reasoner.VerdictHold):
		return formatHoldText(tasks, phase), StopReasonEndTurn
	case string(reasoner.VerdictNo):
		return formatRefusalText(reject, phase), StopReasonRefusal
	default:
		return fmt.Sprintf("[unknown verdict %q at %s-governance]", verdict, phase), StopReasonEndTurn
	}
}

func formatHoldText(tasks []reasoner.AlignmentTask, phase string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "[%s-governance HOLD] entropy-shear flagship reasoner reached HOLD; %d alignment task(s):", phase, len(tasks))
	for _, t := range tasks {
		fmt.Fprintf(&b, "\n  - [%s/%s] %s — %s", t.Priority, t.TargetElement, t.ReasonCode, t.Prompt)
	}
	return b.String()
}

func formatRefusalText(reject *reasoner.RejectInstruction, phase string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "[%s-governance NO] entropy-shear flagship reasoner refused this request.", phase)
	if reject != nil {
		fmt.Fprintf(&b, " reason_code=%s", reject.ReasonCode)
		if len(reject.ConflictingItems) > 0 {
			fmt.Fprintf(&b, " conflicting_items=[%s]", strings.Join(reject.ConflictingItems, ", "))
		}
	}
	return b.String()
}
