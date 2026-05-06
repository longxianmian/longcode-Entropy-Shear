// Command flagship-coder-gateway exposes the flagship coder gateway P0 over
// HTTP. It is independent from the Flagship reasoner service (cmd/flagship-server,
// default :9090) and the Core server (cmd/server, default :8080).
//
// Endpoints:
//
//	GET  /health
//	GET  /v1/models
//	POST /v1/messages/count_tokens
//	POST /v1/messages
//
// Configuration:
//
//	ENTROPY_SHEAR_FLAGSHIP_GATEWAY_ADDR — listen address; default ":9091".
//
// This binary deliberately uses only the in-tree MockProvider; it does not
// import any real LLM SDK and reads no API key.
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

	"entropy-shear/internal/flagship/gateway"
	"entropy-shear/internal/flagship/provider"
)

const defaultAddr = ":9091"

func main() {
	addr := os.Getenv("ENTROPY_SHEAR_FLAGSHIP_GATEWAY_ADDR")
	if addr == "" {
		addr = defaultAddr
	}

	mux := gateway.Mux(provider.NewMockProvider())
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("flagship-coder-gateway listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("flagship-coder-gateway: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("flagship-coder-gateway: shutdown error: %v", err)
	}
}
