// Package provider abstracts the LLM provider behind a small interface so
// the gateway can stay pluggable. P0 ships only MockProvider — no real LLM
// SDK is imported and no API key is read.
package provider

// ProviderMessage is the canonical message shape passed to a provider.
// The gateway lowers Anthropic-shaped messages into this string-content form
// before calling Generate, keeping providers free of HTTP-surface types.
type ProviderMessage struct {
	Role    string
	Content string
}

// GenerateRequest is the input to Provider.Generate. P0 carries only the
// fields the mock provider needs; future providers may extend via additional
// fields without breaking the interface.
type GenerateRequest struct {
	Model     string
	System    string
	Messages  []ProviderMessage
	MaxTokens int
}

// GenerateResponse is the return value of Provider.Generate.
type GenerateResponse struct {
	Text       string
	StopReason string
	TraceID    string
}

// Provider is the LLM provider abstraction.
type Provider interface {
	Name() string
	Generate(req GenerateRequest) (GenerateResponse, error)
}
