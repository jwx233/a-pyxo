package response

import (
	"encoding/json"
	"net/http"
)

func SetHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func JSON(w http.ResponseWriter, data []byte) {
	SetHeaders(w)
	w.Write(data)
}

func Error(w http.ResponseWriter, msg string, code int) {
	SetHeaders(w)
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func Success(w http.ResponseWriter, data interface{}) {
	SetHeaders(w)
	json.NewEncoder(w).Encode(data)
}
