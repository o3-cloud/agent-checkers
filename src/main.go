package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/stackable-specs/agent-checkers/internal/app/session"
	"github.com/stackable-specs/agent-checkers/internal/app/store"
	"github.com/stackable-specs/agent-checkers/src/api"
	"github.com/stackable-specs/agent-checkers/src/config"
	"github.com/stackable-specs/agent-checkers/src/mcp"
)

func main() {
	healthcheck := flag.Bool("healthcheck", false, "check whether the service is healthy")
	mcpMode := flag.Bool("mcp", false, "run the MCP stdio server")
	flag.Parse()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if *healthcheck {
		runHealthcheck(port)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	gameStore := config.NewStore(config.LoadStoreConfig())
	if *mcpMode {
		if err := mcp.NewServer(gameStore).Run(ctx, os.Stdin, os.Stdout); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	// Start background cleanup of completed/drawn games.
	cleanupTTL := envDuration("CLEANUP_TTL", 24*time.Hour)
	cleanupInterval := envDuration("CLEANUP_INTERVAL", time.Hour)
	go runCleanup(ctx, gameStore, cleanupTTL, cleanupInterval)

	sessionManager := session.NewManager(24 * time.Hour)
	server := api.NewServer(api.Config{
		Addr:            ":" + port,
		Store:           gameStore,
		SessionManager:  sessionManager,
		ShutdownTimeout: 10 * time.Second,
	})

	if err := server.ListenAndServe(ctx); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runHealthcheck(port string) {
	client := http.Client{Timeout: 2 * time.Second}
	response, err := client.Get("http://127.0.0.1:" + port + "/health")
	if err != nil || response.StatusCode != http.StatusOK {
		os.Exit(1)
	}
	_ = response.Body.Close()
}

// runCleanup periodically removes completed/drawn games older than the TTL.
func runCleanup(ctx context.Context, gameStore store.GameStore, ttl, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			count, err := gameStore.CleanupCompletedGames(ttl)
			if err != nil {
				log.Printf("cleanup: error removing completed games: %v", err)
				continue
			}
			if count > 0 {
				log.Printf("cleanup: removed %d completed/drawn game(s) older than %s", count, ttl)
			}
		}
	}
}

// envDuration reads a duration from an environment variable, returning the
// fallback if the variable is empty or cannot be parsed.
func envDuration(name string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}
