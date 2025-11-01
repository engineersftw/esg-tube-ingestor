package sync_test

import (
	"testing"
	"time"

	"github.com/engineersftw/youtube-sync/internal/sync"
	"github.com/stretchr/testify/assert"
)

// T071: Unit test for cron schedule parsing
func TestScheduleConfig_ValidCronExpressions(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		valid      bool
	}{
		{name: "every minute", expression: "* * * * *", valid: true},
		{name: "every hour", expression: "0 * * * *", valid: true},
		{name: "every day at midnight", expression: "0 0 * * *", valid: true},
		{name: "every monday at 9am", expression: "0 9 * * 1", valid: true},
		{name: "invalid expression", expression: "invalid", valid: false},
		{name: "empty expression", expression: "", valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sync.ValidateCronExpression(tt.expression)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestScheduleConfig_ParseSchedule(t *testing.T) {
	schedule := &sync.ScheduleConfig{
		ID:         "test-schedule-1",
		Name:       "Daily sync",
		Type:       sync.ScheduleTypeChannel,
		Target:     "UCuAXFkgsw1L7xaCfnd5JJOw",
		CronExpr:   "0 0 * * *",
		Enabled:    true,
		LastRun:    nil,
		NextRun:    nil,
	}

	assert.Equal(t, "test-schedule-1", schedule.ID)
	assert.Equal(t, sync.ScheduleTypeChannel, schedule.Type)
	assert.True(t, schedule.Enabled)
}

// T072: Unit test for scheduled job tracking
func TestScheduler_JobTracking(t *testing.T) {
	job := &sync.ScheduledJob{
		ScheduleID: "test-schedule-1",
		StartTime:  time.Now(),
		Status:     sync.JobStatusRunning,
	}

	assert.Equal(t, "test-schedule-1", job.ScheduleID)
	assert.Equal(t, sync.JobStatusRunning, job.Status)
	assert.False(t, job.StartTime.IsZero())
}

func TestScheduler_JobResult(t *testing.T) {
	result := &sync.JobResult{
		ScheduleID:   "test-schedule-1",
		SuccessCount: 10,
		FailedCount:  2,
		Duration:     5 * time.Minute,
		Error:        nil,
	}

	assert.Equal(t, "test-schedule-1", result.ScheduleID)
	assert.Equal(t, 10, result.SuccessCount)
	assert.Equal(t, 2, result.FailedCount)
	assert.Nil(t, result.Error)
}

func TestScheduleType_Validation(t *testing.T) {
	validTypes := []sync.ScheduleType{
		sync.ScheduleTypeVideo,
		sync.ScheduleTypeChannel,
		sync.ScheduleTypePlaylist,
	}

	for _, scheduleType := range validTypes {
		assert.NotEmpty(t, scheduleType)
	}
}

func TestJobStatus_Validation(t *testing.T) {
	validStatuses := []sync.JobStatus{
		sync.JobStatusPending,
		sync.JobStatusRunning,
		sync.JobStatusCompleted,
		sync.JobStatusFailed,
	}

	for _, status := range validStatuses {
		assert.NotEmpty(t, status)
	}
}

func TestScheduleConfig_WithOptions(t *testing.T) {
	schedule := &sync.ScheduleConfig{
		ID:       "test-schedule-2",
		Name:     "Channel sync with options",
		Type:     sync.ScheduleTypeChannel,
		Target:   "UCuAXFkgsw1L7xaCfnd5JJOw",
		CronExpr: "0 * * * *",
		Enabled:  true,
		Options: map[string]interface{}{
			"limit": 50,
			"force": true,
		},
	}

	assert.NotNil(t, schedule.Options)
	assert.Equal(t, 50, schedule.Options["limit"])
	assert.Equal(t, true, schedule.Options["force"])
}
