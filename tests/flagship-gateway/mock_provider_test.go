package flagship_gateway_test

import (
	"strings"
	"testing"

	"entropy-shear/internal/flagship/provider"
)

func TestMockProviderName(t *testing.T) {
	p := provider.NewMockProvider()
	if p.Name() != "mock" {
		t.Fatalf("name: got %q want mock", p.Name())
	}
}

func TestMockProviderDeterminism(t *testing.T) {
	p := provider.NewMockProvider()
	req := provider.GenerateRequest{
		Model:     "flagship-coder-mock-1",
		System:    "be helpful",
		Messages:  []provider.ProviderMessage{{Role: "user", Content: "hello"}},
		MaxTokens: 100,
	}
	a, err := p.Generate(req)
	if err != nil {
		t.Fatalf("first generate: %v", err)
	}
	b, err := p.Generate(req)
	if err != nil {
		t.Fatalf("second generate: %v", err)
	}
	if a.Text != b.Text {
		t.Fatalf("non-deterministic text: %q vs %q", a.Text, b.Text)
	}
	if a.TraceID != b.TraceID {
		t.Fatalf("non-deterministic trace_id: %s vs %s", a.TraceID, b.TraceID)
	}
}

func TestMockProviderTextShape(t *testing.T) {
	p := provider.NewMockProvider()
	resp, err := p.Generate(provider.GenerateRequest{
		Model:    "x",
		System:   "y",
		Messages: []provider.ProviderMessage{{Role: "user", Content: "z"}},
	})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if !strings.HasPrefix(resp.Text, "[mock-candidate sha:") {
		t.Fatalf("text should start with mock marker; got %q", resp.Text)
	}
	if resp.StopReason != "end_turn" {
		t.Fatalf("stop_reason: got %s want end_turn", resp.StopReason)
	}
	if !strings.HasPrefix(resp.TraceID, "mock-trace-") {
		t.Fatalf("trace_id should start with mock-trace-; got %s", resp.TraceID)
	}
}

func TestMockProviderDifferentRequestsDiffer(t *testing.T) {
	p := provider.NewMockProvider()
	a, _ := p.Generate(provider.GenerateRequest{Model: "m", System: "s1"})
	b, _ := p.Generate(provider.GenerateRequest{Model: "m", System: "s2"})
	if a.Text == b.Text {
		t.Fatalf("different requests should yield different text; got identical %q", a.Text)
	}
}
