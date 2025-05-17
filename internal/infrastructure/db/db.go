package database

import (
	"fmt"

	"log"

	"os"

	"github.com/jmoiron/sqlx"

	"github.com/joho/godotenv"

	// Import PostgresSQL driver for database/sql
	_ "github.com/lib/pq"
)

func InitDB() *sqlx.DB {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	log.Println("Using DSN:", dsn)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}

	return db
}
