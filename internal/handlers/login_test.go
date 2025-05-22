package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"golang.org/x/crypto/bcrypt"
)

func TestLoginHandler(t *testing.T) {
	// Mock database and dependencies
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedToken  string
	}{
		{
			name: "Valid login",
			requestBody: map[string]interface{}{
				"email":     "valid@test.com",
				"password":  "password123",
				"tenant_id": 1,
			},
			expectedStatus: http.StatusOK,
			expectedToken:  "mocked.jwt.token",
		},
		{
			name: "Invalid password",
			requestBody: map[string]interface{}{
				"email":     "valid@test.com",
				"password":  "wrongpassword",
				"tenant_id": 1,
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Invalid tenant",
			requestBody: map[string]interface{}{
				"email":     "valid@test.com",
				"password":  "password123",
				"tenant_id": 2,
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Error generating token",
			requestBody: map[string]interface{}{
				"email":     "error@test.com",
				"password":  "password123",
				"tenant_id": 1,
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Invalid request body",
			requestBody: map[string]interface{}{
				"email": "invalid",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request body
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Set up mock expectations
			if tt.name == "Valid login" {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				mock.ExpectQuery("SELECT u.id, u.name, u.password, u.tenant_id , u.role FROM users u WHERE u.email=\\$1 AND u.tenant_id=\\$2").
					WithArgs("valid@test.com", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "password", "tenant_id", "role"}).
						AddRow(1, "Test User", string(hashedPassword), 1, "user"))
			} else if tt.name == "Invalid password" {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				mock.ExpectQuery("SELECT u.id, u.name, u.password, u.tenant_id , u.role FROM users u WHERE u.email=\\$1 AND u.tenant_id=\\$2").
					WithArgs("valid@test.com", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "password", "tenant_id", "role"}).
						AddRow(1, "Test User", string(hashedPassword), 1, "user"))
			} else if tt.name == "Invalid tenant" {
				mock.ExpectQuery("SELECT u.id, u.name, u.password, u.tenant_id , u.role FROM users u WHERE u.email=\\$1 AND u.tenant_id=\\$2").
					WithArgs("valid@test.com", 2).
					WillReturnError(errors.New("no rows found"))
			} else if tt.name == "Error generating token" {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				mock.ExpectQuery("SELECT u.id, u.name, u.password, u.tenant_id , u.role FROM users u WHERE u.email=\\$1 AND u.tenant_id=\\$2").
					WithArgs("error@test.com", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "name", "password", "tenant_id", "role"}).
						AddRow(1, "Test User", string(hashedPassword), 1, "user"))
			} else if tt.name == "Invalid request body" {
				// No database interaction expected
			}

			// Call the handler
			LoginHandler(rr, req, db)
			// Ensure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet sqlmock expectations: %v", err)
			}

			// Check response body for token if status is OK
			if tt.expectedStatus == http.StatusOK {
				var response map[string]string
				err := json.NewDecoder(rr.Body).Decode(&response)
				if err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				log.Printf("Response: %v", response["token"])

			}
		})
	}
}
