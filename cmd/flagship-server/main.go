// Command flagship-server exposes the flagship P0 reasoner over HTTP.
//
// Endpoints:
//
//	POST /flagship/reason  — run the three-state reasoner
//	GET  /health           — liveness probe (flagship-only, does not touch
//	                         Core /health)
//
// Configuration:
//
//	ENTROPY_SHEAR_FLAGSHIP_ADDR — listen address; default ":9090".
//
// This binary deliberately does not import any Core engine packages and does
// not call any external LLM. It is a closed deterministic reasoning kernel.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"entropy-shear/internal/flagship/reasoner"
)

const defaultAddr = ":9090"

func main() {
	addr := os.Getenv("ENTROPY_SHEAR_FLAGSHIP_ADDR")
	if addr == "" {
		addr = defaultAddr
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/flagship/reason", reasonHandler)
	mux.HandleFunc("/health", healthHandler)

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("flagship-server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("flagship-server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("flagship-server: shutdown error: %v", err)
	}
}

func reasonHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "POST required")
		return
	}
	defer r.Body.Close()

	var in reasoner.Input
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "bad_json", err.Error())
		return
	}

	out := reasoner.Reason(in)
	writeJSON(w, http.StatusOK, out)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "GET required")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"module": "flagship",
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
