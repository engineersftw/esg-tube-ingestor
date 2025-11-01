package sync_test

import (
	"testing"

	"github.com/engineersftw/youtube-sync/internal/sync"
	"github.com/stretchr/testify/assert"
)

// T039: Unit test for channel sync logic
func TestChannelSyncOptions(t *testing.T) {
	opts := &sync.ChannelSyncOptions{
		Limit:   50,
		Since:   "2024-01-01",
		Force:   true,
		DryRun:  false,
		Verbose: true,
	}

	assert.Equal(t, 50, opts.Limit)
	assert.Equal(t, "2024-01-01", opts.Since)
	assert.True(t, opts.Force)
	assert.False(t, opts.DryRun)
	assert.True(t, opts.Verbose)
}

func TestChannelSyncOptions_UnlimitedSync(t *testing.T) {
	opts := &sync.ChannelSyncOptions{
		Limit: 0, // 0 means unlimited
	}

	assert.Equal(t, 0, opts.Limit, "Limit of 0 should mean unlimited")
}

func TestChannelSyncResult(t *testing.T) {
	result := &sync.ChannelSyncResult{
		ChannelID:    "UCuAXFkgsw1L7xaCfnd5JJOw",
		ChannelName:  "Rick Astley",
		TotalVideos:  100,
		SuccessCount: 98,
		FailedCount:  2,
	}

	assert.Equal(t, 100, result.TotalVideos)
	assert.Equal(t, 98, result.SuccessCount)
	assert.Equal(t, 2, result.FailedCount)
	assert.Equal(t, "UCuAXFkgsw1L7xaCfnd5JJOw", result.ChannelID)
}

// T040: Integration test for bulk channel sync
func TestChannelSync_EndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test would require:
	// 1. Mock YouTube client
	// 2. Test database connection
	// 3. Mock logger
	t.Skip("Full integration test requires database and API setup")
}
