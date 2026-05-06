// Package state holds the 5x5 evaluation matrix, the five-element
// interference model (Lite), the weighted conflict-resolution score, and
// the YES / HOLD / NO finite state machine for the flagship P0 reasoner.
//
// All numeric constants are fixed for P0 by the round-1 decision sheet
// (R1, R2, R3) and must not be tuned silently.
package state

import "entropy-shear/internal/flagship/mapper"

// Verdict is the three-state output of the reasoner.
type Verdict string

const (
	VerdictYes  Verdict = "YES"
	VerdictHold Verdict = "HOLD"
	VerdictNo   Verdict = "NO"
)

// Element index order is fixed: Goal, Fact, Evidence, Constraint, Action.
const (
	IdxGoal       = 0
	IdxFact       = 1
	IdxEvidence   = 2
	IdxConstraint = 3
	IdxAction     = 4
)

// ElementNames matches the index order above. Used for normalised weight maps
// and for trace lines.
var ElementNames = [5]string{"Goal", "Fact", "Evidence", "Constraint", "Action"}

// DefaultMatrix is the P0 5x5 dynamic-influence matrix M[i][j] = effect of
// Ei on Ej. Values come from R1 of the round-1 decision sheet and must not
// be edited.
var DefaultMatrix = [5][5]float64{
	// Goal,  Fact,  Evid,  Cons,  Action
	{0.00, 0.10, 0.10, 0.20, -0.60}, // Goal
	{0.10, 0.00, 0.80, 0.10, 0.20},  // Fact
	{0.10, 0.20, 0.00, 0.30, 0.70},  // Evidence
	{0.20, 0.10, -0.40, 0.00, -0.80}, // Constraint
	{0.30, 0.40, 0.30, 0.10, 0.00},  // Action
}

// Interference factors (R2). FactorOverSuppression and FactorReverseImbalance
// are reserved constants — P0 does not auto-classify these patterns; they are
// declared so that downstream stages can raise them without reshaping the API.
const (
	FactorPositivePromotion = 1.00
	FactorRestraintBalance  = 0.80
	FactorOverSuppression   = 1.20
	FactorReverseImbalance  = 1.30
	FactorNoEffect          = 0.00
)

// Score / FSM constants (R3).
const (
	Lambda = 0.20
	Mu     = 1.00
	T1     = 0.70
	T2     = 0.35
)

// DefaultWeights are the base five-element weights (R3). Sum to 1.0.
var DefaultWeights = map[string]float64{
	"Goal":       0.25,
	"Fact":       0.20,
	"Evidence":   0.25,
	"Constraint": 0.20,
	"Action":     0.10,
}

// Risk levels resolve to a factor r in [0, 1] used by ApplyRiskBoost.
var RiskFactors = map[string]float64{
	"low":      0.25,
	"medium":   0.50,
	"high":     0.75,
	"critical": 1.00,
}

// Computation is the result of Compute.
type Computation struct {
	Score             float64            `json:"score"`
	Verdict           Verdict            `json:"verdict"`
	NormalizedWeights map[string]float64 `json:"normalized_weights"`
	WeightedSum       float64            `json:"weighted_sum"`
	InterferenceSum   float64            `json:"interference_sum"`
	HardPenalty       float64            `json:"hard_penalty"`
	HasHardConflict   bool               `json:"has_hard_conflict"`
}

// ResolveWeights validates an optional override against the contract from R3
// (five keys, all >= 0, sum > 0) and falls back to DefaultWeights otherwise.
// fellBack is true when the override was rejected. The returned map is a fresh
// copy; callers may mutate it (e.g. to apply risk boosts).
func ResolveWeights(override map[string]float64) (weights map[string]float64, fellBack bool, reason string) {
	if override == nil {
		return copyWeights(DefaultWeights), false, ""
	}
	for _, name := range ElementNames {
		v, ok := override[name]
		if !ok {
			return copyWeights(DefaultWeights), true, "weights override missing key " + name
		}
		if v < 0 {
			return copyWeights(DefaultWeights), true, "weights override has negative value for " + name
		}
	}
	var sum float64
	for _, name := range ElementNames {
		sum += override[name]
	}
	if sum <= 0 {
		return copyWeights(DefaultWeights), true, "weights override sums to zero"
	}
	out := make(map[string]float64, 5)
	for _, name := range ElementNames {
		out[name] = override[name]
	}
	return out, false, ""
}

