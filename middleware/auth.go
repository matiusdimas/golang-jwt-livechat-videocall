package middleware

import (
	"context"
	"net/http"
	"strings"

	"golang-jwt/handlers"
	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("secret_key") // sesuaikan ini dengan milikmu

func JWTAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Ambil token dari cookie
		cookie, err := r.Cookie("token")
		if err != nil {
			handlers.RespondJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "need a token",
			})
			return
		}

		tokenStr := cookie.Value
		if strings.TrimSpace(tokenStr) == "" {
			handlers.RespondJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Token empty",
			})
			return
		}

		// Parse dan validasi token
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			handlers.RespondJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Token is invalid",
			})
			return
		}

		// Ambil user_id dari claims
		userID, ok := claims["user_id"]
		if !ok {
			handlers.RespondJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "user_id not found in token",
			})
			return
		}

		// Ambil username dari claims
		username, ok := claims["username"]
		if !ok {
			handlers.RespondJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "username not found in token",
			})
			return
		}

		// Simpan user_id dan username ke context
		ctx := context.WithValue(r.Context(), "user_id", userID)
		ctx = context.WithValue(ctx, "username", username)

		// Teruskan request ke handler berikutnya
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}



func RejectIfAuthenticated(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil || strings.TrimSpace(cookie.Value) == "" {
			// Tidak ada token, silakan lanjut
			next(w, r)
			return
		}

		tokenStr := cookie.Value
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err == nil && token.Valid {
			// Token valid → tolak akses
			handlers.RespondJSON(w, http.StatusForbidden, map[string]string{
				"error": "You are Logged In",
			})
			return
		}

		// Token tidak valid → izinkan lanjut
		next(w, r)
	}
}

func JWTAuthWS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if cookie, err := r.Cookie("token"); err == nil {
			claims := jwt.MapClaims{}
			if token, err := jwt.ParseWithClaims(cookie.Value, claims, func(t *jwt.Token) (interface{}, error) {
				return jwtKey, nil
			}); err == nil && token.Valid {
				if userID, ok := claims["user_id"]; ok {
					ctx = context.WithValue(ctx, "user_id", userID)
				}
				if username, ok := claims["username"]; ok {
					ctx = context.WithValue(ctx, "username", username)
				}
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
