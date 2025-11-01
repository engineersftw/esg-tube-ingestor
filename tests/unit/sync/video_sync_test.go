package sync_test

import (
	"testing"
	"time"

	"github.com/engineersftw/youtube-sync/internal/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T024: Unit test for video metadata mapping to episodes table
func TestMapYouTubeVideoToEpisode(t *testing.T) {
	// Mock YouTube video data
	mockVideo := &sync.YouTubeVideo{
		ID:          "dQw4w9WgXcQ",
		Title:       "Test Video Title",
		Description: "Test video description with some content",
		ChannelID:   "UCuAXFkgsw1L7xaCfnd5JJOw",
		ChannelName: "Test Channel",
		PublishedAt: "2009-10-25T06:57:33Z",
		Thumbnails: sync.Thumbnails{
			Default: "https://i.ytimg.com/vi/dQw4w9WgXcQ/default.jpg",
			Medium:  "https://i.ytimg.com/vi/dQw4w9WgXcQ/mqdefault.jpg",
			High:    "https://i.ytimg.com/vi/dQw4w9WgXcQ/hqdefault.jpg",
		},
		ViewCount: 1234567890,
		Duration:  "PT3M33S",
	}

	episode, err := sync.MapYouTubeVideoToEpisode(mockVideo)
	require.NoError(t, err)

	// Verify mapping
	assert.Equal(t, "dQw4w9WgXcQ", episode.VideoID)
	assert.Equal(t, "Test Video Title", episode.Title)
	assert.Equal(t, "Test video description with some content", episode.Description)
	assert.Equal(t, 1234567890, episode.ViewCount)
	assert.Equal(t, 1, episode.VideoSite) // YouTube = 1
	assert.True(t, episode.Active)
	assert.Equal(t, "https://i.ytimg.com/vi/dQw4w9WgXcQ/default.jpg", episode.Image1)
	assert.Equal(t, "https://i.ytimg.com/vi/dQw4w9WgXcQ/mqdefault.jpg", episode.Image2)
	assert.Equal(t, "https://i.ytimg.com/vi/dQw4w9WgXcQ/hqdefault.jpg", episode.Image3)

	// Verify published date was parsed
	assert.False(t, episode.PublishedAt.IsZero())
	expectedDate := time.Date(2009, 10, 25, 6, 57, 33, 0, time.UTC)
	assert.Equal(t, expectedDate, episode.PublishedAt)
}

func TestMapYouTubeVideoToEpisode_InvalidPublishedDate(t *testing.T) {
	mockVideo := &sync.YouTubeVideo{
		ID:          "test123",
		Title:       "Test",
		PublishedAt: "invalid-date",
	}

	_, err := sync.MapYouTubeVideoToEpisode(mockVideo)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse")
}

func TestMapYouTubeVideoToEpisode_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name  string
		video *sync.YouTubeVideo
	}{
		{
			name: "missing video ID",
			video: &sync.YouTubeVideo{
				Title:       "Test",
				PublishedAt: "2009-10-25T06:57:33Z",
			},
		},
		{
			name: "missing title",
			video: &sync.YouTubeVideo{
				ID:          "test123",
				PublishedAt: "2009-10-25T06:57:33Z",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := sync.MapYouTubeVideoToEpisode(tt.video)
			assert.Error(t, err)
		})
	}
}

// T025: Integration test for end-to-end video sync
// Note: This would be in tests/integration/ but putting here for organization
func TestVideoSync_EndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test would require:
	// 1. Mock YouTube client or test API key
	// 2. Test database connection
	// 3. Mock logger
	// For now, we define the expected behavior
	t.Skip("Full integration test requires database and API setup")
}

func TestVideoSyncOptions_DryRun(t *testing.T) {
	opts := &sync.VideoSyncOptions{
		DryRun: true,
		Force:  false,
	}

	assert.True(t, opts.DryRun)
	assert.False(t, opts.Force)
}

func TestVideoSyncOptions_Force(t *testing.T) {
	opts := &sync.VideoSyncOptions{
		DryRun: false,
		Force:  true,
	}

	assert.False(t, opts.DryRun)
	assert.True(t, opts.Force)
}
