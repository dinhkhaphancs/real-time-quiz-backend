package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dinhkhaphancs/real-time-quiz-backend/internal/config"
	_ "github.com/lib/pq"
)

// DB is a wrapper around sql.DB
type DB struct {
	*sql.DB
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(config config.PostgresConfig) (*DB, error) {
	// Format connection string
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.Database, config.SSLMode,
	)

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	return &DB{db}, nil
}

// ExecContext wraps sql.DB's ExecContext for custom error handling if needed
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return db.DB.ExecContext(ctx, query, args...)
}

// QueryContext wraps sql.DB's QueryContext for custom error handling if needed
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return db.DB.QueryContext(ctx, query, args...)
}

// QueryRowContext wraps sql.DB's QueryRowContext for custom error handling if needed
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return db.DB.QueryRowContext(ctx, query, args...)
}

// Transaction executes a function within a database transaction
func (db *DB) Transaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // Re-throw panic after rollback
		}
	}()

	if err := fn(tx); err != nil {
		// Roll back the transaction if the function returned an error
		if rbErr := tx.Rollback(); rbErr != nil {
			// Log the rollback error but return the original error
			return fmt.Errorf("tx error: %v, rollback error: %v", err, rbErr)
		}
		return err
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}