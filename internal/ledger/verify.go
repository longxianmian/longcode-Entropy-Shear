package ledger

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"entropy-shear/internal/schema"
	"entropy-shear/internal/signature"
)

// VerifyResult is the response shape of GET /ledger/verify (§6.3).
type VerifyResult struct {
	OK            bool    `json:"ok"`
	Total         int     `json:"total"`
	BrokenAt      *int    `json:"broken_at"`
	LatestShearID string  `json:"latest_shear_id,omitempty"`
	LatestHash    string  `json:"latest_hash,omitempty"`
	Detail        string  `json:"detail,omitempty"`
}

// Verify walks the JSONL ledger from genesis and re-derives every chained
// hash. broken_at is 1-indexed (the first record is line 1), or nil when
// the chain is intact.
func (l *Ledger) Verify() (VerifyResult, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return verifyReader(l.path)
}

// VerifyFile walks the chain at path without instantiating a stateful
// Ledger. Suitable for offline / one-shot tooling — does not create
// directories, does not write, does not hold any internal state.
func VerifyFile(path string) (VerifyResult, error) {
	return verifyReader(path)
}

func verifyReader(path string) (VerifyResult, error) {
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return VerifyResult{OK: true, Total: 0}, nil
		}
		return VerifyResult{}, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)

	prev := schema.GenesisHash
	total := 0
	var lastID, lastHash string

	for scanner.Scan() {
		total++
		var rec schema.LedgerRecord
		if err := json.Unmarshal(scanner.Bytes(), &rec); err != nil {
			at := total
			return VerifyResult{
				OK: false, Total: total, BrokenAt: &at,
				Detail: fmt.Sprintf("line %d: invalid JSON: %v", total, err),
			}, nil
		}
		if rec.PreviousShearHash != prev {
			at := total
			return VerifyResult{
				OK: false, Total: total, BrokenAt: &at,
				Detail: fmt.Sprintf("line %d: previous_shear_hash mismatch (expected %s, got %s)",
					total, prev, rec.PreviousShearHash),
			}, nil
		}
		recomputed, err := signature.HashLedger(signature.LedgerHashPayload{
			ShearID:           rec.ShearID,
			Timestamp:         rec.Timestamp,
			PolicyID:          rec.PolicyID,
			PolicyVersion:     rec.PolicyVersion,
			InputHash:         rec.InputHash,
			Verdict:           string(rec.Verdict),
			AppliedRuleID:     rec.AppliedRuleID,
			TraceHash:         rec.TraceHash,
			PreviousShearHash: rec.PreviousShearHash,
		})
		if err != nil {
			return VerifyResult{}, err
		}
		if recomputed != rec.CurrentShearHash {
			at := total
			return VerifyResult{
				OK: false, Total: total, BrokenAt: &at,
				Detail: fmt.Sprintf("line %d: current_shear_hash mismatch (expected %s, got %s)",
					total, recomputed, rec.CurrentShearHash),
			}, nil
		}
		prev = rec.CurrentShearHash
		lastID = rec.ShearID
		lastHash = rec.CurrentShearHash
	}
	if err := scanner.Err(); err != nil {
		return VerifyResult{}, err
	}
	return VerifyResult{
		OK:            true,
		Total:         total,
		LatestShearID: lastID,
		LatestHash:    lastHash,
	}, nil
}
