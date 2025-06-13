package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"context"

	"sentinel/internal/auth"

	"github.com/DATA-DOG/go-sqlmock"
	"golang.org/x/crypto/bcrypt"
)

func TestUpdateUserDetails(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	tests := []struct {
		name           string
		requestBody    map[string]any
		expectedStatus int
		mockSetup      func()
	}{
		{
			name: "Valid update with password",
			requestBody: map[string]any{
				"user_id":     1,
				"name":        "Updated Name",
				"email":       "updated@test.com",
				"password":    "newpassword123",
				"tenant_name": "Updated Tenant",
				"team_name":   "Updated Team",
				"role":        "admin",
			},
			expectedStatus: http.StatusOK,
			mockSetup: func() {
				_, _ = bcrypt.GenerateFromPassword([]byte("newpassword123"), bcrypt.DefaultCost)
				mock.ExpectQuery("SELECT id, email, role FROM users WHERE id=\\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "email", "role"}).AddRow(1, "current@test.com", "admin"))
				mock.ExpectExec("UPDATE users SET name=\\$1, email=\\$2, password=\\$3 WHERE id=\\$4").
					WithArgs("Updated Name", "updated@test.com", sqlmock.AnyArg(), 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("UPDATE tenants SET name=\\$1 WHERE id=\\(SELECT tenant_id FROM users WHERE id=\\$2\\)").
					WithArgs("Updated Tenant", 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectQuery("SELECT id FROM teams WHERE name=\\$1 AND tenant_id=\\(SELECT tenant_id FROM users WHERE id=\\$2\\)").
					WithArgs("Updated Team", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectExec("UPDATE user_teams SET team_id=\\$1 WHERE user_id=\\$2").
					WithArgs(1, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("UPDATE users SET role=\\$1 WHERE id=\\$2").
					WithArgs("admin", 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name: "Unauthorized user",
			requestBody: map[string]any{
				"user_id": 2,
				"name":    "Unauthorized Update",
			},
			expectedStatus: http.StatusUnauthorized,
			mockSetup: func() {
				mock.ExpectQuery("SELECT id, email, role FROM users WHERE id=\\$1").
					WithArgs(2).
					WillReturnRows(sqlmock.NewRows([]string{"id", "email", "role"}).AddRow(2, "other@test.com", "member"))
			},
		},
		{
			name: "Invalid user ID",
			requestBody: map[string]interface{}{
				"user_id": "invalid",
			},
			expectedStatus: http.StatusBadRequest,
			mockSetup:      func() {},
		},
		{
			name: "Error updating user",
			requestBody: map[string]interface{}{
				"user_id": 1,
				"name":    "Error Update",
			},
			expectedStatus: http.StatusInternalServerError,
			mockSetup: func() {
				mock.ExpectQuery("SELECT id, email, role FROM users WHERE id=\\$1").
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "email", "role"}).AddRow(1, "current@test.com", "admin"))
				mock.ExpectExec("UPDATE users SET name=\\$1 WHERE id=\\$2").
					WithArgs("Error Update", 1).
					WillReturnError(errors.New("update error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/update-user", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Mock authentication or context setup
			ctx := context.Background()
			ctx = context.WithValue(ctx, auth.EmailKey, "current@test.com")
			ctx = context.WithValue(ctx, auth.TenantIDKey, 123)
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()

			UpdateUserDetails(rr, req, db)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet sqlmock expectations: %v", err)
			}
		})
	}
}
