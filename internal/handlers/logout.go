package handlers

import (
	"net/http"
	"sentinel/internal/db"
	"strings"
)

// LogoutHandler handles user logout and adds token to blacklist
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Get the JWT token from the Authorization header
	token := r.Header.Get("Authorization")

	if token == "" {
		http.Error(w, "Authorization token not found", http.StatusUnauthorized)
		return
	}

	// Save the token to the blacklist (db or cache)
	_, err := db.DB.Exec("INSERT INTO token_blacklist (token) VALUES ($1)", strings.TrimPrefix(token, "Bearer "))
	if err != nil {
		http.Error(w, "Error logging out", http.StatusInternalServerError)
		return
	}

	// Optionally, clear any cookies (if you use them)
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logged out successfully"))
}
