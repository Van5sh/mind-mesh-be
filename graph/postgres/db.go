package postgres

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var DB *sql.DB

func InitDB() {
	var err error

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	DB, err = sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("Failed to ping PostgreSQL: %v", err)
	}

	log.Println("✅ Connected to PostgreSQL successfully!")
}
