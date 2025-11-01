package sync

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/engineersftw/youtube-sync/internal/logger"
	"github.com/robfig/cron/v3"
)

// ScheduleType represents the type of sync to schedule
type ScheduleType string

const (
	ScheduleTypeVideo    ScheduleType = "video"
	ScheduleTypeChannel  ScheduleType = "channel"
	ScheduleTypePlaylist ScheduleType = "playlist"
)

// JobStatus represents the status of a scheduled job
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
)

// ScheduleConfig represents a scheduled sync configuration
type ScheduleConfig struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Type     ScheduleType           `json:"type"`
	Target   string                 `json:"target"` // Video ID, Channel ID, or Playlist ID
	CronExpr string                 `json:"cron_expr"`
	Enabled  bool                   `json:"enabled"`
	Options  map[string]interface{} `json:"options,omitempty"` // force, limit, etc.
	LastRun  *time.Time             `json:"last_run,omitempty"`
	NextRun  *time.Time             `json:"next_run,omitempty"`
}

// ScheduledJob represents a running or completed job
type ScheduledJob struct {
	ScheduleID string
	StartTime  time.Time
	EndTime    *time.Time
	Status     JobStatus
	Error      error
}

// JobResult represents the result of a scheduled job execution
type JobResult struct {
	ScheduleID   string
	SuccessCount int
	FailedCount  int
	Duration     time.Duration
	Error        error
}

// Scheduler manages scheduled sync operations
type Scheduler struct {
	ctx       context.Context
	cron      *cron.Cron
	schedules map[string]*ScheduleConfig
	jobs      map[string]cron.EntryID
	mu        sync.RWMutex
	logger    *logger.Logger
	running   bool
}

// NewScheduler creates a new scheduler
func NewScheduler(ctx context.Context, logger *logger.Logger) *Scheduler {
	return &Scheduler{
		ctx:       ctx,
		cron:      cron.New(),
		schedules: make(map[string]*ScheduleConfig),
		jobs:      make(map[string]cron.EntryID),
		logger:    logger,
		running:   false,
	}
}

// ValidateCronExpression validates a cron expression
func ValidateCronExpression(expression string) error {
	if expression == "" {
		return fmt.Errorf("cron expression cannot be empty")
	}

	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := parser.Parse(expression)
	if err != nil {
		return fmt.Errorf("invalid cron expression: %w", err)
	}

	return nil
}

// AddSchedule adds a new schedule to the scheduler
func (s *Scheduler) AddSchedule(schedule *ScheduleConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate cron expression
	if err := ValidateCronExpression(schedule.CronExpr); err != nil {
		return err
	}

	// Check if schedule already exists
	if _, exists := s.schedules[schedule.ID]; exists {
		return fmt.Errorf("schedule with ID %s already exists", schedule.ID)
	}

	// Store schedule
	s.schedules[schedule.ID] = schedule

	// If scheduler is running and schedule is enabled, add to cron
	if s.running && schedule.Enabled {
		return s.scheduleJob(schedule)
	}

	return nil
}

// RemoveSchedule removes a schedule from the scheduler
func (s *Scheduler) RemoveSchedule(scheduleID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if schedule exists
	if _, exists := s.schedules[scheduleID]; !exists {
		return fmt.Errorf("schedule with ID %s not found", scheduleID)
	}

	// Remove from cron if running
	if entryID, exists := s.jobs[scheduleID]; exists {
		s.cron.Remove(entryID)
		delete(s.jobs, scheduleID)
	}

	// Remove from schedules
	delete(s.schedules, scheduleID)

	return nil
}

// ListSchedules returns all configured schedules
func (s *Scheduler) ListSchedules() []*ScheduleConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	schedules := make([]*ScheduleConfig, 0, len(s.schedules))
	for _, schedule := range s.schedules {
		schedules = append(schedules, schedule)
	}

	return schedules
}

// GetSchedule returns a specific schedule by ID
func (s *Scheduler) GetSchedule(scheduleID string) (*ScheduleConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	schedule, exists := s.schedules[scheduleID]
	if !exists {
		return nil, fmt.Errorf("schedule with ID %s not found", scheduleID)
	}

	return schedule, nil
}

// Start starts the scheduler
func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("scheduler is already running")
	}

	// Add all enabled schedules to cron
	for _, schedule := range s.schedules {
		if schedule.Enabled {
			if err := s.scheduleJob(schedule); err != nil {
				return fmt.Errorf("failed to schedule job %s: %w", schedule.ID, err)
			}
		}
	}

	// Start cron
	s.cron.Start()
	s.running = true

	if s.logger != nil {
		s.logger.WithField("schedule_count", len(s.schedules)).Info("Scheduler started")
	}

	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return fmt.Errorf("scheduler is not running")
	}

	// Stop cron (waits for running jobs to complete)
	ctx := s.cron.Stop()
	<-ctx.Done()

	s.running = false

	if s.logger != nil {
		s.logger.Info("Scheduler stopped")
	}

	return nil
}

// scheduleJob adds a job to the cron scheduler (must be called with lock held)
func (s *Scheduler) scheduleJob(schedule *ScheduleConfig) error {
	// Create job function
	job := func() {
		s.executeJob(schedule)
	}

	// Add to cron
	entryID, err := s.cron.AddFunc(schedule.CronExpr, job)
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	// Store entry ID
	s.jobs[schedule.ID] = entryID

	// Calculate next run time
	entry := s.cron.Entry(entryID)
	nextRun := entry.Next
	schedule.NextRun = &nextRun

	return nil
}

// executeJob executes a scheduled job
func (s *Scheduler) executeJob(schedule *ScheduleConfig) {
	startTime := time.Now()

	if s.logger != nil {
		s.logger.WithFields(map[string]interface{}{
			"schedule_id":   schedule.ID,
			"schedule_name": schedule.Name,
			"schedule_type": schedule.Type,
			"target":        schedule.Target,
		}).Info("Executing scheduled job")
	}

	// Update last run time
	s.mu.Lock()
	schedule.LastRun = &startTime
	s.mu.Unlock()

	// Execute sync based on type
	var err error
	var successCount, failedCount int

	switch schedule.Type {
	case ScheduleTypeVideo:
		// Video sync would be executed here
		// This is a placeholder - actual implementation would call VideoSyncer
		if s.logger != nil {
			s.logger.WithField("video_id", schedule.Target).Info("Video sync scheduled (not implemented)")
		}

	case ScheduleTypeChannel:
		// Channel sync would be executed here
		// This is a placeholder - actual implementation would call ChannelSyncer
		if s.logger != nil {
			s.logger.WithField("channel_id", schedule.Target).Info("Channel sync scheduled (not implemented)")
		}

	case ScheduleTypePlaylist:
		// Playlist sync would be executed here
		// This is a placeholder - actual implementation would call PlaylistSyncer
		if s.logger != nil {
			s.logger.WithField("playlist_id", schedule.Target).Info("Playlist sync scheduled (not implemented)")
		}

	default:
		err = fmt.Errorf("unknown schedule type: %s", schedule.Type)
	}

	duration := time.Since(startTime)

	// Log result
	if s.logger != nil {
		logFields := map[string]interface{}{
			"schedule_id":   schedule.ID,
			"success_count": successCount,
			"failed_count":  failedCount,
			"duration":      duration.String(),
		}

		if err != nil {
			logFields["error"] = err.Error()
			s.logger.WithFields(logFields).Error("Scheduled job failed")
		} else {
			s.logger.WithFields(logFields).Info("Scheduled job completed")
		}
	}
}
