package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/stackable-specs/stack-base-go/internal/app"
)

func main() {
	healthcheck := flag.Bool("healthcheck", false, "check whether the service is healthy")
	flag.Parse()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if *healthcheck {
		client := http.Client{Timeout: 2 * time.Second}
		response, err := client.Get("http://127.0.0.1:" + port + "/healthz")
		if err != nil || response.StatusCode != http.StatusNoContent {
			os.Exit(1)
		}
		_ = response.Body.Close()
		return
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           app.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
