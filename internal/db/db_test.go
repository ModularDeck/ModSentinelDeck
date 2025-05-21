package db

import (
	"database/sql"
	"os"
	"testing"
)

func TestInitDB_SetsDBVariable(t *testing.T) {
	// Save and restore original environment variable and DB
	origDB := DB
	origEnv := os.Getenv("DATABASE_URL")
	defer func() {
		DB = origDB
		os.Setenv("DATABASE_URL", origEnv)
	}()

	// Set a fake database URL (won't actually connect)
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/dbname?sslmode=disable")

	// Patch sql.Open to avoid real DB connection
	origOpen := sqlOpen
	defer func() { sqlOpen = origOpen }()
	sqlOpen = func(driverName, dataSourceName string) (*sql.DB, error) {
		return &sql.DB{}, nil
	}

	InitDB()

	if DB == nil {
		t.Error("DB should not be nil after InitDB")
	}
}

// Patch point for sql.Open to allow mocking in tests.
var sqlOpen = sql.Open

// No need for init() to patch sql.Open; use sqlOpen variable as patch point in your code.
