package handler

import (
	"net/http"

	"github.com/jwx233s/a-service/pkg/response"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	// CORS preflight
	response.SetHeaders(w)
	if r.Method == "OPTIONS" {
		return
	}

	// Simple test endpoint: return a JSON hello message
	response.Success(w, map[string]string{"message": "Hello World"})
}
