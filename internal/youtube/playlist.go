package youtube

import (
	"context"
	"fmt"

	"google.golang.org/api/youtube/v3"
)

// GetPlaylist fetches playlist metadata from YouTube API
func (c *Client) GetPlaylist(ctx context.Context, playlistID string) (*youtube.Playlist, error) {
	// Wait for rate limiter
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter error: %w", err)
	}

	// Fetch playlist metadata
	call := c.service.Playlists.List([]string{"snippet", "contentDetails"}).
		Id(playlistID).
		Context(ctx)

	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch playlist: %w", err)
	}

	if len(response.Items) == 0 {
		return nil, fmt.Errorf("playlist not found: %s", playlistID)
	}

	return response.Items[0], nil
}

// GetPlaylistItems fetches playlist items (videos) with pagination support
func (c *Client) GetPlaylistItems(ctx context.Context, playlistID string, maxResults int64, pageToken string) ([]*youtube.PlaylistItem, string, error) {
	// Wait for rate limiter
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, "", fmt.Errorf("rate limiter error: %w", err)
	}

	// Fetch playlist items
	call := c.service.PlaylistItems.List([]string{"snippet", "contentDetails"}).
		PlaylistId(playlistID).
		MaxResults(maxResults).
		Context(ctx)

	if pageToken != "" {
		call = call.PageToken(pageToken)
	}

	response, err := call.Do()
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch playlist items: %w", err)
	}

	return response.Items, response.NextPageToken, nil
}
