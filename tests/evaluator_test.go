package tests

import (
	"testing"

	"entropy-shear/internal/engine"
	"entropy-shear/internal/schema"
)

// Operator coverage matches §15.1 of the spec — one row per documented case.
func TestOperators(t *testing.T) {
	cases := []struct {
		name  string
		field string
		op    schema.Operator
		val   interface{}
		facts schema.Facts
		want  bool
	}{
		{"eq hit", "user.level", schema.OpEq, "vip",
			schema.Facts{"user": map[string]interface{}{"level": "vip"}}, true},
		{"eq miss", "user.level", schema.OpEq, "vip",
			schema.Facts{"user": map[string]interface{}{"level": "trial"}}, false},
		{"neq hit", "user.status", schema.OpNeq, "banned",
			schema.Facts{"user": map[string]interface{}{"status": "active"}}, true},
		{"gt hit", "order.amount", schema.OpGt, 100.0,
			schema.Facts{"order": map[string]interface{}{"amount": 200.0}}, true},
		{"lt hit", "order.amount", schema.OpLt, 100.0,
			schema.Facts{"order": map[string]interface{}{"amount": 50.0}}, true},
		{"gte hit equal", "user.age", schema.OpGte, 18.0,
			schema.Facts{"user": map[string]interface{}{"age": 18.0}}, true},
		{"lte hit equal", "stock", schema.OpLte, 5.0,
			schema.Facts{"stock": 5.0}, true},
		{"in hit", "user.level", schema.OpIn, []interface{}{"vip", "member"},
			schema.Facts{"user": map[string]interface{}{"level": "member"}}, true},
		{"in miss", "user.level", schema.OpIn, []interface{}{"vip", "member"},
			schema.Facts{"user": map[string]interface{}{"level": "trial"}}, false},
		{"contains array hit", "user.tags", schema.OpContains, "blocked",
			schema.Facts{"user": map[string]interface{}{"tags": []interface{}{"blocked"}}}, true},
		{"contains array miss", "user.tags", schema.OpContains, "blocked",
			schema.Facts{"user": map[string]interface{}{"tags": []interface{}{"vip"}}}, false},
		{"contains string hit", "text", schema.OpContains, "refund",
			schema.Facts{"text": "i need a refund please"}, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, detail := engine.Evaluate(schema.Condition{
				Field: tc.field, Operator: tc.op, Value: tc.val,
			}, tc.facts)
			if got != tc.want {
				t.Fatalf("matched=%v want=%v detail=%s", got, tc.want, detail)
			}
			if detail == "" {
				t.Errorf("trace detail must not be empty")
			}
		})
	}
}

func TestMissingPathReturnsFalse(t *testing.T) {
	got, detail := engine.Evaluate(schema.Condition{
		Field: "user.profile.level", Operator: schema.OpEq, Value: "vip",
	}, schema.Facts{"user": map[string]interface{}{}})
	if got {
		t.Fatalf("missing path should be false, detail=%s", detail)
	}
}

func TestTypeMismatchReturnsFalseNotPanic(t *testing.T) {
	got, detail := engine.Evaluate(schema.Condition{
		Field: "user.age", Operator: schema.OpGt, Value: "not-a-number",
	}, schema.Facts{"user": map[string]interface{}{"age": 30.0}})
	if got {
		t.Fatalf("type mismatch should be false, detail=%s", detail)
	}
	if detail == "" {
		t.Errorf("trace detail must explain mismatch")
	}
}

func TestUnknownOperatorReturnsFalse(t *testing.T) {
	got, detail := engine.Evaluate(schema.Condition{
		Field: "user.level", Operator: schema.Operator("regex"), Value: ".*",
	}, schema.Facts{"user": map[string]interface{}{"level": "vip"}})
	if got {
		t.Fatalf("unknown op must be false, detail=%s", detail)
	}
}

func TestNumericComparisonAcrossIntFloat(t *testing.T) {
	got, _ := engine.Evaluate(schema.Condition{
		Field: "n", Operator: schema.OpEq, Value: 5,
	}, schema.Facts{"n": 5.0})
	if !got {
		t.Errorf("int 5 should equal float 5.0")
	}
}
