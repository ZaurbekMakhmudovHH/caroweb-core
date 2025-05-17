package migrations

import (
	"fmt"

	"log"

	"os"

	"github.com/golang-migrate/migrate/v4"

	// Import PostgresSQL database driver for migrations
	_ "github.com/golang-migrate/migrate/v4/database/postgres"

	// Import Migrate driver for migrations
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations() {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)

	m, err := migrate.New(
		"file://internal/infrastructure/db/migrations",
		dsn,
	)
	if err != nil {
		log.Fatalf("Could not create migration instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("✅ Database migrations ran successfully")
}
