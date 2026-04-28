package tests

import (
	"testing"

	"entropy-shear/internal/signature"
)

func TestSumIsDeterministicAcrossKeyOrder(t *testing.T) {
	a := map[string]interface{}{
		"b": 1, "a": 2, "c": map[string]interface{}{"y": 1, "x": 2},
	}
	// Same data, different in-memory key insertion order.
	b := map[string]interface{}{
		"c": map[string]interface{}{"x": 2, "y": 1}, "a": 2, "b": 1,
	}
	ha, err := signature.Sum(a)
	if err != nil {
		t.Fatal(err)
	}
	hb, err := signature.Sum(b)
	if err != nil {
		t.Fatal(err)
	}
	if ha != hb {
		t.Fatalf("canonical encoding not stable: %s vs %s", ha, hb)
	}
}

func TestCanonicalJSONHasSortedKeys(t *testing.T) {
	got, err := signature.CanonicalJSON(map[string]interface{}{
		"b": 1, "a": "x",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := `{"a":"x","b":1}`
	if string(got) != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
