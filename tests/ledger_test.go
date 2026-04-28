package tests

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"entropy-shear/internal/ledger"
	"entropy-shear/internal/schema"
)

// fixedClock returns increasing timestamps so chained records get distinct
// timestamps even when called rapidly.
type fixedClock struct {
	now time.Time
}

func (c *fixedClock) tick() time.Time {
	c.now = c.now.Add(time.Second)
	return c.now
}

func newTestLedger(t *testing.T) (*ledger.Ledger, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "shear-chain.jsonl")
	clk := &fixedClock{now: time.Date(2026, 4, 28, 10, 0, 0, 0, time.UTC)}
	l, err := ledger.NewWithClock(path, clk.tick)
	if err != nil {
		t.Fatalf("ledger new: %v", err)
	}
	return l, path
}

func samplePolicy() schema.Policy {
	return schema.Policy{ID: "p1", Version: "1.0.0", DefaultEffect: schema.Hold, DefaultReason: "x"}
}

func TestAppendAndVerifyOK(t *testing.T) {
	l, path := newTestLedger(t)

	for i := 0; i < 3; i++ {
		_, err := l.Append(ledger.AppendInput{
			Policy: samplePolicy(),
			Facts:  schema.Facts{"i": float64(i)},
			Verdict: schema.Yes,
			Trace: []schema.TraceItem{{RuleID: "r", Evaluated: true, Matched: true, Detail: "ok"}},
		})
		if err != nil {
			t.Fatalf("append %d: %v", i, err)
		}
	}

	res, err := l.Verify()
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !res.OK {
		t.Fatalf("verify ok=false detail=%s broken_at=%v", res.Detail, res.BrokenAt)
	}
	if res.Total != 3 {
		t.Fatalf("total=%d want 3", res.Total)
	}
	if res.LatestShearID == "" || res.LatestHash == "" {
		t.Errorf("latest fields must be populated when chain non-empty")
	}

	// Sanity: file actually has 3 lines and shear_ids are zero-padded daily seqs.
	count := 0
	f, _ := os.Open(path)
	defer f.Close()
	sc := bufio.NewScanner(f)
	wantPrefixes := []string{
		"entropy-shear-20260428-000001",
		"entropy-shear-20260428-000002",
		"entropy-shear-20260428-000003",
	}
	for sc.Scan() {
		var rec schema.LedgerRecord
		if err := json.Unmarshal(sc.Bytes(), &rec); err != nil {
			t.Fatalf("line %d not JSON: %v", count, err)
		}
		if rec.ShearID != wantPrefixes[count] {
			t.Errorf("line %d shear_id=%s want %s", count, rec.ShearID, wantPrefixes[count])
		}
		count++
	}
	if count != 3 {
		t.Fatalf("file lines=%d want 3", count)
	}
}

func TestVerifyDetectsTamper(t *testing.T) {
	l, path := newTestLedger(t)
	for i := 0; i < 3; i++ {
		if _, err := l.Append(ledger.AppendInput{
			Policy: samplePolicy(),
			Facts:  schema.Facts{"i": float64(i)},
			Verdict: schema.Yes,
			Trace: []schema.TraceItem{{RuleID: "r", Evaluated: true, Matched: true, Detail: "ok"}},
		}); err != nil {
			t.Fatalf("append %d: %v", i, err)
		}
	}

	// Tamper: rewrite line 2 with a flipped verdict but keep its hash.
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(string(raw), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("file lines=%d", len(lines))
	}
	var mid schema.LedgerRecord
	if err := json.Unmarshal([]byte(lines[1]), &mid); err != nil {
		t.Fatal(err)
	}
	mid.Verdict = schema.No
	tampered, _ := json.Marshal(mid)
	lines[1] = string(tampered)
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := l.Verify()
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if res.OK {
		t.Fatalf("verify must fail after tamper")
	}
	if res.BrokenAt == nil || *res.BrokenAt != 2 {
		t.Errorf("broken_at=%v want 2", res.BrokenAt)
	}
}

