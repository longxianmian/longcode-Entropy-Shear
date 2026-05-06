package gateway

import (
	"encoding/json"
	"net/http"
	"time"

	"entropy-shear/internal/flagship/provider"
)

// Mux returns a configured HTTP mux with all gateway handlers attached.
// Reused by cmd/flagship-coder-gateway/main.go and by tests so the same
// handler code runs in production and under test.
//
// The provider is injected so tests can swap MockProvider for a recorder if
// needed without redefining HTTP wiring.
func Mux(prov provider.Provider) *http.ServeMux {
	s := &server{provider: prov, now: time.Now}
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/messages", s.messagesHandler)
	mux.HandleFunc("/v1/messages/count_tokens", s.countTokensHandler)
	mux.HandleFunc("/v1/models", s.modelsHandler)
	mux.HandleFunc("/health", s.healthHandler)
	return mux
}

type server struct {
	provider provider.Provider
	now      func() time.Time
}

func (s *server) messagesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "POST required")
		return
	}
	defer r.Body.Close()
	var req MessagesRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_json", err.Error())
		return
	}
	resp := runGovernance(req, s.provider, s.now())
	writeJSON(w, http.StatusOK, resp)
}

func (s *server) countTokensHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "POST required")
		return
	}
	defer r.Body.Close()
	var req CountTokensRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_json", err.Error())
		return
	}
	tokens := approxTokens(req.Messages, req.System)
	writeJSON(w, http.StatusOK, CountTokensResponse{InputTokens: tokens})
}

func (s *server) modelsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "GET required")
		return
	}
	writeJSON(w, http.StatusOK, ModelsResponse{
		Data: []ModelDescriptor{
			{ID: ModelID, Type: "model", DisplayName: "Flagship Coder Mock"},
		},
	})
}

func (s *server) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "GET required")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"module": "flagship-coder-gateway",
	})
}

func writeJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, map[string]string{
		"error":   code,
		"message": msg,
	})
}
