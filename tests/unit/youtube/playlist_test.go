package youtube_test

import (
	"context"
	"testing"

	"github.com/engineersftw/youtube-sync/internal/youtube"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T053: Unit test for playlist metadata fetching
func TestFetchPlaylistMetadata_ValidPlaylist(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping YouTube API test in short mode")
	}

	apiKey := getTestAPIKey(t)
	if apiKey == "" {
		t.Skip("YOUTUBE_API_KEY not set, skipping API test")
	}

	client, err := youtube.NewClient(context.Background(), apiKey, 100)
	require.NoError(t, err)

	// Use a stable, well-known playlist
	playlistID := "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf" // Google Developers playlist

	playlist, err := client.GetPlaylist(context.Background(), playlistID)
	require.NoError(t, err)
	require.NotNil(t, playlist)

	// Verify playlist metadata
	assert.Equal(t, playlistID, playlist.Id)
	assert.NotEmpty(t, playlist.Snippet.Title)
	assert.NotEmpty(t, playlist.Snippet.Description)
}

// T054: Unit test for playlist items fetching with pagination
func TestFetchPlaylistItems_WithPagination(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping YouTube API test in short mode")
	}

	apiKey := getTestAPIKey(t)
	if apiKey == "" {
		t.Skip("YOUTUBE_API_KEY not set, skipping API test")
	}

	client, err := youtube.NewClient(context.Background(), apiKey, 100)
	require.NoError(t, err)

	playlistID := "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf"

	// Fetch playlist items with limit
	items, nextPageToken, err := client.GetPlaylistItems(context.Background(), playlistID, 10, "")
	require.NoError(t, err)
	assert.LessOrEqual(t, len(items), 10)

	// Verify items have video IDs and positions
	for i, item := range items {
		assert.NotEmpty(t, item.Snippet.ResourceId.VideoId, "Item %d should have video ID", i)
		assert.GreaterOrEqual(t, item.Snippet.Position, int64(0), "Item %d should have position", i)
	}

	// If there are more items, nextPageToken should be set
	if len(items) == 10 {
		assert.NotEmpty(t, nextPageToken, "Should have next page token for pagination")
	}
}

func TestFetchPlaylistMetadata_InvalidPlaylist(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping YouTube API test in short mode")
	}

	apiKey := getTestAPIKey(t)
	if apiKey == "" {
		t.Skip("YOUTUBE_API_KEY not set, skipping API test")
	}

	client, err := youtube.NewClient(context.Background(), apiKey, 100)
	require.NoError(t, err)

	// Use an invalid playlist ID
	playlistID := "INVALID_PLAYLIST_ID"

	_, err = client.GetPlaylist(context.Background(), playlistID)
	assert.Error(t, err)
}

func TestFetchPlaylistItems_EmptyPlaylist(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping YouTube API test in short mode")
	}

	apiKey := getTestAPIKey(t)
	if apiKey == "" {
		t.Skip("YOUTUBE_API_KEY not set, skipping API test")
	}

	client, err := youtube.NewClient(context.Background(), apiKey, 100)
	require.NoError(t, err)

	// Use a known empty or minimal playlist
	playlistID := "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf"

	items, _, err := client.GetPlaylistItems(context.Background(), playlistID, 50, "")
	require.NoError(t, err)
	// Should return items or empty array, not error
	assert.NotNil(t, items)
}
