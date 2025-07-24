package handlers

import (
	"encoding/json"
	"net/http"
)

// Fungsi untuk merespon data JSON
func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
