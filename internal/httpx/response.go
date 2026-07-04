package httpx

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Message string              `json:"message"`
	Errors  map[string][]string `json:"errors,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func WriteError(w http.ResponseWriter, status int, message string, errors map[string][]string) {
	WriteJSON(w, status, ErrorResponse{
		Message: message,
		Errors:  errors,
	})
}
