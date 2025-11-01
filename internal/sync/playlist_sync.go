package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/engineersftw/youtube-sync/internal/database"
	"github.com/engineersftw/youtube-sync/internal/logger"
	"github.com/engineersftw/youtube-sync/internal/parser"
	"github.com/engineersftw/youtube-sync/internal/youtube"
)

// PlaylistSyncOptions contains options for playlist sync operations
type PlaylistSyncOptions struct {
	Force   bool
	DryRun  bool
	Verbose bool
}

// PlaylistSyncResult contains the result of a playlist sync operation
type PlaylistSyncResult struct {
	PlaylistID    string
	PlaylistTitle string
	TotalItems    int
	SuccessCount  int
	FailedCount   int
	SkippedCount  int
	Duration      time.Duration
	FailedVideos  []FailedVideo
}

// PlaylistSyncer handles syncing YouTube playlists
type PlaylistSyncer struct {
	ytClient *youtube.Client
	dbClient *database.Client
	logger   *logger.Logger
}

// NewPlaylistSyncer creates a new playlist syncer
func NewPlaylistSyncer(ytClient *youtube.Client, dbClient *database.Client, logger *logger.Logger) *PlaylistSyncer {
	return &PlaylistSyncer{
		ytClient: ytClient,
		dbClient: dbClient,
		logger:   logger,
	}
}

// SyncPlaylist syncs a YouTube playlist with all its videos to the database
func (ps *PlaylistSyncer) SyncPlaylist(ctx context.Context, urlOrID string, opts *PlaylistSyncOptions) (*PlaylistSyncResult, error) {
	startTime := time.Now()
	correlationID := fmt.Sprintf("playlist-sync-%d", startTime.Unix())
	log := ps.logger.WithCorrelationID(correlationID)

	// Extract playlist ID from URL or validate bare ID
	playlistID, err := parser.ExtractPlaylistID(urlOrID)
	if err != nil {
		return nil, fmt.Errorf("invalid playlist URL or ID: %w", err)
	}

	log.WithFields(map[string]interface{}{
		"playlist_id": playlistID,
		"dry_run":     opts.DryRun,
		"force":       opts.Force,
	}).Info("Starting playlist sync")

	// Fetch playlist metadata
	playlistMeta, err := ps.ytClient.GetPlaylist(ctx, playlistID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch playlist metadata: %w", err)
	}

	log.WithFields(map[string]interface{}{
		"playlist_title": playlistMeta.Snippet.Title,
		"item_count":     playlistMeta.ContentDetails.ItemCount,
	}).Info("Fetched playlist metadata")

	// Initialize result
	result := &PlaylistSyncResult{
		PlaylistID:    playlistID,
		PlaylistTitle: playlistMeta.Snippet.Title,
		FailedVideos:  []FailedVideo{},
	}

	if opts.DryRun {
		log.Info("Dry run mode - skipping database operations")
		result.Duration = time.Since(startTime)
		return result, nil
	}

	// Create/update playlist in database
	dbPlaylist := &database.Playlist{
		PlaylistID:  playlistID,
		Title:       playlistMeta.Snippet.Title,
		Description: playlistMeta.Snippet.Description,
		Slug:        parser.GenerateSlug(playlistMeta.Snippet.Title),
		Active:      true,
	}

	// Get thumbnail URLs
	if playlistMeta.Snippet.Thumbnails != nil {
		if playlistMeta.Snippet.Thumbnails.High != nil {
			dbPlaylist.Image1 = playlistMeta.Snippet.Thumbnails.High.Url
		}
		if playlistMeta.Snippet.Thumbnails.Medium != nil {
			dbPlaylist.Image2 = playlistMeta.Snippet.Thumbnails.Medium.Url
		}
		if playlistMeta.Snippet.Thumbnails.Default != nil {
			dbPlaylist.Image3 = playlistMeta.Snippet.Thumbnails.Default.Url
		}
	}

	// Upsert playlist
	err = ps.dbClient.UpsertPlaylist(ctx, dbPlaylist)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert playlist: %w", err)
	}

	log.WithField("playlist_db_id", dbPlaylist.ID).Info("Playlist upserted")

	// Clear existing playlist items (to handle deletions/reorderings)
	err = ps.dbClient.DeletePlaylistItems(ctx, dbPlaylist.ID)
	if err != nil {
		log.WithField("error", err.Error()).Warn("Failed to clear playlist items")
	}

	// Fetch all playlist items with pagination
	pageToken := ""
	position := 0

	for {
		items, nextPageToken, err := ps.ytClient.GetPlaylistItems(ctx, playlistID, 50, pageToken)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch playlist items: %w", err)
		}

		log.WithField("items_count", len(items)).Debug("Fetched playlist items page")

		// Sync each video in the playlist
		for _, item := range items {
			videoID := item.Snippet.ResourceId.VideoId
			result.TotalItems++

			// Sync the video first (ensure it exists in episodes table)
			videoSyncer := NewVideoSyncer(ps.ytClient, ps.dbClient, ps.logger)
			videoOpts := &VideoSyncOptions{
				Force:   opts.Force,
				DryRun:  false,
				Verbose: opts.Verbose,
			}

			episode, err := videoSyncer.SyncVideo(ctx, videoID, videoOpts)
			if err != nil {
				log.WithFields(map[string]interface{}{
					"video_id": videoID,
					"error":    err.Error(),
				}).Warn("Failed to sync video in playlist")
				result.FailedCount++
				result.FailedVideos = append(result.FailedVideos, FailedVideo{
					VideoID: videoID,
					Error:   err.Error(),
				})
				continue
			}

			// Add to playlist_items
			playlistItem := &database.PlaylistItem{
				PlaylistID: dbPlaylist.ID,
				EpisodeID:  int64(episode.ID),
				SortOrder:  position,
			}

			err = ps.dbClient.UpsertPlaylistItem(ctx, playlistItem)
			if err != nil {
				log.WithFields(map[string]interface{}{
					"video_id": videoID,
					"error":    err.Error(),
				}).Warn("Failed to add video to playlist_items")
				result.FailedCount++
				continue
			}

			result.SuccessCount++
			position++

			if opts.Verbose {
				log.WithFields(map[string]interface{}{
					"video_id": videoID,
					"title":    episode.Title,
					"position": position,
				}).Debug("Synced playlist video")
			}
		}

		pageToken = nextPageToken
		if pageToken == "" {
			break
		}
	}

	result.Duration = time.Since(startTime)

	log.WithFields(map[string]interface{}{
		"total_items":   result.TotalItems,
		"success_count": result.SuccessCount,
		"failed_count":  result.FailedCount,
		"duration":      result.Duration.String(),
	}).Info("Playlist sync complete")

	return result, nil
}
