package main

import (
	"fmt"
	"log"
	"net/http"

	"golang-jwt/db"
	"golang-jwt/handlers"
	"golang-jwt/middleware"
)

func main() {
	db.Init()

	mux := http.NewServeMux()

	// Register dan Login ditolak kalau sudah login (punya token)
	mux.HandleFunc("/register", withMethod("POST", middleware.RejectIfAuthenticated(handlers.Register)))
	mux.HandleFunc("/login", withMethod("POST", middleware.RejectIfAuthenticated(handlers.Login)))

	// Logout hanya bisa jika token valid
	mux.HandleFunc("/logout", withMethod("DELETE", middleware.JWTAuth(handlers.Logout)))

	// Profile hanya untuk user yang login
	mux.HandleFunc("/profile", middleware.JWTAuth(withMethod("GET", handlers.Profile)))

	// websocket live chat
	mux.HandleFunc("/ws", middleware.JWTAuthWS(handlers.WebSocketHandler))

	// websocket vid call
	mux.HandleFunc("/ws-video", middleware.JWTAuthWS(handlers.VideoCallWebSocketHandler))

	fs := http.FileServer(http.Dir("./views"))
	mux.Handle("/", fs)

	fmt.Println("Server Listening :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}


// Wrapper untuk membatasi method (GET, POST, dll)
func withMethod(method string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.Header().Set("Allow", method)
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		handler(w, r)
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Hanya izinkan dari frontend lokal (bisa diganti)
		if origin == "http://127.0.0.1:5500" || origin == "http://localhost:5500" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
