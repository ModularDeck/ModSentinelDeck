// internal/handlers/user.go

package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sentinel/internal/auth"
	"sentinel/internal/db"
	"sentinel/internal/models"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// LoginHandler verifies the user's credentials and returns a JWT with tenant support
// It also sets the JWT token in a cookie for the client
// The function expects the request body to contain email, password, and tenant_id
// The function also checks if the user exists in the database and if the password matches
// If the user is valid, it generates a JWT token and sets it in a cookie
// If the user is invalid, it returns an error response
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var dbUser models.User

	// Create a struct to parse tenant_id from the request body
	var loginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		TenantID int    `json:"tenant_id"`
	}

	// Decode request body to get email, password, and tenant_id
	err := json.NewDecoder(r.Body).Decode(&loginRequest)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Fetch the user and tenant ID from the database
	err = db.DB.QueryRow(`
		SELECT u.id, u.name, u.password, u.tenant_id , u.role
		FROM users u 
		WHERE u.email=$1 AND u.tenant_id=$2`, loginRequest.Email, loginRequest.TenantID).Scan(&dbUser.ID, &dbUser.Name, &dbUser.Password, &dbUser.TenantID, &dbUser.Role)

	if err != nil {
		http.Error(w, "Invalid username, password, or tenant", http.StatusUnauthorized)
		return
	}

	// Compare provided password with the hashed password in the database
	err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(loginRequest.Password))
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Create JWT token with email and tenant_id
	tokenString, err := auth.GenerateJWT(loginRequest.Email, loginRequest.TenantID, dbUser.Role)
	if err != nil {
		log.Println("Error generating JWT token:", err)
		http.Error(w, "Could not create token", http.StatusInternalServerError)
		return
	}

	// Set token expiration time (24 hours)
	expirationTime := time.Now().Add(24 * time.Hour)

	// Set JWT token in the response as a cookie
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})

	// Return token in JSON response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}
