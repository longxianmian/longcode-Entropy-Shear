package schema

// Verdict is one of the three legal verdicts.
type Verdict string

const (
	Yes  Verdict = "Yes"
	No   Verdict = "No"
	Hold Verdict = "Hold"
)

// Operator is the set of supported condition operators (P0: 8 operators, see §10).
type Operator string

const (
	OpEq       Operator = "=="
	OpNeq      Operator = "!="
	OpGt       Operator = ">"
	OpLt       Operator = "<"
	OpGte      Operator = ">="
	OpLte      Operator = "<="
	OpIn       Operator = "in"
	OpContains Operator = "contains"
)

// SupportedOperators lists all operators recognized by the engine.
var SupportedOperators = map[Operator]struct{}{
	OpEq: {}, OpNeq: {}, OpGt: {}, OpLt: {}, OpGte: {}, OpLte: {},
	OpIn: {}, OpContains: {},
}

// Condition is a single rule's match condition (§7.3).
type Condition struct {
	Field    string      `json:"field"`
	Operator Operator    `json:"operator"`
	Value    interface{} `json:"value"`
}

// Rule is a single policy rule (§7.2). Effect is constrained to Yes/No;
// Hold is only emitted via default_effect or via runtime escape paths.
type Rule struct {
	ID        string    `json:"id"`
	Priority  int       `json:"priority"`
	Condition Condition `json:"condition"`
	Effect    Verdict   `json:"effect"`
	Route     string    `json:"route,omitempty"`
	Reason    string    `json:"reason"`
}

// Policy is the full ruleset submitted by the caller (§7.1).
type Policy struct {
	ID            string  `json:"id"`
	Version       string  `json:"version"`
	Rules         []Rule  `json:"rules"`
	DefaultEffect Verdict `json:"default_effect"`
	DefaultReason string  `json:"default_reason"`
}

// Facts is the structured business state (§7.4) — arbitrary JSON object.
type Facts map[string]interface{}

// ShearRequest is the wire shape of POST /shear.
type ShearRequest struct {
	Policy Policy `json:"policy"`
	Facts  Facts  `json:"facts"`
}

// TraceItem captures a single rule's evaluation outcome (§7.5).
type TraceItem struct {
	RuleID    string `json:"rule_id"`
	Evaluated bool   `json:"evaluated"`
	Matched   bool   `json:"matched"`
	Detail    string `json:"detail"`
}

// ShearResult is the wire shape of the response (§7.6).
type ShearResult struct {
	Verdict       Verdict     `json:"verdict"`
	AppliedRuleID *string     `json:"applied_rule_id"`
	Route         string      `json:"route,omitempty"`
	Reason        string      `json:"reason"`
	Trace         []TraceItem `json:"trace"`
	Signature     string      `json:"signature"`
	ShearID       string      `json:"shear_id"`
}

// LedgerRecord is the per-line shape of the JSONL ledger (§7.7).
type LedgerRecord struct {
	ShearID           string  `json:"shear_id"`
	Timestamp         string  `json:"timestamp"`
	PolicyID          string  `json:"policy_id"`
	PolicyVersion     string  `json:"policy_version"`
	InputHash         string  `json:"input_hash"`
	Verdict           Verdict `json:"verdict"`
	AppliedRuleID     *string `json:"applied_rule_id"`
	TraceHash         string  `json:"trace_hash"`
	PreviousShearHash string  `json:"previous_shear_hash"`
	CurrentShearHash  string  `json:"current_shear_hash"`
}

// GenesisHash is the seed value for previous_shear_hash on the first record.
const GenesisHash = "sha256:genesis"
