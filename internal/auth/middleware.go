package auth

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/time/rate"
)

const (
	ErrTooManyRequests = "Too Many Requests"
)

type CtxKey string

const (
	EmailKey    CtxKey = "email"
	TenantIDKey CtxKey = "tenant_id"
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
		ctx = context.WithValue(ctx, EmailKey, claims.Email)
		ctx = context.WithValue(ctx, TenantIDKey, claims.TenantID)
		r = r.WithContext(ctx)

		// Proceed to the next handler
		next.ServeHTTP(w, r)
	})
}

// GetTenantID fetches the tenant_id from the context
func GetTenantID(ctx context.Context) (int, error) {
	tenantIDValue := ctx.Value(TenantIDKey) // Fetch from context using tenantIDKey
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
	emailValue := ctx.Value(EmailKey) // Fetch from context using emailKey
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

func getUserLimiter(user string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	limiter, exists := userLimiters[user]
	if !exists {
		// Fetch rate limit configuration from environment variables or a config file
		rateLimit := 1  // Default to 1 request per second
		burstLimit := 3 // Default to a burst of 3 requests

		// Example: Fetch from environment variables (you can replace this with your config logic)
		if rl, ok := os.LookupEnv("RATE_LIMIT"); ok {
			if parsedRate, err := strconv.Atoi(rl); err == nil {
				rateLimit = parsedRate
			}
		}
		if bl, ok := os.LookupEnv("BURST_LIMIT"); ok {
			if parsedBurst, err := strconv.Atoi(bl); err == nil {
				burstLimit = parsedBurst
			}
		}

		limiter = rate.NewLimiter(rate.Limit(rateLimit), burstLimit)
		userLimiters[user] = limiter
	}
	return limiter
}

// RateLimitMiddleware applies rate limiting per user
func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use the Authorization header as the user identifier
		if r.URL.Path == "/health" || r.URL.Path == "/register" || r.URL.Path == "/login" || r.URL.Path == "/logout" {
			next.ServeHTTP(w, r)
			return
		}
		// Extract user identifier from the Authorization header
		user := r.Header.Get("Authorization")
		if user == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		limiter := getUserLimiter(user)
		if !limiter.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
