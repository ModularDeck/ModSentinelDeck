package auth

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/time/rate"
)

const (
	ErrTooManyRequests = "Too Many Requests"
)

type ctxKey string

const (
	emailKey    ctxKey = "email"
	tenantIDKey ctxKey = "tenant_id"
)

// AuthMiddleware checks for the JWT token and validates it
func AuthMiddleware(next http.Handler, dbInstance *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Extract token
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := ValidateToken(tokenStr) // Ensure ValidateToken is implemented and imported
		log.Printf("Printing token %s", tokenStr)
		if err != nil {
			log.Println("Middleware error")
			log.Println(err)
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		var exists bool
		x := dbInstance.QueryRow("SELECT EXISTS(SELECT 1 FROM token_blacklist WHERE token=$1)", tokenStr).Scan(&exists)
		if x != nil || exists {
			log.Printf("Printing token 2 %s", tokenStr)
			http.Error(w, "Token is invalid", http.StatusUnauthorized)
			return
		}

		// Add claims to request context
		ctx := r.Context()
		ctx = context.WithValue(ctx, emailKey, claims.Email)
		ctx = context.WithValue(ctx, tenantIDKey, claims.TenantID)
		r = r.WithContext(ctx)

		// Proceed to the next handler
		next.ServeHTTP(w, r)
	})
}

// GetTenantID fetches the tenant_id from the context
func GetTenantID(ctx context.Context) (int, error) {
	tenantIDValue := ctx.Value(tenantIDKey) // Fetch from context using tenantIDKey
	tenantID, ok := tenantIDValue.(int)
	log.Printf("tenantid %x", tenantID)

	if !ok {
		log.Println("Error: tenant_id not found in context or invalid type")
		return 0, errors.New("tenant_id not found in context or invalid type")
	}
	return tenantID, nil
}

// GetEmail fetches the email from the context
func GetEmail(ctx context.Context) (string, error) {
	emailValue := ctx.Value(emailKey) // Fetch from context using emailKey
	email, ok := emailValue.(string)
	log.Printf("email: %s", email)

	if !ok {
		log.Println("Error: email not found in context or invalid type")
		return "", errors.New("email not found in context or invalid type")
	}
	return email, nil
}

// User-specific rate limiter map
var (
	userLimiters = make(map[string]*rate.Limiter)
	mu           sync.Mutex
)

// Function to get or create a rate limiter for a specific user (by IP or token)
func getUserLimiter(user string, rateLimit rate.Limit, burstSize int) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	limiter, exists := userLimiters[user]
	if exists {
		return limiter
	}
	limiter = rate.NewLimiter(rateLimit, burstSize)
	return limiter
}

// RateLimitMiddleware applies rate limiting per user
func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Header.Get("Authorization") // Or use r.Header.Get("Authorization") for token-based
		const rateLimit = 1.0                 // requests per second (minimum for testing)
		const burstSize = 3                   // burst size (minimum for testing)
		limiter := getUserLimiter(user, rate.Limit(rateLimit), burstSize)
		log.Println("Rate limiting started for user", user)

		if !limiter.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
