package youtube

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"golang.org/x/time/rate"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// Client wraps the YouTube API client with rate limiting and retry logic
type Client struct {
	service *youtube.Service
	limiter *rate.Limiter
}

// NewClient creates a new YouTube API client with rate limiting
func NewClient(ctx context.Context, apiKey string, requestsPerMinute int) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("YouTube API key is required")
	}

	// Create YouTube service
	service, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create YouTube service: %w", err)
	}

	// Create rate limiter (requests per minute converted to requests per second)
	rps := float64(requestsPerMinute) / 60.0
	limiter := rate.NewLimiter(rate.Limit(rps), 1)

	return &Client{
		service: service,
		limiter: limiter,
	}, nil
}

// GetVideo retrieves video metadata by video ID with retry logic
func (c *Client) GetVideo(ctx context.Context, videoID string) (*youtube.Video, error) {
	var video *youtube.Video

	operation := func() error {
		// Wait for rate limiter
		if err := c.limiter.Wait(ctx); err != nil {
			return fmt.Errorf("rate limiter error: %w", err)
		}

		// Make API call
		call := c.service.Videos.List([]string{"snippet", "statistics", "contentDetails"}).
			Id(videoID).
			Context(ctx)

		response, err := call.Do()
		if err != nil {
			return fmt.Errorf("YouTube API error: %w", err)
		}

		if len(response.Items) == 0 {
			return backoff.Permanent(fmt.Errorf("video not found: %s", videoID))
		}

		video = response.Items[0]
		return nil
	}

	// Configure backoff
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 30 * time.Second

	if err := backoff.Retry(operation, backoff.WithContext(b, ctx)); err != nil {
		return nil, err
	}

	return video, nil
}

// Service returns the underlying YouTube service for advanced usage
func (c *Client) Service() *youtube.Service {
	return c.service
}
