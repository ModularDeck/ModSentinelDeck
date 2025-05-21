package auth

import (
	"errors"
	"log"
	"sentinel/internal/db"
	"time"

	"github.com/golang-jwt/jwt"
)

// Claims struct to hold JWT claims
type Claims struct {
	Email    string `json:"email"`
	TenantID int    `json:"tenant_id"` // Add Tenant ID to support multi-tenancy
	Role     string `json:"role"`
	jwt.StandardClaims
}

const jwtKey string = "ikud1U6vzc8OhVoNw0vadTKt7MA20Vlk"

// GenerateJWT creates a JWT for authenticated users
func GenerateJWT(email string, tenantID int, role string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	log.Println("JWT Generation started. jwtkey")
	claims := &Claims{
		Email:    email,
		TenantID: tenantID,
		Role:     role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	log.Println(claims)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtKey := []byte(jwtKey)
	log.Println("JWT Generation started. jwtkey")
	log.Println(jwtKey)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// Patch point for blacklist check
var isTokenBlacklisted = func(token string) (bool, error) {
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM token_blacklist WHERE token=$1)", token).Scan(&exists)
	return exists, err
}

// Patch point for ValidateToken to allow mocking in tests
var ValidateToken = validateToken

func validateToken(tokenStr string) (*Claims, error) {
	log.Println("Starting JWT token validation...")

	// Fetch JWT secret from environment variables
	if jwtKey == "" {
		log.Println("JWT secret not found in environment variables.")
		return nil, errors.New("JWT secret not found")
	}

	log.Println("JWT secret successfully loaded.")
	claims := &Claims{}

	// Parse the token with claims
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtKey), nil
	})

	if err != nil {
		log.Printf("Error parsing token: %v\n", err)
		// Handle specific JWT validation errors
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, errors.New("token is expired")
			}
			if ve.Errors&jwt.ValidationErrorSignatureInvalid != 0 {
				return nil, errors.New("invalid token signature")
			}
		}
		return nil, errors.New("error validating token")
	}

	// Check if the token is blacklisted
	exists, err := isTokenBlacklisted(tokenStr)
	if err != nil {
		log.Println("Token blacklist check error.")
		return nil, errors.New("token is expired")
	}
	if exists {
		log.Println("Token in token_blacklist.")
		return nil, errors.New("token is expired")
	}

	if !token.Valid {
		log.Println("Token is not valid.")
		return nil, errors.New("invalid token")
	}

	log.Printf("Token valid, claims: %v\n", claims)
	return claims, nil
}
