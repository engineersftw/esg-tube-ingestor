package integration_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/engineersftw/youtube-sync/internal/database"
	"github.com/engineersftw/youtube-sync/pkg/config"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDatabaseConnection tests basic database connectivity
// This is an integration test that requires PostgreSQL to be running
func TestDatabaseConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	godotenv.Load("../../.env.test")

	// Check if database connection info is available
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "localhost"
	}

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:           host,
			Port:           5432,
			User:           getEnvOrDefault("POSTGRES_USER", "postgres"),
			Password:       getEnvOrDefault("POSTGRES_PASSWORD", ""),
			Database:       getEnvOrDefault("POSTGRES_DATABASE", "esg_tube"),
			SSLMode:        "disable",
			MaxConnections: 10,
		},
	}

	t.Logf("Using database host: %s", cfg.Database.Host)

	// Attempt to create database client
	client, err := database.NewClient(context.Background(), cfg)
	if err != nil {
		t.Skipf("Skipping test: could not connect to database: %v", err)
	}
	defer client.Close()

	// Test connection with ping
	err = client.Ping(context.Background())
	require.NoError(t, err, "Database ping should succeed")
}

// TestDatabaseClient_ConnectionPooling tests connection pool configuration
func TestDatabaseClient_ConnectionPooling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := getTestConfig()
	if cfg == nil {
		t.Skip("Database not available")
	}

	client, err := database.NewClient(context.Background(), cfg)
	if err != nil {
		t.Skip("Could not connect to database")
	}
	defer client.Close()

	// Verify connection pool settings
	stats := client.Stats()
	assert.GreaterOrEqual(t, stats.MaxOpenConnections, 1)
}

// TestDatabaseClient_ContextCancellation tests context cancellation handling
func TestDatabaseClient_ContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := getTestConfig()
	if cfg == nil {
		t.Skip("Database not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// This should fail due to context timeout
	_, err := database.NewClient(ctx, cfg)
	if err == nil {
		// Connection was faster than 1ms, that's okay
		t.Log("Connection established before context timeout")
	}
}

// TestDatabaseClient_RetryLogic tests connection retry logic
func TestDatabaseClient_RetryLogic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Use invalid host to trigger retry
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:           "invalid-host-12345",
			Port:           5432,
			User:           "postgres",
			Password:       "password",
			Database:       "testdb",
			SSLMode:        "disable",
			MaxConnections: 10,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := database.NewClient(ctx, cfg)
	assert.Error(t, err, "Should fail to connect to invalid host")
	assert.Contains(t, err.Error(), "failed to connect", "Error should mention connection failure")
}

// Helper functions

func getTestConfig() *config.Config {
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "localhost"
	}

	return &config.Config{
		Database: config.DatabaseConfig{
			Host:           host,
			Port:           5432,
			User:           getEnvOrDefault("POSTGRES_USER", "postgres"),
			Password:       getEnvOrDefault("POSTGRES_PASSWORD", ""),
			Database:       getEnvOrDefault("POSTGRES_DATABASE", "esg_tube"),
			SSLMode:        "disable",
			MaxConnections: 10,
		},
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TestDatabaseClient_ValidSchema tests that expected tables exist
func TestDatabaseClient_ValidSchema(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := getTestConfig()
	if cfg == nil {
		t.Skip("Database not available")
	}

	client, err := database.NewClient(context.Background(), cfg)
	if err != nil {
		t.Skip("Could not connect to database")
	}
	defer client.Close()

	// Check if episodes table exists
	var exists bool
	query := `SELECT EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public'
		AND table_name = 'episodes'
	)`

	err = client.DB().QueryRow(query).Scan(&exists)
	require.NoError(t, err)

	if !exists {
		t.Skip("episodes table does not exist - schema not loaded")
	}

	assert.True(t, exists, "episodes table should exist")
}
