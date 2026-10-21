package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bikerental/internal/handler"
	"bikerental/internal/store"
)

func main() {
	demo := flag.Bool("demo", false, "Synthetische Beispieldaten laden")
	flag.Parse()
	if flag.NArg() != 0 {
		log.Fatal("Unbekannte Argumente. Verwendung: go run ./cmd/server [-demo]")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, *demo); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, demo bool) error {
	app, err := handler.New(store.New(demo))
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		return fmt.Errorf("Server starten: %w", err)
	}
	server := &http.Server{
		Handler: app, ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second,
		IdleTimeout: 60 * time.Second,
	}
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	log.Printf("BikeRental Phase 1: http://127.0.0.1:8080 (Demomodus: %t). Beenden mit Ctrl+C.", demo)
	select {
	case err := <-done:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("HTTP-Server: %w", err)
		}
		return nil
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			_ = server.Close()
			return fmt.Errorf("Server beenden: %w", err)
		}
		return nil
	}
}
