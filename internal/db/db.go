package db

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/jackc/pgx/v4/stdlib"
)

var DB *sql.DB

// InitDB initializes the database connection
func InitDB() {
	var err error
	dbURL := os.Getenv("DATABASE_URL")
	DB, err = sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}
}
