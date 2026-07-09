package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/stackable-specs/agent-checkers/internal/app/session"
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
