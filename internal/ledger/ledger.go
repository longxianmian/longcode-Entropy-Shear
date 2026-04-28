package ledger

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"entropy-shear/internal/schema"
	"entropy-shear/internal/signature"
)

// Ledger appends shear records to a JSONL file under a single mutex.
// On startup it scans the file to recover the last hash and the per-day
// sequence counter so shear_ids stay monotonic across restarts.
type Ledger struct {
	mu       sync.Mutex
	path     string
	lastHash string
	dayKey   string
	daySeq   int
	clock    func() time.Time
}

// AppendInput is the pre-hash data the API layer hands to Append. It owns
// the verdict and trace; the ledger owns timestamp, shear_id, hashing,
// and chain linkage.
type AppendInput struct {
	Policy        schema.Policy
	Facts         schema.Facts
	Verdict       schema.Verdict
	AppliedRuleID *string
	Trace         []schema.TraceItem
}

// AppendOutput contains everything the API needs to build the response.
type AppendOutput struct {
	Record    schema.LedgerRecord
	Signature string
}

// New opens (or creates) a ledger at path and recovers chain state.
func New(path string) (*Ledger, error) {
	return NewWithClock(path, time.Now)
}

// NewWithClock is like New but allows test injection of a clock.
func NewWithClock(path string, clock func() time.Time) (*Ledger, error) {
	if path == "" {
		return nil, errors.New("ledger path required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("ledger mkdir: %w", err)
	}
	l := &Ledger{
		path:     path,
		lastHash: schema.GenesisHash,
		clock:    clock,
	}
	if err := l.recover(); err != nil {
		return nil, err
	}
	return l, nil
}

func (l *Ledger) recover() error {
	f, err := os.Open(l.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("ledger open: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	var rec schema.LedgerRecord
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(strings.TrimSpace(string(line))) == 0 {
			continue
		}
		rec = schema.LedgerRecord{}
		if err := json.Unmarshal(line, &rec); err != nil {
			return fmt.Errorf("ledger corrupt at line: %w", err)
		}
		l.lastHash = rec.CurrentShearHash
		if dk, seq, ok := parseShearID(rec.ShearID); ok {
			if dk == l.dayKey {
				if seq > l.daySeq {
					l.daySeq = seq
				}
			} else {
				l.dayKey = dk
				l.daySeq = seq
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("ledger scan: %w", err)
	}
	return nil
}

// Append builds the next chained record, writes it, and returns the
// record + response signature.
func (l *Ledger) Append(in AppendInput) (AppendOutput, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.clock().UTC()
	timestamp := now.Format(time.RFC3339)
	dayKey := now.Format("20060102")

	if dayKey != l.dayKey {
		l.dayKey = dayKey
		l.daySeq = 0
	}
	l.daySeq++
	shearID := fmt.Sprintf("entropy-shear-%s-%06d", dayKey, l.daySeq)

	inputHash, err := signature.HashFacts(in.Facts)
	if err != nil {
		return AppendOutput{}, fmt.Errorf("hash facts: %w", err)
	}
	traceHash, err := signature.HashTrace(in.Trace)
	if err != nil {
		return AppendOutput{}, fmt.Errorf("hash trace: %w", err)
	}

	sig, err := signature.SignResult(signature.ResultSignaturePayload{
		PolicyID:      in.Policy.ID,
		PolicyVersion: in.Policy.Version,
		InputHash:     inputHash,
		Verdict:       string(in.Verdict),
		AppliedRuleID: in.AppliedRuleID,
		TraceHash:     traceHash,
		PreviousHash:  l.lastHash,
	})
	if err != nil {
		return AppendOutput{}, fmt.Errorf("sign result: %w", err)
	}

	currentHash, err := signature.HashLedger(signature.LedgerHashPayload{
		ShearID:           shearID,
		Timestamp:         timestamp,
		PolicyID:          in.Policy.ID,
		PolicyVersion:     in.Policy.Version,
		InputHash:         inputHash,
		Verdict:           string(in.Verdict),
		AppliedRuleID:     in.AppliedRuleID,
		TraceHash:         traceHash,
		PreviousShearHash: l.lastHash,
	})
	if err != nil {
		return AppendOutput{}, fmt.Errorf("hash ledger: %w", err)
	}

	rec := schema.LedgerRecord{
		ShearID:           shearID,
		Timestamp:         timestamp,
		PolicyID:          in.Policy.ID,
		PolicyVersion:     in.Policy.Version,
		InputHash:         inputHash,
		Verdict:           in.Verdict,
		AppliedRuleID:     in.AppliedRuleID,
		TraceHash:         traceHash,
		PreviousShearHash: l.lastHash,
		CurrentShearHash:  currentHash,
	}

	if err := l.appendLine(rec); err != nil {
		// Roll back the in-memory counter so the next call doesn't skip.
		l.daySeq--
		return AppendOutput{}, err
	}

	l.lastHash = currentHash
	return AppendOutput{Record: rec, Signature: sig}, nil
}

func (l *Ledger) appendLine(rec schema.LedgerRecord) error {
	f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("ledger open for append: %w", err)
	}
	defer f.Close()
	b, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("ledger marshal: %w", err)
	}
	b = append(b, '\n')
	if _, err := f.Write(b); err != nil {
		return fmt.Errorf("ledger write: %w", err)
	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("ledger sync: %w", err)
	}
	return nil
}

// Get returns the record matching shear_id, or os.ErrNotExist.
func (l *Ledger) Get(shearID string) (schema.LedgerRecord, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	f, err := os.Open(l.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return schema.LedgerRecord{}, os.ErrNotExist
		}
		return schema.LedgerRecord{}, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		var rec schema.LedgerRecord
		if err := json.Unmarshal(scanner.Bytes(), &rec); err != nil {
			continue
		}
		if rec.ShearID == shearID {
			return rec, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return schema.LedgerRecord{}, err
	}
	return schema.LedgerRecord{}, os.ErrNotExist
}

// Path returns the file path the ledger writes to.
func (l *Ledger) Path() string { return l.path }

// parseShearID parses "entropy-shear-YYYYMMDD-NNNNNN".
func parseShearID(id string) (dayKey string, seq int, ok bool) {
	parts := strings.Split(id, "-")
	if len(parts) != 4 || parts[0] != "entropy" || parts[1] != "shear" {
		return "", 0, false
	}
	if len(parts[2]) != 8 {
		return "", 0, false
	}
	n, err := strconv.Atoi(parts[3])
	if err != nil {
		return "", 0, false
	}
	return parts[2], n, true
}
