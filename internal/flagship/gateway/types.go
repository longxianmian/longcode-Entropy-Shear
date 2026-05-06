// Package gateway exposes the Anthropic-Messages-compatible HTTP surface for
// the flagship coder gateway P0. Types in this file are the wire shapes
// (request / response bodies). Handlers live in server.go and the
// pre→generate→post pipeline lives in governance.go.
package gateway

import (
	"entropy-shear/internal/flagship/coder"
)

// ContentBlock is a single block inside a Message's content array.
// P0 supports only Type == "text" per GD-12.
type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// Message is a single chat-style message inside MessagesRequest.Messages.
type Message struct {
	Role    string         `json:"role"`
	Content []ContentBlock `json:"content"`
}

// MessagesRequest is the POST /v1/messages body, P0 minimal subset (GD-1).
type MessagesRequest struct {
	Model     string                 `json:"model"`
	MaxTokens int                    `json:"max_tokens"`
	Messages  []Message              `json:"messages"`
	System    string                 `json:"system,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Usage carries token-counting placeholders per GD-2.
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// MessagesResponse is the POST /v1/messages response body. Beyond the
// Anthropic-compatible fields, it also surfaces a custom Verdict and
// GatewayAudit so callers can introspect the governance result without an
// extra round trip.
type MessagesResponse struct {
	ID           string                    `json:"id"`
	Type         string                    `json:"type"`
	Role         string                    `json:"role"`
	Content      []ContentBlock            `json:"content"`
	Model        string                    `json:"model"`
	StopReason   string                    `json:"stop_reason"`
	StopSequence *string                   `json:"stop_sequence"`
	Usage        Usage                     `json:"usage"`
	Verdict      string                    `json:"verdict"`
	GatewayAudit *coder.GatewayAuditRecord `json:"gateway_audit,omitempty"`
}

// CountTokensRequest mirrors MessagesRequest minus max_tokens (token counting
// is independent of the generation budget).
type CountTokensRequest struct {
	Model    string                 `json:"model"`
	Messages []Message              `json:"messages"`
	System   string                 `json:"system,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// CountTokensResponse is the POST /v1/messages/count_tokens body.
type CountTokensResponse struct {
	InputTokens int `json:"input_tokens"`
}

// ModelDescriptor is one entry of GET /v1/models data array.
type ModelDescriptor struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	DisplayName string `json:"display_name"`
}

// ModelsResponse is the GET /v1/models response body.
type ModelsResponse struct {
	Data []ModelDescriptor `json:"data"`
}
