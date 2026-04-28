package signature

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"entropy-shear/internal/schema"
)

// HashPrefix is prepended to every hex digest emitted by this package.
const HashPrefix = "sha256:"

// Sum returns "sha256:<hex>" for the canonical encoding of v.
// Canonical = stable_json_encode (object keys sorted lexicographically,
// HTML escaping disabled, no trailing newline).
func Sum(v interface{}) (string, error) {
	buf, err := canonicalJSON(v)
	if err != nil {
		return "", err
	}
	d := sha256.Sum256(buf)
	return HashPrefix + hex.EncodeToString(d[:]), nil
}

// SumBytes hashes raw bytes (already canonical) and returns "sha256:<hex>".
func SumBytes(b []byte) string {
	d := sha256.Sum256(b)
	return HashPrefix + hex.EncodeToString(d[:])
}

// CanonicalJSON exposes canonicalJSON for callers that need to inspect or
// store the canonical representation alongside its hash.
func CanonicalJSON(v interface{}) ([]byte, error) {
	return canonicalJSON(v)
}

// HashTrace produces a stable hash over a trace slice (§11.1: trace hash).
func HashTrace(trace []schema.TraceItem) (string, error) {
	return Sum(trace)
}

// HashFacts produces a stable hash over the facts (§11.1: facts hash).
func HashFacts(f schema.Facts) (string, error) {
	return Sum(f)
}

// ResultSignaturePayload is exactly what enters the response signature.
// Per §11.3 it intentionally omits timestamp, so identical inputs against
// the same previous_hash always produce the same signature.
type ResultSignaturePayload struct {
	PolicyID      string  `json:"policy_id"`
	PolicyVersion string  `json:"policy_version"`
	InputHash     string  `json:"input_hash"`
	Verdict       string  `json:"verdict"`
	AppliedRuleID *string `json:"applied_rule_id"`
	TraceHash     string  `json:"trace_hash"`
	PreviousHash  string  `json:"previous_shear_hash"`
}

// SignResult returns the response signature, "sha256:..." per §11.2.
func SignResult(p ResultSignaturePayload) (string, error) {
	return Sum(p)
}

// LedgerHashPayload is the canonical pre-image for current_shear_hash on
// disk. Per §11.3 it includes timestamp so the chain advances even when
// the same decision is replayed.
type LedgerHashPayload struct {
	ShearID           string  `json:"shear_id"`
	Timestamp         string  `json:"timestamp"`
	PolicyID          string  `json:"policy_id"`
	PolicyVersion     string  `json:"policy_version"`
	InputHash         string  `json:"input_hash"`
	Verdict           string  `json:"verdict"`
	AppliedRuleID     *string `json:"applied_rule_id"`
	TraceHash         string  `json:"trace_hash"`
	PreviousShearHash string  `json:"previous_shear_hash"`
}

// HashLedger returns current_shear_hash for a record's pre-image.
func HashLedger(p LedgerHashPayload) (string, error) {
	return Sum(p)
}

// canonicalJSON marshals v into a deterministic byte sequence:
//   - object keys are sorted alphabetically at every depth
//   - HTML-unsafe characters are NOT escaped (<, >, & stay literal)
//   - no trailing newline
//
// Implementation: round-trip through encoding/json into a generic tree,
// then re-emit that tree with sorted keys.
func canonicalJSON(v interface{}) ([]byte, error) {
	var first bytes.Buffer
	enc := json.NewEncoder(&first)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	var generic interface{}
	dec := json.NewDecoder(bytes.NewReader(first.Bytes()))
	dec.UseNumber()
	if err := dec.Decode(&generic); err != nil {
		return nil, err
	}
	var out bytes.Buffer
	if err := writeCanonical(&out, generic); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func writeCanonical(out *bytes.Buffer, v interface{}) error {
	switch t := v.(type) {
	case nil:
		out.WriteString("null")
		return nil
	case bool:
		if t {
			out.WriteString("true")
		} else {
			out.WriteString("false")
		}
		return nil
	case json.Number:
		out.WriteString(t.String())
		return nil
	case string:
		return writeJSONString(out, t)
	case []interface{}:
		out.WriteByte('[')
		for i, item := range t {
			if i > 0 {
				out.WriteByte(',')
			}
			if err := writeCanonical(out, item); err != nil {
				return err
			}
		}
		out.WriteByte(']')
		return nil
	case map[string]interface{}:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		out.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				out.WriteByte(',')
			}
			if err := writeJSONString(out, k); err != nil {
				return err
			}
			out.WriteByte(':')
			if err := writeCanonical(out, t[k]); err != nil {
				return err
			}
		}
		out.WriteByte('}')
		return nil
	default:
		return fmt.Errorf("canonicalJSON: unsupported type %T", v)
	}
}

// writeJSONString writes a JSON string literal without HTML escaping.
func writeJSONString(out *bytes.Buffer, s string) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(s); err != nil {
		return err
	}
	// json.Encoder appends a trailing newline; strip it.
	b := buf.Bytes()
	if n := len(b); n > 0 && b[n-1] == '\n' {
		b = b[:n-1]
	}
	out.Write(b)
	return nil
}
