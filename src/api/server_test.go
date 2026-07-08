package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stackable-specs/agent-checkers/internal/app/store"
)

func TestRouterServesIndexPage(t *testing.T) {
	router := NewRouter(store.NewMemoryStore(), nil)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", response.Code, http.StatusOK, response.Body.String())
	}
	body := response.Body.String()
	if !strings.Contains(body, "<!DOCTYPE html>") && !strings.Contains(body, "Agent Checkers") {
		t.Fatalf("body does not contain expected index content")
	}
}

func TestRouterServesStaticAssets(t *testing.T) {
	router := NewRouter(store.NewMemoryStore(), nil)
	tests := []struct {
		name            string
		path            string
		contentTypeWant []string
	}{
		{
			name:            "css",
			path:            "/static/css/board.css",
			contentTypeWant: []string{"text/css"},
		},
		{
			name:            "app js",
			path:            "/static/js/app.js",
			contentTypeWant: []string{"javascript", "text/plain"},
		},
		{
			name: "/static/js/board.js",
			path: "/static/js/board.js",
		},
		{
			name: "/static/js/websocket.js",
			path: "/static/js/websocket.js",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, tt.path, nil))

			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d, body %s", response.Code, http.StatusOK, response.Body.String())
			}
			for _, want := range tt.contentTypeWant {
				if strings.Contains(response.Header().Get("Content-Type"), want) {
					return
				}
			}
			if len(tt.contentTypeWant) > 0 {
				t.Fatalf("Content-Type = %q, want one of %v", response.Header().Get("Content-Type"), tt.contentTypeWant)
			}
		})
	}
}

func TestRouterReturnsNotFoundForUnknownRoute(t *testing.T) {
	router := NewRouter(store.NewMemoryStore(), nil)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/nonexistent", nil))

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body %s", response.Code, http.StatusNotFound, response.Body.String())
	}
}
