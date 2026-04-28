package engine

import (
	"encoding/json"
	"fmt"
	"strings"

	"entropy-shear/internal/schema"
)

// Evaluate runs a single condition against facts and returns (matched, detail).
// Per §10/§13: missing path or type mismatch yields false (not an error), and
// any internal panic is caught and surfaced into the trace detail so the
// service cannot crash on a single bad rule.
func Evaluate(cond schema.Condition, facts schema.Facts) (matched bool, detail string) {
	defer func() {
		if r := recover(); r != nil {
			matched = false
			detail = fmt.Sprintf("evaluation panic on field=%s op=%s: %v", cond.Field, cond.Operator, r)
		}
	}()

	left, exists := lookupPath(facts, cond.Field)
	if !exists {
		return false, fmt.Sprintf("%s 路径不存在，视为 false", cond.Field)
	}

	switch cond.Operator {
	case schema.OpEq:
		return equalsOp(cond.Field, left, cond.Value, true)
	case schema.OpNeq:
		return equalsOp(cond.Field, left, cond.Value, false)
	case schema.OpGt:
		return numberOp(cond.Field, left, cond.Value, ">")
	case schema.OpLt:
		return numberOp(cond.Field, left, cond.Value, "<")
	case schema.OpGte:
		return numberOp(cond.Field, left, cond.Value, ">=")
	case schema.OpLte:
		return numberOp(cond.Field, left, cond.Value, "<=")
	case schema.OpIn:
		return inOp(cond.Field, left, cond.Value)
	case schema.OpContains:
		return containsOp(cond.Field, left, cond.Value)
	default:
		return false, fmt.Sprintf("未知操作符 %q", cond.Operator)
	}
}

// lookupPath walks a dot-separated path through nested maps. Numeric path
// segments index into slices. Returns (value, true) when found.
func lookupPath(root interface{}, path string) (interface{}, bool) {
	if path == "" {
		return nil, false
	}
	parts := strings.Split(path, ".")
	var cur interface{} = root
	for _, p := range parts {
		switch v := cur.(type) {
		case map[string]interface{}:
			next, ok := v[p]
			if !ok {
				return nil, false
			}
			cur = next
		case schema.Facts:
			next, ok := v[p]
			if !ok {
				return nil, false
			}
			cur = next
		default:
			return nil, false
		}
	}
	return cur, true
}

func equalsOp(field string, left, right interface{}, want bool) (bool, string) {
	eq, ok := looseEqual(left, right)
	if !ok {
		return false, fmt.Sprintf("%s 与 value 类型不可比较，视为 false", field)
	}
	matched := eq == want
	op := "=="
	if !want {
		op = "!="
	}
	return matched, fmt.Sprintf("%s %s %v，实际值为 %v，%s",
		field, op, right, left, hitText(matched))
}

// looseEqual compares two values with JSON-friendly semantics. Returns
// (equal, comparable). Numbers are compared as float64; bools and strings
// only compare to their own kind.
func looseEqual(a, b interface{}) (bool, bool) {
	if a == nil && b == nil {
		return true, true
	}
	if a == nil || b == nil {
		return false, true
	}
	if af, aok := toFloat(a); aok {
		if bf, bok := toFloat(b); bok {
			return af == bf, true
		}
		return false, false
	}
	if as, aok := a.(string); aok {
		if bs, bok := b.(string); bok {
			return as == bs, true
		}
		return false, false
	}
	if ab, aok := a.(bool); aok {
		if bb, bok := b.(bool); bok {
			return ab == bb, true
		}
		return false, false
	}
	return false, false
}

func numberOp(field string, left, right interface{}, op string) (bool, string) {
	lf, lok := toFloat(left)
	rf, rok := toFloat(right)
	if !lok || !rok {
		return false, fmt.Sprintf("%s 或 value 非数值类型，视为 false", field)
	}
	var matched bool
	switch op {
	case ">":
		matched = lf > rf
	case "<":
		matched = lf < rf
	case ">=":
		matched = lf >= rf
	case "<=":
		matched = lf <= rf
	}
	return matched, fmt.Sprintf("%s %s %v，实际值为 %v，%s",
		field, op, right, left, hitText(matched))
}

func inOp(field string, left, right interface{}) (bool, string) {
	arr, ok := right.([]interface{})
	if !ok {
		return false, fmt.Sprintf("%s in 操作的右值必须是数组，视为 false", field)
	}
	for _, item := range arr {
		if eq, comparable := looseEqual(left, item); comparable && eq {
			return true, fmt.Sprintf("%s in %v，实际值为 %v，命中", field, right, left)
		}
	}
	return false, fmt.Sprintf("%s in %v，实际值为 %v，未命中", field, right, left)
}

func containsOp(field string, left, right interface{}) (bool, string) {
	switch lv := left.(type) {
	case []interface{}:
		for _, item := range lv {
			if eq, comparable := looseEqual(item, right); comparable && eq {
				return true, fmt.Sprintf("%s contains %v，命中", field, right)
			}
		}
		return false, fmt.Sprintf("%s contains %v，未命中", field, right)
	case string:
		needle, ok := right.(string)
		if !ok {
			return false, fmt.Sprintf("%s 是字符串但 value 非字符串，视为 false", field)
		}
		matched := strings.Contains(lv, needle)
		return matched, fmt.Sprintf("%s contains %q，%s", field, needle, hitText(matched))
	default:
		return false, fmt.Sprintf("%s 非数组或字符串，contains 视为 false", field)
	}
}

// toFloat coerces JSON numeric forms (and Go's int families) to float64.
// Handles json.Number for streams decoded with UseNumber(), and rejects
// boolean values explicitly so true!=1.0.
func toFloat(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case json.Number:
		f, err := n.Float64()
		if err != nil {
			return 0, false
		}
		return f, true
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	default:
		return 0, false
	}
}

func hitText(matched bool) string {
	if matched {
		return "命中"
	}
	return "未命中"
}
