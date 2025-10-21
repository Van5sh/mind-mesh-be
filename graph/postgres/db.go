package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var DB *sql.DB

func InitDB() {
	var err error
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	DB, err = sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Failed to Connect to the Database : %v", err)
	}
	if err = DB.Ping(); err != nil {
		log.Fatalf("Failed to Ping the Database : %v", err)
	}
	log.Println("Connected to the Database successfully")
}
