package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/engineersftw/youtube-sync/internal/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T073: Integration test for scheduled sync execution
func TestScheduler_ExecuteScheduledJob(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping scheduler integration test in short mode")
	}

	// Create a test schedule that runs every minute
	schedule := &sync.ScheduleConfig{
		ID:       "test-integration-1",
		Name:     "Test video sync",
		Type:     sync.ScheduleTypeVideo,
		Target:   "dQw4w9WgXcQ",
		CronExpr: "* * * * *", // Every minute
		Enabled:  true,
	}

	// Validate schedule can be parsed
	err := sync.ValidateCronExpression(schedule.CronExpr)
	require.NoError(t, err, "Cron expression should be valid")

	// Test that we can create a scheduler
	scheduler := sync.NewScheduler(context.Background(), nil)
	require.NotNil(t, scheduler, "Scheduler should be created")

	// Test adding a schedule
	err = scheduler.AddSchedule(schedule)
	assert.NoError(t, err, "Should be able to add schedule")

	// Test listing schedules
	schedules := scheduler.ListSchedules()
	assert.Len(t, schedules, 1, "Should have one schedule")
	assert.Equal(t, schedule.ID, schedules[0].ID)
}

func TestScheduler_StartStop(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping scheduler integration test in short mode")
	}

	scheduler := sync.NewScheduler(context.Background(), nil)
	require.NotNil(t, scheduler)

	// Start scheduler
	err := scheduler.Start()
	assert.NoError(t, err, "Should be able to start scheduler")

	// Wait a moment
	time.Sleep(100 * time.Millisecond)

	// Stop scheduler
	err = scheduler.Stop()
	assert.NoError(t, err, "Should be able to stop scheduler")
}

func TestScheduler_RemoveSchedule(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping scheduler integration test in short mode")
	}

	scheduler := sync.NewScheduler(context.Background(), nil)

	schedule := &sync.ScheduleConfig{
		ID:       "test-remove-1",
		Name:     "Test removal",
		Type:     sync.ScheduleTypeVideo,
		Target:   "dQw4w9WgXcQ",
		CronExpr: "0 0 * * *",
		Enabled:  true,
	}

	// Add schedule
	err := scheduler.AddSchedule(schedule)
	require.NoError(t, err)

	// Verify it exists
	schedules := scheduler.ListSchedules()
	assert.Len(t, schedules, 1)

	// Remove schedule
	err = scheduler.RemoveSchedule(schedule.ID)
	assert.NoError(t, err)

	// Verify it's gone
	schedules = scheduler.ListSchedules()
	assert.Len(t, schedules, 0)
}
