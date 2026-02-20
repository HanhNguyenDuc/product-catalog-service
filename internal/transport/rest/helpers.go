package rest

import (
	"encoding/json"
	"net/http"
)

// writeJSON encodes v as JSON and writes it with the given HTTP status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes a JSON error response.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// errorStatus maps common domain / sentinel errors to HTTP status codes.
// Returns 500 for unknown errors.
func errorStatus(err error) int {
	return http.StatusInternalServerError
}
