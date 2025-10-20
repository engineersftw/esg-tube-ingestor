package contract_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/engineersftw/youtube-sync/internal/database"
	"github.com/engineersftw/youtube-sync/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEpisodesTableSchema tests that the episodes table exists and has expected columns
func TestEpisodesTableSchema(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping contract test in short mode")
	}

	client := setupTestDatabase(t)
	if client == nil {
		t.Skip("Database not available")
	}
	defer client.Close()

	// Check if episodes table exists
	var exists bool
	query := `SELECT EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public'
		AND table_name = 'episodes'
	)`
	err := client.DB().QueryRow(query).Scan(&exists)
	require.NoError(t, err)

	if !exists {
		t.Skip("episodes table does not exist - create schema first")
	}

	// Verify expected columns exist
	expectedColumns := []string{
		"id", "video_id", "title", "description", "published_at",
		"image1", "image2", "image3", "view_count", "video_site",
		"active", "sort_order", "created_at", "updated_at",
	}

	for _, col := range expectedColumns {
		var columnExists bool
		query := `SELECT EXISTS (
			SELECT FROM information_schema.columns
			WHERE table_name = 'episodes'
			AND column_name = $1
		)`
		err := client.DB().QueryRow(query, col).Scan(&columnExists)
		require.NoError(t, err)
		assert.True(t, columnExists, "Column %s should exist in episodes table", col)
	}
}

// TestEpisodesTableUpsert tests the upsert operation (insert and update)
func TestEpisodesTableUpsert(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping contract test in short mode")
	}

	client := setupTestDatabase(t)
	if client == nil {
		t.Skip("Database not available")
	}
	defer client.Close()

	// Create a test episode
	testVideoID := "TEST_" + time.Now().Format("20060102150405")
	defer cleanupTestEpisode(t, client, testVideoID)

	episode := &database.Episode{
		VideoID:     testVideoID,
		Title:       "Test Video Title",
		Description: "Test video description",
		PublishedAt: time.Now().Add(-24 * time.Hour),
		Image1:      "https://example.com/thumb1.jpg",
		Image2:      "https://example.com/thumb2.jpg",
		Image3:      "https://example.com/thumb3.jpg",
		ViewCount:   1000,
		VideoSite:   1,
		Active:      true,
	}

	// Test INSERT
	err := client.UpsertEpisode(context.Background(), episode)
	require.NoError(t, err, "First upsert (insert) should succeed")

	// Verify episode was inserted
	var dbEpisode database.Episode
	query := `SELECT id, video_id, title, description, published_at, view_count, video_site, active
		FROM episodes WHERE video_id = $1`
	err = client.DB().QueryRow(query, testVideoID).Scan(
		&dbEpisode.ID,
		&dbEpisode.VideoID,
		&dbEpisode.Title,
		&dbEpisode.Description,
		&dbEpisode.PublishedAt,
		&dbEpisode.ViewCount,
		&dbEpisode.VideoSite,
		&dbEpisode.Active,
	)
	require.NoError(t, err)
	assert.Equal(t, testVideoID, dbEpisode.VideoID)
	assert.Equal(t, "Test Video Title", dbEpisode.Title)
	assert.Equal(t, 1000, dbEpisode.ViewCount)
	assert.True(t, dbEpisode.Active)

	// Test UPDATE (upsert with changed data)
	episode.Title = "Updated Title"
	episode.ViewCount = 2000
	episode.Active = false

	err = client.UpsertEpisode(context.Background(), episode)
	require.NoError(t, err, "Second upsert (update) should succeed")

	// Verify episode was updated
	err = client.DB().QueryRow(query, testVideoID).Scan(
		&dbEpisode.ID,
		&dbEpisode.VideoID,
		&dbEpisode.Title,
		&dbEpisode.Description,
		&dbEpisode.PublishedAt,
		&dbEpisode.ViewCount,
		&dbEpisode.VideoSite,
		&dbEpisode.Active,
	)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", dbEpisode.Title)
	assert.Equal(t, 2000, dbEpisode.ViewCount)
	assert.False(t, dbEpisode.Active, "Active should be updated to false")
}

// TestEpisodesTableUniqueConstraint tests that video_id is unique
func TestEpisodesTableUniqueConstraint(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping contract test in short mode")
	}

	client := setupTestDatabase(t)
	if client == nil {
		t.Skip("Database not available")
	}
	defer client.Close()

	testVideoID := "UNIQUE_TEST_" + time.Now().Format("20060102150405")
	defer cleanupTestEpisode(t, client, testVideoID)

	// Insert first episode
	query := `INSERT INTO episodes (video_id, title, video_site, active)
		VALUES ($1, $2, 1, true)`
	_, err := client.DB().Exec(query, testVideoID, "First")
	require.NoError(t, err)

	// Try to insert duplicate video_id (should fail)
	_, err = client.DB().Exec(query, testVideoID, "Duplicate")
	assert.Error(t, err, "Duplicate video_id should fail")
	assert.Contains(t, err.Error(), "duplicate", "Error should mention duplicate key")
}

// Helper functions

func setupTestDatabase(t *testing.T) *database.Client {
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "localhost"
	}

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:           host,
			Port:           5432,
			User:           getEnvOrDefault("POSTGRES_USER", "postgres"),
			Password:       getEnvOrDefault("POSTGRES_PASSWORD", "postgres"),
			Database:       getEnvOrDefault("POSTGRES_DATABASE", "esg_tube"),
			SSLMode:        "disable",
			MaxConnections: 10,
		},
	}

	client, err := database.NewClient(context.Background(), cfg)
	if err != nil {
		t.Logf("Could not connect to database: %v", err)
		return nil
	}

	return client
}

func cleanupTestEpisode(t *testing.T, client *database.Client, videoID string) {
	_, err := client.DB().Exec("DELETE FROM episodes WHERE video_id = $1", videoID)
	if err != nil {
		t.Logf("Warning: Could not cleanup test episode: %v", err)
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
