package flagship_test

import (
	"math"
	"testing"

	"entropy-shear/internal/flagship/mapper"
)

func boolPtr(b bool) *bool { return &b }

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

func TestMapEmptyInputs(t *testing.T) {
	got := mapper.Map(nil, nil, nil, nil, nil)
	if got.Goal != 0 || got.Fact != 0 || got.Evidence != 0 || got.Action != 0 {
		t.Fatalf("expected zero states for missing inputs, got %+v", got)
	}
	// Constraint is unimpeded when no constraints are supplied.
	if got.Constraint != 1.0 {
		t.Fatalf("expected constraint state 1.0 with no constraints, got %v", got.Constraint)
	}
}

func TestMapGoalCompleteness(t *testing.T) {
	cases := []struct {
		name string
		in   *mapper.Goal
		want float64
	}{
		{"missing", nil, 0},
		{"empty id", &mapper.Goal{ID: ""}, 0},
		{"id only", &mapper.Goal{ID: "g-1"}, 0.5},
		{"id and description", &mapper.Goal{ID: "g-1", Description: "do x"}, 1.0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mapper.Map(tc.in, nil, nil, nil, nil)
			if got.Goal != tc.want {
				t.Fatalf("goal state: got %v want %v", got.Goal, tc.want)
			}
		})
	}
}

func TestMapEvidenceVerifiedWeight(t *testing.T) {
	ev := []mapper.Evidence{
		{ID: "e1", Kind: "x", Confidence: 0.8, Verified: true},
		{ID: "e2", Kind: "y", Confidence: 0.6, Verified: false},
	}
	got := mapper.Map(nil, nil, ev, nil, nil)
	want := (0.8*1.0 + 0.6*0.5) / 2.0 // 0.55
	if !almostEqual(got.Evidence, want) {
		t.Fatalf("evidence state: got %v want %v", got.Evidence, want)
	}
}

func TestMapConstraintSatisfiedRatio(t *testing.T) {
	cs := []mapper.Constraint{
		{ID: "c1", Kind: "policy", Severity: "soft", Satisfied: boolPtr(true)},
		{ID: "c2", Kind: "policy", Severity: "soft", Satisfied: boolPtr(false)},
		{ID: "c3", Kind: "policy", Severity: "soft"},
	}
	got := mapper.Map(nil, nil, nil, cs, nil)
	want := 1.0 / 3.0
	if !almostEqual(got.Constraint, want) {
		t.Fatalf("constraint state: got %v want %v", got.Constraint, want)
	}
}

func TestMapActionCostInverse(t *testing.T) {
	acts := []mapper.Action{
		{ID: "a1", Name: "x", Reversible: true, Cost: 0.2},
		{ID: "a2", Name: "y", Reversible: false, Cost: 0.6},
	}
	got := mapper.Map(nil, nil, nil, nil, acts)
	want := ((1.0 - 0.2) + (1.0 - 0.6)) / 2.0 // 0.6
	if !almostEqual(got.Action, want) {
		t.Fatalf("action state: got %v want %v", got.Action, want)
	}
}

func TestMapFactRatio(t *testing.T) {
	facts := []mapper.Fact{
		{ID: "f1", Key: "k1", Value: "v1"},
		{ID: "f2", Key: "", Value: "v2"},
		{ID: "f3", Key: "k3", Value: nil},
	}
	got := mapper.Map(nil, facts, nil, nil, nil)
	want := 1.0 / 3.0
	if !almostEqual(got.Fact, want) {
		t.Fatalf("fact state: got %v want %v", got.Fact, want)
	}
}
