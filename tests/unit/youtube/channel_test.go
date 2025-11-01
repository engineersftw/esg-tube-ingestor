package youtube_test

import (
	"context"
	"testing"

	"github.com/engineersftw/youtube-sync/internal/youtube"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T037: Unit test for channel metadata fetching
func TestFetchChannelMetadata_ValidChannel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping YouTube API test in short mode")
	}

	apiKey := getTestAPIKey(t)
	if apiKey == "" {
		t.Skip("YOUTUBE_API_KEY not set, skipping API test")
	}

	client, err := youtube.NewClient(context.Background(), apiKey, 100)
	require.NoError(t, err)

	// Use a stable, well-known channel
	channelID := "UCuAXFkgsw1L7xaCfnd5JJOw" // Rick Astley

	channel, err := client.GetChannel(context.Background(), channelID)
	require.NoError(t, err)
	require.NotNil(t, channel)

	// Verify channel metadata
	assert.Equal(t, channelID, channel.Id)
	assert.NotEmpty(t, channel.Snippet.Title)
	assert.NotEmpty(t, channel.Snippet.Description)
}

// T038: Unit test for pagination handling
func TestFetchChannelVideos_WithPagination(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping YouTube API test in short mode")
	}

	apiKey := getTestAPIKey(t)
	if apiKey == "" {
		t.Skip("YOUTUBE_API_KEY not set, skipping API test")
	}

	client, err := youtube.NewClient(context.Background(), apiKey, 100)
	require.NoError(t, err)

	channelID := "UCuAXFkgsw1L7xaCfnd5JJOw"

	// Fetch videos with limit
	videos, nextPageToken, err := client.GetChannelVideos(context.Background(), channelID, 10, "")
	require.NoError(t, err)
	assert.LessOrEqual(t, len(videos), 10)

	// If there are more videos, nextPageToken should be set
	if len(videos) == 10 {
		assert.NotEmpty(t, nextPageToken, "Should have next page token for pagination")
	}
}

func TestFetchChannelVideos_InvalidChannel(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping YouTube API test in short mode")
	}

	apiKey := getTestAPIKey(t)
	if apiKey == "" {
		t.Skip("YOUTUBE_API_KEY not set, skipping API test")
	}

	client, err := youtube.NewClient(context.Background(), apiKey, 100)
	require.NoError(t, err)

	// Use an invalid channel ID
	channelID := "INVALID_CHANNEL_ID"

	_, _, err = client.GetChannelVideos(context.Background(), channelID, 10, "")
	assert.Error(t, err)
}
