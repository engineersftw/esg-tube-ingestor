package parser

import (
	"fmt"
	"strings"
)

// ValidateYouTubeURL performs comprehensive validation of YouTube URLs
func ValidateYouTubeURL(input string) error {
	if input == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	input = strings.TrimSpace(input)

	// Try to extract video ID - if this succeeds, URL is valid
	_, err := ExtractVideoID(input)
	if err != nil {
		return fmt.Errorf("invalid YouTube URL: %w", err)
	}

	return nil
}

// ValidateVideoIDFormat validates just the format of a video ID
func ValidateVideoIDFormat(videoID string) error {
	if videoID == "" {
		return fmt.Errorf("video ID cannot be empty")
	}

	if !IsValidVideoID(videoID) {
		return fmt.Errorf("invalid video ID format: must be 11 characters (alphanumeric, underscore, or hyphen)")
	}

	return nil
}
