package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"entropy-shear/internal/api"
	"entropy-shear/internal/ledger"
)

const version = "v1.0.0"

func main() {
	addr := getenv("ENTROPY_SHEAR_ADDR", ":8080")
	ledgerPath := getenv("ENTROPY_SHEAR_LEDGER", "ledger/shear-chain.jsonl")

	l, err := ledger.New(ledgerPath)
	if err != nil {
		log.Fatalf("entropy-shear: ledger init failed: %v", err)
	}

	srv := &api.Server{Ledger: l, Version: version}
	server := &http.Server{
		Addr:              addr,
		Handler:           srv.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	idle := make(chan struct{})
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("entropy-shear: shutdown: %v", err)
		}
		close(idle)
	}()

	log.Printf("entropy-shear %s listening on %s, ledger=%s", version, addr, ledgerPath)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("entropy-shear: serve failed: %v", err)
	}
	<-idle
	log.Printf("entropy-shear: stopped")
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
