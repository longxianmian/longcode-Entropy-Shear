package gateway

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"entropy-shear/internal/flagship/coder"
	"entropy-shear/internal/flagship/provider"
	"entropy-shear/internal/flagship/reasoner"
)

// ModelID is the single mock model id this P0 gateway advertises (GD-9).
const ModelID = "flagship-coder-mock-1"

// runGovernance executes the full pre-governance → generate → post-governance
// pipeline for one MessagesRequest and returns the fully-populated response.
//
// The HOLD/NO branches at pre-governance short-circuit the provider call so
// no candidate is wasted on a request the kernel has already rejected.
func runGovernance(req MessagesRequest, prov provider.Provider, now time.Time) MessagesResponse {
	seedID := requestSeedID(req)
	requestID := "gw-req-" + seedID

	preFields := lowerToPreFields(req)
	preInput := coder.BuildPreGovernanceInput(seedID, preFields)
	preOut := reasoner.Reason(preInput)

	if preOut.Verdict != reasoner.VerdictYes {
		text, stopReason := coder.BuildAssistantContent(string(preOut.Verdict), "",
			preOut.AlignmentTasks, preOut.RejectInstruction, "pre")
		audit := coder.NewGatewayAuditRecord(requestID, prov.Name(), "",
			string(preOut.Verdict), preOut.AuditRecord, nil, now)
		return buildResponse(requestID, req, text, stopReason, string(preOut.Verdict), &audit)
	}

	genResp, _ := prov.Generate(provider.GenerateRequest{
		Model:     req.Model,
		System:    req.System,
		Messages:  toProviderMessages(req.Messages),
		MaxTokens: req.MaxTokens,
	})

	postInput := coder.BuildPostGovernanceInput(seedID, preInput, genResp.Text)
	postOut := reasoner.Reason(postInput)

	candidate := genResp.Text
	if postOut.Verdict != reasoner.VerdictYes {
		// On post-governance HOLD/NO the candidate text is suppressed; the
		// content surface only carries the governance summary so callers
		// don't accidentally execute a refused candidate.
		candidate = ""
	}
	text, stopReason := coder.BuildAssistantContent(string(postOut.Verdict), candidate,
		postOut.AlignmentTasks, postOut.RejectInstruction, "post")
	postAudit := postOut.AuditRecord
	audit := coder.NewGatewayAuditRecord(requestID, prov.Name(), genResp.TraceID,
		string(postOut.Verdict), preOut.AuditRecord, &postAudit, now)
	return buildResponse(requestID, req, text, stopReason, string(postOut.Verdict), &audit)
}

func buildResponse(requestID string, req MessagesRequest, text, stopReason, verdict string, audit *coder.GatewayAuditRecord) MessagesResponse {
	model := req.Model
	if model == "" {
		model = ModelID
	}
	return MessagesResponse{
		ID:           "msg_" + requestID,
		Type:         "message",
		Role:         "assistant",
		Content:      []ContentBlock{{Type: "text", Text: text}},
		Model:        model,
		StopReason:   stopReason,
		StopSequence: nil,
		Usage: Usage{
			InputTokens:  approxTokens(req.Messages, req.System),
			OutputTokens: approxTokensForText(text),
		},
		Verdict:      verdict,
		GatewayAudit: audit,
	}
}

func lowerToPreFields(req MessagesRequest) coder.PreGovernanceFields {
	var userMsgs, asstMsgs []string
	for _, m := range req.Messages {
		text := joinTextBlocks(m.Content)
		switch m.Role {
		case "user":
			userMsgs = append(userMsgs, text)
		case "assistant":
			asstMsgs = append(asstMsgs, text)
		}
	}
	return coder.PreGovernanceFields{
		System:            req.System,
		UserMessages:      userMsgs,
		AssistantMessages: asstMsgs,
		Metadata:          req.Metadata,
	}
}

func joinTextBlocks(blocks []ContentBlock) string {
	parts := make([]string, 0, len(blocks))
	for _, b := range blocks {
		if b.Type == "text" {
			parts = append(parts, b.Text)
		}
	}
	return strings.Join(parts, "\n")
}

func toProviderMessages(msgs []Message) []provider.ProviderMessage {
	out := make([]provider.ProviderMessage, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, provider.ProviderMessage{
			Role:    m.Role,
			Content: joinTextBlocks(m.Content),
		})
	}
	return out
}

func requestSeedID(req MessagesRequest) string {
	body, _ := json.Marshal(req)
	h := sha256.Sum256(body)
	return hex.EncodeToString(h[:])[:16]
}

// approxTokens implements the GD-6 placeholder: ceil(JSON byte length / 4).
func approxTokens(msgs []Message, system string) int {
	body, _ := json.Marshal(struct {
		System   string    `json:"system"`
		Messages []Message `json:"messages"`
	}{System: system, Messages: msgs})
	return ceilDiv(len(body), 4)
}

func approxTokensForText(text string) int {
	return ceilDiv(len(text), 4)
}

func ceilDiv(n, d int) int {
	if n <= 0 {
		return 0
	}
	return (n + d - 1) / d
}
