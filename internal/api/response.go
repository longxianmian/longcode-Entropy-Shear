package api

import (
	"encoding/json"
	"net/http"

	apperr "entropy-shear/internal/errors"
)

// errorBody is the wire shape returned for all error responses (§13).
type errorBody struct {
	Error  string `json:"error"`
	Detail string `json:"detail,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
}

func writeError(w http.ResponseWriter, err error) {
	if ae, ok := err.(*apperr.APIError); ok {
		writeJSON(w, ae.Status, errorBody{Error: ae.Code, Detail: ae.Detail})
		return
	}
	writeJSON(w, http.StatusInternalServerError, errorBody{
		Error:  apperr.CodeServiceUnavailable,
		Detail: err.Error(),
	})
}
