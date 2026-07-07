package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stackable-specs/stack-base-go/internal/app"
)

func TestHandlerReturnsExpectedResponse(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		statusCode int
		body       string
	}{
		{name: "returns service name", path: "/", statusCode: http.StatusOK, body: "stack-base-go\n"},
		{name: "returns healthy status", path: "/healthz", statusCode: http.StatusNoContent},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			response := httptest.NewRecorder()

			app.Handler().ServeHTTP(response, request)

			if response.Code != test.statusCode {
				t.Fatalf("status code = %d, want %d", response.Code, test.statusCode)
			}
			if response.Body.String() != test.body {
				t.Fatalf("body = %q, want %q", response.Body.String(), test.body)
			}
		})
	}
}
