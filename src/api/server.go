// Package api configures the REST HTTP server.
package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/stackable-specs/agent-checkers/internal/app/session"
	"github.com/stackable-specs/agent-checkers/internal/app/store"
	"github.com/stackable-specs/agent-checkers/src/api/dto"
	"github.com/stackable-specs/agent-checkers/src/api/handlers"
	apiws "github.com/stackable-specs/agent-checkers/src/api/websocket"
)

// Server wraps the HTTP server and its dependencies.
type Server struct {
	httpServer      *http.Server
	shutdownTimeout time.Duration
}

// Config controls server construction.
type Config struct {
	Addr            string
	Store           store.GameStore
	SessionManager  *session.Manager
	ShutdownTimeout time.Duration
}

// NewServer creates an HTTP server with all REST routes registered.
func NewServer(config Config) *Server {
	router := NewRouter(config.Store, config.SessionManager)
	return &Server{
		shutdownTimeout: config.ShutdownTimeout,
		httpServer: &http.Server{
			Addr:              config.Addr,
			Handler:           router,
			ReadHeaderTimeout: 5 * time.Second,
		},
	}
}

// NewRouter creates the REST router.
func NewRouter(gameStore store.GameStore, sessionManager *session.Manager) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	router.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		writeRouterError(w, http.StatusNotFound, "route not found")
	})
	router.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
		writeRouterError(w, http.StatusMethodNotAllowed, "method not allowed")
	})

	hub := apiws.NewHub()
	handlers.NewWithBroadcaster(gameStore, sessionManager, hub).RegisterRoutes(router)
	router.Handle("/api/v1/games/{id}/ws", apiws.NewWebSocketHandler(hub, sessionManager, gameStore))
	return router
}

// ListenAndServe starts the HTTP server and shuts it down when ctx is canceled.
func (s *Server) ListenAndServe(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		if s.shutdownTimeout == 0 {
			s.shutdownTimeout = 10 * time.Second
		}
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
		defer cancel()
		return s.httpServer.Shutdown(shutdownCtx)
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func writeRouterError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(dto.ErrorResponse{
		Error:      message,
		StatusCode: status,
	})
}
