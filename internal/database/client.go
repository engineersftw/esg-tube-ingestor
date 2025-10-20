package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/engineersftw/youtube-sync/pkg/config"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// Client wraps the database connection with additional functionality
type Client struct {
	db *sql.DB
}

// NewClient creates a new database client with connection pooling and retry logic
func NewClient(ctx context.Context, cfg *config.Config) (*Client, error) {
	dsn := cfg.GetDSN()

	// Configure exponential backoff for retries
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 30 * time.Second // Total retry time limit

	var db *sql.DB
	var err error

	// Retry connection with exponential backoff
	operation := func() error {
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}

		// Test connection
		if err = db.PingContext(ctx); err != nil {
			db.Close()
			return fmt.Errorf("failed to ping database: %w", err)
		}

		return nil
	}

	// Use backoff with context
	backoffWithContext := backoff.WithContext(b, ctx)
	if err := backoff.Retry(operation, backoffWithContext); err != nil {
		return nil, fmt.Errorf("failed to connect to database after retries: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.Database.MaxConnections)
	db.SetMaxIdleConns(cfg.Database.MaxConnections / 2)
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(10 * time.Minute)

	return &Client{db: db}, nil
}

// DB returns the underlying *sql.DB for direct access when needed
func (c *Client) DB() *sql.DB {
	return c.db
}

// Close closes the database connection
func (c *Client) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// Ping checks if the database connection is alive
func (c *Client) Ping(ctx context.Context) error {
	return c.db.PingContext(ctx)
}

// Stats returns database connection pool statistics
func (c *Client) Stats() sql.DBStats {
	return c.db.Stats()
}

// BeginTx starts a new database transaction
func (c *Client) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return c.db.BeginTx(ctx, opts)
}

// Exec executes a query without returning rows
func (c *Client) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return c.db.ExecContext(ctx, query, args...)
}

// Query executes a query that returns rows
func (c *Client) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return c.db.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that returns a single row
func (c *Client) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return c.db.QueryRowContext(ctx, query, args...)
}
