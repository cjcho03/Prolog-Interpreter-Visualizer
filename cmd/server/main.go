package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	prolog "prolog-interpreter"
)

func predicate(name string, args ...prolog.Term) prolog.Predicate {
	return prolog.Predicate{
		Name: name,
		Args: args,
	}
}

func main() {
	engine := prolog.Engine{
		Clauses: []prolog.Clause{
			prolog.Fact(
				predicate(
					"parent",
					prolog.Atom("alice"),
					prolog.Atom("bob"),
				),
			),

			prolog.Fact(
				predicate(
					"parent",
					prolog.Atom("alice"),
					prolog.Atom("carol"),
				),
			),

			prolog.Fact(
				predicate(
					"parent",
					prolog.Atom("bob"),
					prolog.Atom("diana"),
				),
			),

			prolog.Fact(
				predicate(
					"parent",
					prolog.Atom("carol"),
					prolog.Atom("eli"),
				),
			),

			prolog.Rule(
				predicate(
					"grandparent",
					prolog.Var("X"),
					prolog.Var("Z"),
				),
				predicate(
					"parent",
					prolog.Var("X"),
					prolog.Var("Y"),
				),
				predicate(
					"parent",
					prolog.Var("Y"),
					prolog.Var("Z"),
				),
			),
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", prolog.HealthHandler())
	mux.HandleFunc("/api/demo", prolog.DemoHandler(engine))
	mux.HandleFunc("/api/solve", prolog.SolveHandler())

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}

	allowedOrigins := splitCommaSeparated(os.Getenv("ALLOWED_ORIGINS"))

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           prolog.CORSMiddleware(allowedOrigins, mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.Printf("Go API listening on :%s", port)
		errCh <- server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed: %v", err)
		}
	case <-ctx.Done():
		log.Println("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}

func splitCommaSeparated(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}

	return result
}