// ApplyRiskBoost mutates weights with R3 risk modulation
// (Evidence_boost, Constraint_boost, Action_decay) and re-normalises so the
// sum stays at 1. RiskLevel is resolved against RiskFactors; an unknown level
// yields r = 0 (no modulation).
func ApplyRiskBoost(weights map[string]float64, riskLevel string) {
	r, ok := RiskFactors[riskLevel]
	if !ok {
		r = 0
	}
	weights["Evidence"] *= 1.0 + 0.30*r
	weights["Constraint"] *= 1.0 + 0.40*r
	weights["Action"] *= 1.0 - 0.20*r
	if weights["Action"] < 0 {
		weights["Action"] = 0
	}
	normalize(weights)
}

func normalize(weights map[string]float64) {
	var sum float64
	for _, name := range ElementNames {
		sum += weights[name]
	}
	if sum <= 0 {
		// Degenerate; restore defaults to avoid NaN.
		for k, v := range DefaultWeights {
			weights[k] = v
		}
		return
	}
	for _, name := range ElementNames {
		weights[name] /= sum
	}
}

// InteractionFactor maps a matrix cell sign to the P0 interference factor.
// P0 only uses the positive-promotion and restraint-balance branches; the
// over-suppression and reverse-imbalance branches are reserved.
func InteractionFactor(m float64) float64 {
	switch {
	case m > 0:
		return FactorPositivePromotion
	case m < 0:
		return FactorRestraintBalance
	default:
		return FactorNoEffect
	}
}

// Relation = min(state(Ei), state(Ej)) * InteractionFactor(M[i][j]).
func Relation(stateI, stateJ, m float64) float64 {
	return min2(stateI, stateJ) * InteractionFactor(m)
}

// Compute evaluates the score and the FSM verdict.
//
// Score = Σ Wi * state(Ei) + λ * Σ M[i][j] * Relation(Ei, Ej) - μ * HardPenalty
//
// Verdict:
//   - HasHardConflict == true              → NO
//   - Score >= T1 and no hard conflict     → YES
//   - T2 <= Score < T1                     → HOLD
//   - Score < T2                           → NO
func Compute(states mapper.ElementStates, weights map[string]float64, hardPenalty float64, hasHardConflict bool) Computation {
	stateVec := [5]float64{states.Goal, states.Fact, states.Evidence, states.Constraint, states.Action}

	var weighted float64
	for i, name := range ElementNames {
		weighted += weights[name] * stateVec[i]
	}

	var interference float64
	for i := 0; i < 5; i++ {
		for j := 0; j < 5; j++ {
			m := DefaultMatrix[i][j]
			if m == 0 {
				continue
			}
			interference += m * Relation(stateVec[i], stateVec[j], m)
		}
	}

	score := weighted + Lambda*interference - Mu*hardPenalty

	verdict := VerdictNo
	switch {
	case hasHardConflict:
		verdict = VerdictNo
	case score >= T1:
		verdict = VerdictYes
	case score >= T2:
		verdict = VerdictHold
	}

	return Computation{
		Score:             score,
		Verdict:           verdict,
		NormalizedWeights: copyWeights(weights),
		WeightedSum:       weighted,
		InterferenceSum:   interference,
		HardPenalty:       hardPenalty,
		HasHardConflict:   hasHardConflict,
	}
}

func copyWeights(in map[string]float64) map[string]float64 {
	out := make(map[string]float64, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func min2(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
