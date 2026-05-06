package flagship_test

import (
	"testing"

	"entropy-shear/internal/flagship/mapper"
	"entropy-shear/internal/flagship/state"
)

func TestDefaultMatrixFixed(t *testing.T) {
	// R1: matrix values are frozen for P0; this guard fails if anyone edits
	// a single cell.
	want := [5][5]float64{
		{0.00, 0.10, 0.10, 0.20, -0.60},
		{0.10, 0.00, 0.80, 0.10, 0.20},
		{0.10, 0.20, 0.00, 0.30, 0.70},
		{0.20, 0.10, -0.40, 0.00, -0.80},
		{0.30, 0.40, 0.30, 0.10, 0.00},
	}
	if state.DefaultMatrix != want {
		t.Fatalf("DefaultMatrix changed; expected R1 frozen values\n got: %v\nwant: %v", state.DefaultMatrix, want)
	}
}

func TestInteractionFactorBranches(t *testing.T) {
	if got := state.InteractionFactor(0.3); got != state.FactorPositivePromotion {
		t.Fatalf("positive cell: got %v", got)
	}
	if got := state.InteractionFactor(-0.4); got != state.FactorRestraintBalance {
		t.Fatalf("negative cell: got %v", got)
	}
	if got := state.InteractionFactor(0); got != state.FactorNoEffect {
		t.Fatalf("zero cell: got %v", got)
	}
}

func TestResolveWeightsFallback(t *testing.T) {
	cases := []struct {
		name string
		in   map[string]float64
		fall bool
	}{
		{"nil → defaults, no fallback flag", nil, false},
		{"missing key", map[string]float64{"Goal": 0.2, "Fact": 0.2, "Evidence": 0.2, "Constraint": 0.2}, true},
		{"negative", map[string]float64{"Goal": 0.2, "Fact": 0.2, "Evidence": -0.1, "Constraint": 0.2, "Action": 0.1}, true},
		{"sum zero", map[string]float64{"Goal": 0, "Fact": 0, "Evidence": 0, "Constraint": 0, "Action": 0}, true},
		{"valid", map[string]float64{"Goal": 1, "Fact": 1, "Evidence": 1, "Constraint": 1, "Action": 1}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, fell, _ := state.ResolveWeights(tc.in)
			if fell != tc.fall {
				t.Fatalf("fellBack: got %v want %v", fell, tc.fall)
			}
		})
	}
}

func TestApplyRiskBoostNormalizes(t *testing.T) {
	w := map[string]float64{"Goal": 0.25, "Fact": 0.20, "Evidence": 0.25, "Constraint": 0.20, "Action": 0.10}
	state.ApplyRiskBoost(w, "high")
	var sum float64
	for _, v := range w {
		sum += v
	}
	if abs(sum-1.0) > 1e-9 {
		t.Fatalf("normalized weights must sum to 1; got %v", sum)
	}
	// At r=0.75 Evidence/Constraint should be boosted relative to defaults
	// after normalization, and Action should shrink.
	if w["Evidence"] <= 0.25 {
		t.Fatalf("expected Evidence boosted, got %v", w["Evidence"])
	}
	if w["Constraint"] <= 0.20 {
		t.Fatalf("expected Constraint boosted, got %v", w["Constraint"])
	}
	if w["Action"] >= 0.10 {
		t.Fatalf("expected Action decayed, got %v", w["Action"])
	}
}

func TestComputeFSMVerdicts(t *testing.T) {
	weights := map[string]float64{"Goal": 0.25, "Fact": 0.20, "Evidence": 0.25, "Constraint": 0.20, "Action": 0.10}
	cases := []struct {
		name            string
		states          mapper.ElementStates
		hardPenalty     float64
		hasHardConflict bool
		want            state.Verdict
	}{
		{
			name:   "all-high → YES",
			states: mapper.ElementStates{Goal: 1, Fact: 1, Evidence: 1, Constraint: 1, Action: 1},
			want:   state.VerdictYes,
		},
		{
			name:   "all-zero, no penalty → NO (score below T2)",
			states: mapper.ElementStates{Goal: 0, Fact: 0, Evidence: 0, Constraint: 0, Action: 0},
			want:   state.VerdictNo,
		},
		{
			name:            "hard conflict overrides high score → NO",
			states:          mapper.ElementStates{Goal: 1, Fact: 1, Evidence: 1, Constraint: 1, Action: 1},
			hardPenalty:     1.0,
			hasHardConflict: true,
			want:            state.VerdictNo,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := copyMap(weights)
			got := state.Compute(tc.states, w, tc.hardPenalty, tc.hasHardConflict)
			if got.Verdict != tc.want {
				t.Fatalf("verdict: got %v score=%v want %v", got.Verdict, got.Score, tc.want)
			}
		})
	}
}

func TestComputeHoldBand(t *testing.T) {
	// Hand-tuned states designed to land between T2 and T1. With Goal=0.5,
	// Fact=1.0, Evidence=0.15, Constraint=0, Action=0.3 and default weights
	// (medium risk-modulated) the score is in the HOLD band — verified
	// numerically against the round-1 decision sheet.
	weights := map[string]float64{"Goal": 0.25, "Fact": 0.20, "Evidence": 0.25, "Constraint": 0.20, "Action": 0.10}
	state.ApplyRiskBoost(weights, "medium")
	states := mapper.ElementStates{Goal: 0.5, Fact: 1.0, Evidence: 0.15, Constraint: 0, Action: 0.3}
	got := state.Compute(states, weights, 0, false)
	if got.Verdict != state.VerdictHold {
		t.Fatalf("expected HOLD, got %v at score %v", got.Verdict, got.Score)
	}
	if got.Score < state.T2 || got.Score >= state.T1 {
		t.Fatalf("score %v not in [%v, %v)", got.Score, state.T2, state.T1)
	}
}

func TestComputeNormalizedWeightsSumOne(t *testing.T) {
	weights, _, _ := state.ResolveWeights(nil)
	state.ApplyRiskBoost(weights, "critical")
	got := state.Compute(mapper.ElementStates{Goal: 1, Fact: 1, Evidence: 1, Constraint: 1, Action: 1}, weights, 0, false)
	var sum float64
	for _, v := range got.NormalizedWeights {
		sum += v
	}
	if abs(sum-1.0) > 1e-9 {
		t.Fatalf("normalized weights must sum to 1; got %v", sum)
	}
}

func copyMap(m map[string]float64) map[string]float64 {
	out := make(map[string]float64, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