func TestVerifyOnEmptyLedger(t *testing.T) {
	l, _ := newTestLedger(t)
	res, err := l.Verify()
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK {
		t.Errorf("empty ledger must verify ok")
	}
	if res.Total != 0 {
		t.Errorf("empty total=%d", res.Total)
	}
}

func TestGetByShearID(t *testing.T) {
	l, _ := newTestLedger(t)
	out, err := l.Append(ledger.AppendInput{
		Policy: samplePolicy(), Facts: schema.Facts{"a": 1},
		Verdict: schema.Yes,
		Trace:   []schema.TraceItem{{RuleID: "r", Evaluated: true, Matched: true, Detail: "ok"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	rec, err := l.Get(out.Record.ShearID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if rec.ShearID != out.Record.ShearID {
		t.Errorf("got %s want %s", rec.ShearID, out.Record.ShearID)
	}
}

func TestGetMissingReturnsNotExist(t *testing.T) {
	l, _ := newTestLedger(t)
	if _, err := l.Get("does-not-exist"); !os.IsNotExist(err) {
		t.Fatalf("expected ErrNotExist, got %v", err)
	}
}

func TestVerifyFileOffline(t *testing.T) {
	l, path := newTestLedger(t)
	for i := 0; i < 2; i++ {
		if _, err := l.Append(ledger.AppendInput{
			Policy:  samplePolicy(),
			Facts:   schema.Facts{"i": float64(i)},
			Verdict: schema.Yes,
			Trace:   []schema.TraceItem{{RuleID: "r", Evaluated: true, Matched: true, Detail: "ok"}},
		}); err != nil {
			t.Fatal(err)
		}
	}
	res, err := ledger.VerifyFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !res.OK || res.Total != 2 {
		t.Errorf("VerifyFile: ok=%v total=%d", res.OK, res.Total)
	}
}

func TestVerifyFileOnMissingPathReturnsEmpty(t *testing.T) {
	res, err := ledger.VerifyFile(filepath.Join(t.TempDir(), "does-not-exist.jsonl"))
	if err != nil {
		t.Fatalf("missing path should not error: %v", err)
	}
	if !res.OK || res.Total != 0 {
		t.Errorf("missing path should be ok=true total=0, got %+v", res)
	}
}

func TestRecoverPersistsLastHash(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "shear-chain.jsonl")
	clk := &fixedClock{now: time.Date(2026, 4, 28, 10, 0, 0, 0, time.UTC)}
	l1, err := ledger.NewWithClock(path, clk.tick)
	if err != nil {
		t.Fatal(err)
	}
	out1, _ := l1.Append(ledger.AppendInput{
		Policy: samplePolicy(), Facts: schema.Facts{"a": 1},
		Verdict: schema.Yes,
		Trace:   []schema.TraceItem{{RuleID: "r", Matched: true, Evaluated: true, Detail: "ok"}},
	})

	// Reopen — the new ledger must continue the chain from out1.
	l2, err := ledger.NewWithClock(path, clk.tick)
	if err != nil {
		t.Fatal(err)
	}
	out2, err := l2.Append(ledger.AppendInput{
		Policy: samplePolicy(), Facts: schema.Facts{"a": 2},
		Verdict: schema.Yes,
		Trace:   []schema.TraceItem{{RuleID: "r", Matched: true, Evaluated: true, Detail: "ok"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if out2.Record.PreviousShearHash != out1.Record.CurrentShearHash {
		t.Errorf("chain not continued across reopen: prev=%s want %s",
			out2.Record.PreviousShearHash, out1.Record.CurrentShearHash)
	}
	if out2.Record.ShearID != "entropy-shear-20260428-000002" {
		t.Errorf("seq did not resume: got %s", out2.Record.ShearID)
	}
}
