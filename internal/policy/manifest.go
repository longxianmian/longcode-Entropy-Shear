package policy

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// ManifestEntry is one row of policies/manifest.json. It pins a policy
// file path to its identifier, version, content hash, and the scenario
// it applies to. The hash field is what makes the manifest tamper-evident.
type ManifestEntry struct {
	Path          string `json:"path"`
	PolicyID      string `json:"policy_id"`
	PolicyVersion string `json:"policy_version"`
	Hash          string `json:"hash"`
	Scenario      string `json:"scenario"`
	Maintainer    string `json:"maintainer"`
	CreatedAt     string `json:"created_at"`
}

// Manifest is policies/manifest.json.
type Manifest struct {
	ManifestVersion string          `json:"manifest_version"`
	GeneratedAt     string          `json:"generated_at"`
	Policies        []ManifestEntry `json:"policies"`
}

// LoadManifest reads policies/manifest.json.
func LoadManifest(path string) (Manifest, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("read manifest: %w", err)
	}
	var m Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return Manifest{}, fmt.Errorf("decode manifest: %w", err)
	}
	return m, nil
}

// VerifyManifest re-derives every entry's hash from the corresponding
// policy file and reports the diff. paths are resolved relative to root.
func VerifyManifest(root string, m Manifest) (ManifestVerifyResult, error) {
	res := ManifestVerifyResult{Total: len(m.Policies)}
	for _, e := range m.Policies {
		full := filepath.Join(root, e.Path)
		if _, err := os.Stat(full); err != nil {
			res.Missing = append(res.Missing, e.Path)
			continue
		}
		got, err := HashFile(full)
		if err != nil {
			return res, fmt.Errorf("hash %s: %w", e.Path, err)
		}
		if got != e.Hash {
			res.Mismatched = append(res.Mismatched, ManifestMismatch{
				Path: e.Path, Expected: e.Hash, Actual: got,
			})
		}
	}
	sort.Strings(res.Missing)
	res.OK = len(res.Missing) == 0 && len(res.Mismatched) == 0
	return res, nil
}

// ManifestMismatch identifies one entry whose hash drifted from disk.
type ManifestMismatch struct {
	Path     string `json:"path"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
}

// ManifestVerifyResult is the test-friendly summary returned by VerifyManifest.
type ManifestVerifyResult struct {
	OK         bool               `json:"ok"`
	Total      int                `json:"total"`
	Missing    []string           `json:"missing,omitempty"`
	Mismatched []ManifestMismatch `json:"mismatched,omitempty"`
}
