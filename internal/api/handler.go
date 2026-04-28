package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"

	"entropy-shear/internal/engine"
	apperr "entropy-shear/internal/errors"
	"entropy-shear/internal/ledger"
	"entropy-shear/internal/schema"
)

// Server bundles the dependencies needed by the HTTP handlers.
type Server struct {
	Ledger  *ledger.Ledger
	Version string
}

// Routes returns an http.Handler with all routes mounted (§6).
// Uses Go 1.22 stdlib pattern routing — no external router dependency.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("POST /shear", s.shear)
	mux.HandleFunc("GET /ledger/verify", s.ledgerVerify)
	mux.HandleFunc("GET /ledger/{shear_id}", s.ledgerGet)
	return mux
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":      true,
		"service": "entropy-shear",
		"version": s.Version,
	})
}

func (s *Server) shear(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	dec.UseNumber()
	var req schema.ShearRequest
	if err := dec.Decode(&req); err != nil {
		writeError(w, apperr.New(http.StatusBadRequest, apperr.CodeInvalidJSON, err.Error()))
		return
	}

	if err := schema.ValidatePolicy(&req.Policy); err != nil {
		writeError(w, err)
		return
	}
	if err := schema.ValidateFacts(req.Facts); err != nil {
		writeError(w, err)
		return
	}

	decision := engine.Shear(req.Policy, req.Facts)

	out, err := s.Ledger.Append(ledger.AppendInput{
		Policy:        req.Policy,
		Facts:         req.Facts,
		Verdict:       decision.Verdict,
		AppliedRuleID: decision.AppliedRuleID,
		Trace:         decision.Trace,
	})
	if err != nil {
		writeError(w, apperr.New(http.StatusServiceUnavailable,
			apperr.CodeLedgerUnavailable, err.Error()))
		return
	}

	resp := schema.ShearResult{
		Verdict:       decision.Verdict,
		AppliedRuleID: decision.AppliedRuleID,
		Route:         decision.Route,
		Reason:        decision.Reason,
		Trace:         decision.Trace,
		Signature:     out.Signature,
		ShearID:       out.Record.ShearID,
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) ledgerGet(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("shear_id"))
	if id == "" {
		writeError(w, apperr.New(http.StatusBadRequest, apperr.CodeInvalidJSON, "shear_id required"))
		return
	}
	rec, err := s.Ledger.Get(id)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			writeError(w, apperr.New(http.StatusNotFound, apperr.CodeNotFound,
				"ledger record not found: "+id))
			return
		}
		writeError(w, apperr.New(http.StatusServiceUnavailable,
			apperr.CodeLedgerUnavailable, err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, rec)
}

func (s *Server) ledgerVerify(w http.ResponseWriter, _ *http.Request) {
	res, err := s.Ledger.Verify()
	if err != nil {
		writeError(w, apperr.New(http.StatusServiceUnavailable,
			apperr.CodeLedgerUnavailable, err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, res)
}
