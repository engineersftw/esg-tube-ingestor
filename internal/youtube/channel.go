package youtube

import (
	"context"
	"fmt"

	"github.com/cenkalti/backoff/v4"
	"google.golang.org/api/youtube/v3"
	"time"
)

// GetChannel retrieves channel metadata by channel ID with retry logic
func (c *Client) GetChannel(ctx context.Context, channelID string) (*youtube.Channel, error) {
	var channel *youtube.Channel

	operation := func() error {
		// Wait for rate limiter
		if err := c.limiter.Wait(ctx); err != nil {
			return fmt.Errorf("rate limiter error: %w", err)
		}

		// Make API call
		call := c.service.Channels.List([]string{"snippet", "statistics", "contentDetails"}).
			Context(ctx)

		// Handle different channel ID formats
		if len(channelID) == 24 && channelID[:2] == "UC" {
			// Standard channel ID
			call = call.Id(channelID)
		} else {
			// Custom URL or username - use forUsername
			call = call.ForUsername(channelID)
		}

		response, err := call.Do()
		if err != nil {
			return fmt.Errorf("YouTube API error: %w", err)
		}

		if len(response.Items) == 0 {
			return backoff.Permanent(fmt.Errorf("channel not found: %s", channelID))
		}

		channel = response.Items[0]
		return nil
	}

	// Configure backoff
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 30 * time.Second

	if err := backoff.Retry(operation, backoff.WithContext(b, ctx)); err != nil {
		return nil, err
	}

	return channel, nil
}

// GetChannelVideos retrieves videos from a channel with pagination
func (c *Client) GetChannelVideos(ctx context.Context, channelID string, maxResults int64, pageToken string) ([]*youtube.SearchResult, string, error) {
	var videos []*youtube.SearchResult
	var nextPageToken string

	operation := func() error {
		// Wait for rate limiter
		if err := c.limiter.Wait(ctx); err != nil {
			return fmt.Errorf("rate limiter error: %w", err)
		}

		// Search for videos in the channel
		call := c.service.Search.List([]string{"snippet"}).
			ChannelId(channelID).
			Type("video").
			Order("date"). // Most recent first
			MaxResults(maxResults).
			Context(ctx)

		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		response, err := call.Do()
		if err != nil {
			return fmt.Errorf("YouTube API error: %w", err)
		}

		videos = response.Items
		nextPageToken = response.NextPageToken
		return nil
	}

	// Configure backoff
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 30 * time.Second

	if err := backoff.Retry(operation, backoff.WithContext(b, ctx)); err != nil {
		return nil, "", err
	}

	return videos, nextPageToken, nil
}

// GetUploadPlaylistID gets the uploads playlist ID for a channel
// Every channel has an "uploads" playlist containing all their videos
func (c *Client) GetUploadPlaylistID(ctx context.Context, channelID string) (string, error) {
	channel, err := c.GetChannel(ctx, channelID)
	if err != nil {
		return "", err
	}

	if channel.ContentDetails == nil || channel.ContentDetails.RelatedPlaylists == nil {
		return "", fmt.Errorf("channel has no uploads playlist")
	}

	return channel.ContentDetails.RelatedPlaylists.Uploads, nil
}
