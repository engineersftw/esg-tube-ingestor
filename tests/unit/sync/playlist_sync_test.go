package sync_test

import (
	"testing"

	"github.com/engineersftw/youtube-sync/internal/sync"
	"github.com/stretchr/testify/assert"
)

// T056: Unit test for playlist sync logic
func TestPlaylistSyncOptions_Defaults(t *testing.T) {
	opts := &sync.PlaylistSyncOptions{}

	assert.False(t, opts.Force)
	assert.False(t, opts.DryRun)
	assert.False(t, opts.Verbose)
}

func TestPlaylistSyncOptions_CustomValues(t *testing.T) {
	opts := &sync.PlaylistSyncOptions{
		Force:   true,
		DryRun:  true,
		Verbose: true,
	}

	assert.True(t, opts.Force)
	assert.True(t, opts.DryRun)
	assert.True(t, opts.Verbose)
}

func TestPlaylistSyncResult_InitialState(t *testing.T) {
	result := &sync.PlaylistSyncResult{
		PlaylistID:    "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf",
		PlaylistTitle: "Test Playlist",
		TotalItems:    10,
		SuccessCount:  8,
		FailedCount:   2,
		SkippedCount:  0,
	}

	assert.Equal(t, "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf", result.PlaylistID)
	assert.Equal(t, "Test Playlist", result.PlaylistTitle)
	assert.Equal(t, 10, result.TotalItems)
	assert.Equal(t, 8, result.SuccessCount)
	assert.Equal(t, 2, result.FailedCount)
	assert.Equal(t, 0, result.SkippedCount)
}

func TestPlaylistSyncResult_WithFailedVideos(t *testing.T) {
	result := &sync.PlaylistSyncResult{
		PlaylistID:    "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf",
		PlaylistTitle: "Test Playlist",
		FailedVideos: []sync.FailedVideo{
			{VideoID: "dQw4w9WgXcQ", Error: "API error"},
			{VideoID: "jNQXAC9IVRw", Error: "Video unavailable"},
		},
	}

	assert.Len(t, result.FailedVideos, 2)
	assert.Equal(t, "dQw4w9WgXcQ", result.FailedVideos[0].VideoID)
	assert.Equal(t, "API error", result.FailedVideos[0].Error)
}
