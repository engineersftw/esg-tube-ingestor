package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/engineersftw/youtube-sync/internal/database"
	"github.com/engineersftw/youtube-sync/internal/logger"
	"github.com/engineersftw/youtube-sync/internal/parser"
	"github.com/engineersftw/youtube-sync/internal/youtube"
	"go.uber.org/zap"
)

// ChannelSyncOptions contains options for channel sync operations
type ChannelSyncOptions struct {
	Limit   int    // Max number of videos to sync (0 = unlimited)
	Since   string // Only sync videos published after this date (YYYY-MM-DD)
	Force   bool   // Force update even if videos already exist
	DryRun  bool   // Simulate sync without writing to database
	Verbose bool   // Enable verbose logging
}

// ChannelSyncResult contains the results of a channel sync operation
type ChannelSyncResult struct {
	ChannelID    string
	ChannelName  string
	TotalVideos  int
	SuccessCount int
	FailedCount  int
	SkippedCount int
	FailedVideos []FailedVideo
	Duration     time.Duration
}

// FailedVideo represents a video that failed to sync
type FailedVideo struct {
	VideoID string
	Error   string
}

// ChannelSyncer handles syncing YouTube channels to the database
type ChannelSyncer struct {
	ytClient *youtube.Client
	dbClient *database.Client
	logger   *logger.Logger
}

// NewChannelSyncer creates a new channel syncer
func NewChannelSyncer(ytClient *youtube.Client, dbClient *database.Client, log *logger.Logger) *ChannelSyncer {
	return &ChannelSyncer{
		ytClient: ytClient,
		dbClient: dbClient,
		logger:   log,
	}
}

// SyncChannel syncs all videos from a YouTube channel to the database
func (cs *ChannelSyncer) SyncChannel(ctx context.Context, urlOrID string, opts *ChannelSyncOptions) (*ChannelSyncResult, error) {
	startTime := time.Now()

	// Generate correlation ID for tracking
	correlationID := fmt.Sprintf("channel-sync-%d", time.Now().UnixNano())
	log := cs.logger.WithCorrelationID(correlationID)

	// Extract channel ID from URL or validate bare ID
	channelID, err := parser.ExtractChannelID(urlOrID)
	if err != nil {
		log.Error("Invalid channel URL or ID", zap.String("input", urlOrID), zap.Error(err))
		return nil, fmt.Errorf("invalid channel URL or ID: %w", err)
	}

	log.Info("Starting channel sync", zap.String("channel_id", channelID))

	// Fetch channel metadata
	log.Debug("Fetching channel metadata")
	channel, err := cs.ytClient.GetChannel(ctx, channelID)
	if err != nil {
		log.Error("Failed to fetch channel from YouTube", zap.Error(err))
		return nil, fmt.Errorf("failed to fetch channel: %w", err)
	}

	// Log channel info (but don't store it per FR-020)
	channelName := channel.Snippet.Title
	log.Info("Channel metadata fetched",
		zap.String("channel_name", channelName),
		zap.String("channel_id", channel.Id),
	)

	// Initialize result
	result := &ChannelSyncResult{
		ChannelID:    channel.Id,
		ChannelName:  channelName,
		FailedVideos: []FailedVideo{},
	}

	// Fetch videos from channel with pagination
	var allVideos []string // Video IDs
	pageToken := ""
	maxResults := int64(50) // Fetch 50 videos per page

	for {
		log.Debug("Fetching channel videos", zap.String("page_token", pageToken))

		searchResults, nextPageToken, err := cs.ytClient.GetChannelVideos(ctx, channel.Id, maxResults, pageToken)
		if err != nil {
			log.Error("Failed to fetch channel videos", zap.Error(err))
			return nil, fmt.Errorf("failed to fetch channel videos: %w", err)
		}

		// Extract video IDs
		for _, item := range searchResults {
			if item.Id.VideoId != "" {
				allVideos = append(allVideos, item.Id.VideoId)
			}
		}

		log.Debug("Fetched videos page",
			zap.Int("videos_in_page", len(searchResults)),
			zap.Int("total_videos", len(allVideos)),
		)

		// Check if we've reached the limit
		if opts.Limit > 0 && len(allVideos) >= opts.Limit {
			allVideos = allVideos[:opts.Limit]
			log.Info("Reached video limit", zap.Int("limit", opts.Limit))
			break
		}

		// Check if there are more pages
		if nextPageToken == "" {
			break
		}
		pageToken = nextPageToken
	}

	result.TotalVideos = len(allVideos)
	log.Info("Total videos to sync", zap.Int("count", result.TotalVideos))

	// Sync each video
	videoSyncer := NewVideoSyncer(cs.ytClient, cs.dbClient, cs.logger)

	for i, videoID := range allVideos {
		videoLog := log.WithFields(map[string]interface{}{
			"video_id":    videoID,
			"progress":    fmt.Sprintf("%d/%d", i+1, len(allVideos)),
			"correlation": correlationID,
		})

		if opts.Verbose {
			videoLog.Info("Syncing video")
		}

		// Sync the video
		videoOpts := &VideoSyncOptions{
			DryRun:  opts.DryRun,
			Force:   opts.Force,
			Verbose: opts.Verbose,
		}

		_, err := videoSyncer.SyncVideo(ctx, videoID, videoOpts)
		if err != nil {
			// Log error but continue with other videos (per FR-013)
			videoLog.Warn("Failed to sync video", zap.Error(err))
			result.FailedCount++
			result.FailedVideos = append(result.FailedVideos, FailedVideo{
				VideoID: videoID,
				Error:   err.Error(),
			})
			continue
		}

		result.SuccessCount++
	}

	result.Duration = time.Since(startTime)

	log.Info("Channel sync complete",
		zap.Int("total", result.TotalVideos),
		zap.Int("success", result.SuccessCount),
		zap.Int("failed", result.FailedCount),
		zap.Duration("duration", result.Duration),
	)

	return result, nil
}
