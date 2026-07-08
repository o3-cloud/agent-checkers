package handlers

import (
	"net/http"

	"github.com/stackable-specs/agent-checkers/src/api/openapi"
)

// ServeOpenAPIJSON returns the OpenAPI 3.1 contract as JSON.
func (h *Handlers) ServeOpenAPIJSON(w http.ResponseWriter, _ *http.Request) {
	spec, err := openapi.JSON()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate OpenAPI JSON")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(spec)
}

// ServeOpenAPIYAML returns the OpenAPI 3.1 contract as YAML.
func (h *Handlers) ServeOpenAPIYAML(w http.ResponseWriter, _ *http.Request) {
	spec, err := openapi.YAML()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate OpenAPI YAML")
		return
	}
	w.Header().Set("Content-Type", "application/yaml")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(spec)
}
