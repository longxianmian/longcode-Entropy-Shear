package errors

// Error codes used in JSON error responses, per §13 of the spec.
const (
	CodeInvalidJSON            = "invalid_json"
	CodePolicySchemaViolation  = "policy_schema_violation"
	CodeFactsSchemaViolation   = "facts_schema_violation"
	CodeUnsupportedOperator    = "unsupported_operator"
	CodeLedgerUnavailable      = "ledger_unavailable"
	CodeServiceUnavailable     = "service_unavailable"
	CodeNotFound               = "not_found"
)

// APIError represents a structured error response.
type APIError struct {
	Status int
	Code   string
	Detail string
}

func (e *APIError) Error() string {
	return e.Code + ": " + e.Detail
}

func New(status int, code, detail string) *APIError {
	return &APIError{Status: status, Code: code, Detail: detail}
}
