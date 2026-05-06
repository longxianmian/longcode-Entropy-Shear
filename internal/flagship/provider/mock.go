package provider

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// MockProviderName is the value MockProvider.Name() returns.
const MockProviderName = "mock"

// MockProvider returns a deterministic placeholder candidate per GD-7. It
// performs no real LLM call, reads no API key, and makes no network IO.
type MockProvider struct{}

// NewMockProvider constructs a MockProvider. Kept as a constructor so future
// providers can follow the same factory pattern without re-shaping callers.
func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

// Name returns "mock".
func (m *MockProvider) Name() string { return MockProviderName }

// Generate hashes the canonical JSON encoding of the request and returns a
// fixed-template text containing the first 8 hex chars of that hash. The
// same request therefore always produces the same response (testable).
func (m *MockProvider) Generate(req GenerateRequest) (GenerateResponse, error) {
	body, _ := json.Marshal(req)
	short := first8Hex(body)
	text := "[mock-candidate sha:" + short + "] entropy-shear flagship coder gateway P0 placeholder; not from a real LLM."
	return GenerateResponse{
		Text:       text,
		StopReason: "end_turn",
		TraceID:    "mock-trace-" + short,
	}, nil
}

func first8Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])[:8]
}
