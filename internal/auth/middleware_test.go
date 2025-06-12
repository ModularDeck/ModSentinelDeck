package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// --- Tests ---

func TestAuthMiddleware_NoAuthHeader(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()
	handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Should not call next handler")
	}), nil) // Replace mockDB.ToSQLDB() with nil
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}
func TestAuthMiddleware(t *testing.T) {
	// Mock ValidateToken to return valid claims
	origValidateToken := ValidateToken
	defer func() { ValidateToken = origValidateToken }()
	ValidateToken = func(token string) (*Claims, error) {
		if token == "valid-token" {
			return &Claims{Email: "test@example.com", TenantID: 42}, nil
		}
		return nil, errors.New("invalid token")
	}

	// Create a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}
	defer db.Close()

	// Mock the query to return a valid result
	mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM token_blacklist WHERE token=\$1\)`).
		WithArgs("valid-token").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	// Create a test response recorder
	rr := httptest.NewRecorder()

	// Call the middleware
	handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), db)

	handler.ServeHTTP(rr, req)

	// Assert the response
	if rr.Code != http.StatusOK {
		t.Errorf("expected status OK, got %v", rr.Code)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
func TestAuthMiddleware_InvalidToken(t *testing.T) {
	origValidateToken := ValidateToken
	defer func() { ValidateToken = origValidateToken }()
	ValidateToken = func(token string) (*Claims, error) {
		return nil, errors.New("invalid token")
	}

	// Create a test request
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer invalidtoken")

	// Create a test response recorder
	rr := httptest.NewRecorder()

	// Create a mock database
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}
	defer db.Close()

	// Call the middleware
	handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Should not call next handler")
	}), db)

	handler.ServeHTTP(rr, req)

	// Assert the response
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestAuthMiddleware_BlacklistedToken(t *testing.T) {
	origValidateToken := ValidateToken
	defer func() { ValidateToken = origValidateToken }()
	ValidateToken = func(token string) (*Claims, error) {
		return &Claims{Email: "test@example.com", TenantID: 42}, nil
	}

	// Create a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}
	defer db.Close()

	// Mock the query to simulate a blacklisted token
	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM token_blacklist WHERE token=\\$1\\)").
		WithArgs("blacklistedtoken").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Create a test request
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer blacklistedtoken")

	// Create a test response recorder
	rr := httptest.NewRecorder()

	// Call the middleware
	handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Should not call next handler")
	}), db)

	handler.ServeHTTP(rr, req)

	// Assert the response
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	origValidateToken := ValidateToken
	defer func() { ValidateToken = origValidateToken }()
	ValidateToken = func(token string) (*Claims, error) {
		return &Claims{Email: "test@example.com", TenantID: 42}, nil
	}

	// Create a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}
	defer db.Close()

	// Mock the query to simulate a non-blacklisted token
	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM token_blacklist WHERE token=\\$1\\)").
		WithArgs("validtoken").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	// Create a test request
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer validtoken")

	// Create a test response recorder
	rr := httptest.NewRecorder()

	// Call the middleware
	called := false
	handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		email, err := GetEmail(r.Context())
		if err != nil || email != "test@example.com" {
			t.Errorf("expected email in context, got %v, err: %v", email, err)
		}
		tenantID, err := GetTenantID(r.Context())
		if err != nil || tenantID != 42 {
			t.Errorf("expected tenantID 42, got %v, err: %v", tenantID, err)
		}
	}), db)

	handler.ServeHTTP(rr, req)

	// Assert the response
	if !called {
		t.Error("next handler was not called")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestAuthMiddleware_NilDB(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rr := httptest.NewRecorder()

	handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Should not call next handler")
	}), nil) // Pass nil as the database instance

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 429, got %d", rr.Code)
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	origValidateToken := ValidateToken
	defer func() { ValidateToken = origValidateToken }()
	ValidateToken = func(token string) (*Claims, error) {
		return nil, errors.New("token is expired")
	}

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer expired-token")
	rr := httptest.NewRecorder()

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}
	defer db.Close()

	handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Should not call next handler")
	}), db)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer validuser")

	//rr := httptest.NewRecorder()

	handler := RateLimitMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Simulate multiple requests from the same user
	for i := 0; i < 5; i++ {
		rr := httptest.NewRecorder() // Reset the response recorder for each request
		handler.ServeHTTP(rr, req)

		if i < 3 && rr.Code != http.StatusOK {
			t.Errorf("expected 200, got %d on request %d", rr.Code, i+1)
		}
		if i >= 3 && rr.Code != http.StatusTooManyRequests {
			t.Errorf("expected 429, got %d on request %d", rr.Code, i+1)
		}
	}
}

func TestGetTenantID(t *testing.T) {
	ctx := context.WithValue(context.Background(), TenantIDKey, 42)
	tenantID, err := GetTenantID(ctx)
	if err != nil || tenantID != 42 {
		t.Errorf("expected tenantID 42, got %v, err: %v", tenantID, err)
	}

	// Test with missing tenant ID
	ctx = context.Background()
	_, err = GetTenantID(ctx)
	if err == nil {
		t.Error("expected error, got nil")
	}
}
