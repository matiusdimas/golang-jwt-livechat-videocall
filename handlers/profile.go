package handlers

import (
	"fmt"
	"net/http"
)

func Profile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("username")
	fmt.Fprintf(w, "Welcome user #%v", userID)
}
