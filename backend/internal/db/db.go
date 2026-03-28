package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

// DB holds the database connection
type DB struct {
	Conn *sql.DB
}

// New creates a new database connection
func New() (*DB, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_NAME")
		dbURL = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbname)
	}

	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)

	fmt.Println("✅ Database connected successfully")

	return &DB{Conn: conn}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.Conn.Close()
}

// RunMigrations runs the migration files
func (db *DB) RunMigrations() error {
	// Read and execute migration files
	migrations := []string{
		"migrations/001_initial_schema.sql",
		"migrations/002_seed_data.sql",
	}

	for _, migration := range migrations {
		content, err := os.ReadFile(migration)
		if err != nil {
			// Try with backend prefix
			content, err = os.ReadFile("backend/" + migration)
			if err != nil {
				fmt.Printf("⚠️  Could not read migration %s: %v\n", migration, err)
				continue
			}
		}

		if _, err := db.Conn.Exec(string(content)); err != nil {
			fmt.Printf("⚠️  Migration %s failed (may already exist): %v\n", migration, err)
		} else {
			fmt.Printf("✅ Migration %s applied successfully\n", migration)
		}
	}

	return nil
}
