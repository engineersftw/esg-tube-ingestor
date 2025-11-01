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
	ytapi "google.golang.org/api/youtube/v3"
)

// YouTubeVideo represents simplified YouTube video metadata
type YouTubeVideo struct {
	ID          string
	Title       string
	Description string
	ChannelID   string
	ChannelName string
	PublishedAt string
	Thumbnails  Thumbnails
	ViewCount   int
	Duration    string
}

// Thumbnails represents video thumbnail URLs
type Thumbnails struct {
	Default string
	Medium  string
	High    string
}

// VideoSyncOptions contains options for video sync operations
type VideoSyncOptions struct {
	DryRun  bool
	Force   bool
	Verbose bool
}

// VideoSyncer handles syncing YouTube videos to the database
type VideoSyncer struct {
	ytClient *youtube.Client
	dbClient *database.Client
	logger   *logger.Logger
}

// NewVideoSyncer creates a new video syncer
func NewVideoSyncer(ytClient *youtube.Client, dbClient *database.Client, log *logger.Logger) *VideoSyncer {
	return &VideoSyncer{
		ytClient: ytClient,
		dbClient: dbClient,
		logger:   log,
	}
}

// SyncVideo syncs a single YouTube video to the database
func (vs *VideoSyncer) SyncVideo(ctx context.Context, urlOrID string, opts *VideoSyncOptions) (*database.Episode, error) {
	// Generate correlation ID for tracking
	correlationID := fmt.Sprintf("video-sync-%d", time.Now().UnixNano())
	log := vs.logger.WithCorrelationID(correlationID)

	// Extract video ID from URL or validate bare ID
	videoID, err := parser.ExtractVideoID(urlOrID)
	if err != nil {
		log.Error("Invalid video URL or ID", zap.String("input", urlOrID), zap.Error(err))
		return nil, fmt.Errorf("invalid video URL or ID: %w", err)
	}

	log.Info("Starting video sync", zap.String("video_id", videoID))

	// Check if video already exists (unless force flag is set)
	if !opts.Force {
		existing, err := vs.dbClient.GetEpisodeByVideoID(ctx, videoID)
		if err == nil && existing != nil {
			log.Info("Video already exists", zap.Int("database_id", existing.ID))
			if !opts.DryRun {
				return existing, nil
			}
		}
	}

	// Fetch video metadata from YouTube API
	log.Debug("Fetching video metadata from YouTube API")
	ytVideo, err := vs.ytClient.GetVideo(ctx, videoID)
	if err != nil {
		log.Error("Failed to fetch video from YouTube", zap.Error(err))
		return nil, fmt.Errorf("failed to fetch video: %w", err)
	}

	// Convert YouTube API response to our simplified format
	video := convertYouTubeAPIVideo(ytVideo)

	// Map to database episode
	episode, err := MapYouTubeVideoToEpisode(video)
	if err != nil {
		log.Error("Failed to map video to episode", zap.Error(err))
		return nil, fmt.Errorf("failed to map video: %w", err)
	}

	// Log channel info (but don't store it per FR-020)
	log.Info("Video metadata fetched",
		zap.String("title", episode.Title),
		zap.String("channel", video.ChannelName),
		zap.Int("views", episode.ViewCount),
	)

	// Dry run - don't write to database
	if opts.DryRun {
		log.Info("Dry run - skipping database write")
		return episode, nil
	}

	// Upsert episode to database
	log.Debug("Upserting episode to database")
	err = vs.dbClient.UpsertEpisode(ctx, episode)
	if err != nil {
		log.Error("Failed to upsert episode", zap.Error(err))
		return nil, fmt.Errorf("failed to save episode: %w", err)
	}

	log.Info("Video sync complete",
		zap.Int("database_id", episode.ID),
		zap.Bool("created", episode.CreatedAt.Equal(episode.UpdatedAt)),
	)

	return episode, nil
}

// MapYouTubeVideoToEpisode maps a YouTube video to a database episode
func MapYouTubeVideoToEpisode(video *YouTubeVideo) (*database.Episode, error) {
	// Validate required fields
	if video.ID == "" {
		return nil, fmt.Errorf("video ID is required")
	}
	if video.Title == "" {
		return nil, fmt.Errorf("video title is required")
	}

	// Parse published date
	publishedAt, err := time.Parse(time.RFC3339, video.PublishedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse published date: %w", err)
	}

	episode := &database.Episode{
		VideoID:     video.ID,
		Title:       video.Title,
		Description: video.Description,
		PublishedAt: publishedAt,
		Image1:      video.Thumbnails.Default,
		Image2:      video.Thumbnails.Medium,
		Image3:      video.Thumbnails.High,
		ViewCount:   video.ViewCount,
		VideoSite:   1, // YouTube = 1
		Active:      true,
	}

	return episode, nil
}

// convertYouTubeAPIVideo converts YouTube API video to our simplified format
func convertYouTubeAPIVideo(ytVideo *ytapi.Video) *YouTubeVideo {
	video := &YouTubeVideo{
		ID:          ytVideo.Id,
		Title:       ytVideo.Snippet.Title,
		Description: ytVideo.Snippet.Description,
		ChannelID:   ytVideo.Snippet.ChannelId,
		ChannelName: ytVideo.Snippet.ChannelTitle,
		PublishedAt: ytVideo.Snippet.PublishedAt,
		Duration:    ytVideo.ContentDetails.Duration,
	}

	// Extract thumbnails
	if ytVideo.Snippet.Thumbnails != nil {
		if ytVideo.Snippet.Thumbnails.Default != nil {
			video.Thumbnails.Default = ytVideo.Snippet.Thumbnails.Default.Url
		}
		if ytVideo.Snippet.Thumbnails.Medium != nil {
			video.Thumbnails.Medium = ytVideo.Snippet.Thumbnails.Medium.Url
		}
		if ytVideo.Snippet.Thumbnails.High != nil {
			video.Thumbnails.High = ytVideo.Snippet.Thumbnails.High.Url
		}
	}

	// Extract view count
	if ytVideo.Statistics != nil {
		video.ViewCount = int(ytVideo.Statistics.ViewCount)
	}

	return video
}
