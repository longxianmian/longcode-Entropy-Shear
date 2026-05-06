// Package hold generates AlignmentTask objects when the FSM reaches HOLD.
//
// Each AlignmentTask points at a single five-element gap and tells the caller
// what to supply, verify, confirm or downgrade so the next reasoning round
// can produce YES or NO.
package hold

import (
	"sort"
	"strconv"

	"entropy-shear/internal/flagship/mapper"
	"entropy-shear/internal/flagship/rules"
)

// AlignmentTask is the P0 minimal field set per R4.
// requested_from is intentionally omitted (R4 v0).
type AlignmentTask struct {
	ID                   string `json:"id"`
	TargetElement        string `json:"target_element"`
	ReasonCode           string `json:"reason_code"`
	Gap                  string `json:"gap"`
	RequiredAction       string `json:"required_action"`
	ExpectedEvidenceKind string `json:"expected_evidence_kind,omitempty"`
	Prompt               string `json:"prompt"`
	Priority             string `json:"priority"`
}

const (
	ReasonGoalMissing       = "FLAGSHIP_GOAL_MISSING"
	ReasonFactInsufficient  = "FLAGSHIP_FACT_INSUFFICIENT"
	ReasonEvidenceWeak      = "FLAGSHIP_EVIDENCE_WEAK"
	ReasonConstraintPending = "FLAGSHIP_CONSTRAINT_PENDING"
	ReasonActionUnclear     = "FLAGSHIP_ACTION_UNCLEAR"
	ReasonGenericGap        = "FLAGSHIP_GENERIC_GAP"
)

// Thresholds below which an element triggers an alignment task. Kept loose so
// HOLD always emits at least one actionable task.
const (
	thresholdLow = 0.50
)

// Generate emits one or more AlignmentTask records based on which elements
// are weakest. HOLD must always yield at least one task; if none of the
// per-element gaps fire, a generic gap targeting the lowest-state element is
// emitted as a fallback.
func Generate(requestID string, states mapper.ElementStates, ruleRes rules.Result,
	g *mapper.Goal, facts []mapper.Fact, ev []mapper.Evidence, cs []mapper.Constraint, acts []mapper.Action) []AlignmentTask {
	var tasks []AlignmentTask
	idx := 0
	nextID := func() string {
		idx++
		return "align-" + requestID + "-" + strconv.Itoa(idx)
	}

	if g == nil || g.ID == "" {
		tasks = append(tasks, AlignmentTask{
			ID:             nextID(),
			TargetElement:  "Goal",
			ReasonCode:     ReasonGoalMissing,
			Gap:            "goal is missing or has empty id",
			RequiredAction: "supply a goal with stable id and human-readable description",
			Prompt:         "请补充本轮目标的 id 和描述。",
			Priority:       "high",
		})
	}

	if states.Fact < thresholdLow {
		tasks = append(tasks, AlignmentTask{
			ID:                   nextID(),
			TargetElement:        "Fact",
			ReasonCode:           ReasonFactInsufficient,
			Gap:                  "fact set is empty or several facts have empty key/value",
			RequiredAction:       "supply more facts with non-empty key and non-nil value",
			ExpectedEvidenceKind: "structured-fact",
			Prompt:               "请补充关键事实，保证每条 fact 都有 key 与 value。",
			Priority:             "medium",
		})
	}

	if states.Evidence < thresholdLow {
		tasks = append(tasks, AlignmentTask{
			ID:                   nextID(),
			TargetElement:        "Evidence",
			ReasonCode:           ReasonEvidenceWeak,
			Gap:                  "evidence is missing, low-confidence or unverified",
			RequiredAction:       "supply verified evidence with confidence >= 0.7",
			ExpectedEvidenceKind: "verified-source",
			Prompt:               "请补充已核验的证据来源，置信度需达 0.7 以上。",
			Priority:             "high",
		})
	}

	if hasUnsatisfiedSoftConstraint(cs) {
		tasks = append(tasks, AlignmentTask{
			ID:             nextID(),
			TargetElement:  "Constraint",
			ReasonCode:     ReasonConstraintPending,
			Gap:            "one or more soft constraints are still unsatisfied",
			RequiredAction: "confirm, downgrade or remediate the pending soft constraints",
			Prompt:         "请确认或降级未满足的软约束，逐项标注 satisfied=true。",
			Priority:       "medium",
		})
	}

	if states.Action < thresholdLow {
		tasks = append(tasks, AlignmentTask{
			ID:             nextID(),
			TargetElement:  "Action",
			ReasonCode:     ReasonActionUnclear,
			Gap:            "no action proposed or proposed actions have high cost",
			RequiredAction: "specify a reversible action with lower cost or split it into smaller steps",
			Prompt:         "请明确可回滚、低成本的行动方案。",
			Priority:       "medium",
		})
	}

	if len(tasks) == 0 {
		tasks = append(tasks, AlignmentTask{
			ID:             nextID(),
			TargetElement:  weakestElement(states),
			ReasonCode:     ReasonGenericGap,
			Gap:            "score is in the HOLD band but no specific gap was detected",
			RequiredAction: "review the weakest element and supply additional context",
			Prompt:         "请补充最弱五元的输入信息以推进结论。",
			Priority:       "low",
		})
	}

	// Sort by priority (high → medium → low) for stable output.
	sort.SliceStable(tasks, func(i, j int) bool {
		return priorityRank(tasks[i].Priority) < priorityRank(tasks[j].Priority)
	})
	return tasks
}

func hasUnsatisfiedSoftConstraint(cs []mapper.Constraint) bool {
	for _, c := range cs {
		if c.Severity == "soft" && (c.Satisfied == nil || !*c.Satisfied) {
			return true
		}
	}
	return false
}

func weakestElement(s mapper.ElementStates) string {
	type kv struct {
		name  string
		value float64
	}
	pairs := []kv{
		{"Goal", s.Goal},
		{"Fact", s.Fact},
		{"Evidence", s.Evidence},
		{"Constraint", s.Constraint},
		{"Action", s.Action},
	}
	min := pairs[0]
	for _, p := range pairs[1:] {
		if p.value < min.value {
			min = p
		}
	}
	return min.name
}

func priorityRank(p string) int {
	switch p {
	case "high":
		return 0
	case "medium":
		return 1
	case "low":
		return 2
	default:
		return 3
	}
}
