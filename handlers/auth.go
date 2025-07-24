package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"golang-jwt/db"
	"golang-jwt/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("secret_key")

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Register(w http.ResponseWriter, r *http.Request) {
	var creds Credentials

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		RespondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "JSON Format is not valid",
		})
		return
	}

	// Validasi field kosong
	if creds.Username == "" || creds.Password == "" {
		RespondJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Username and Password is required",
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), bcrypt.DefaultCost)
	if err != nil {
		RespondJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed Encyrpt Password",
		})
		return
	}

	_, err = db.DB.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", creds.Username, string(hashedPassword))
	if err != nil {
		RespondJSON(w, http.StatusConflict, map[string]string{
			"error": "Username is Already User",
		})
		return
	}

	RespondJSON(w, http.StatusCreated, map[string]string{
		"message": "Register Success",
	})
}



func Login(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON Format is Not Valid"})
		return
	}
	if creds.Username == "" || creds.Password == "" {
		RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "Username and password is Required"})
		return
	}

	var user models.User
	err := db.DB.QueryRow("SELECT id, username, password FROM users WHERE username=$1", creds.Username).
	Scan(&user.ID, &user.Username, &user.Password)

	if err != nil {
		RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": "User Not Found"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
		RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": "Wrong Password"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
	"user_id":  user.ID,
	"username": user.Username, 
	})


	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create token"})
		return
	}

	// Simpan token di cookie (tanpa Expires → session cookie)
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,                           // karena belum pakai https
		SameSite: http.SameSiteLaxMode,
		MaxAge:   3600,
		Expires:  time.Now().Add(1 * time.Hour),
	})


	RespondJSON(w, http.StatusOK, map[string]string{"message": "Login success"})
}

func Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,                           // karena belum pakai https
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	
	})
	RespondJSON(w, http.StatusOK, map[string]string{
		"message": "Success logout",
	})
}
