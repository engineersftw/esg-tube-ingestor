package parser

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// VideoID patterns for YouTube
var (
	videoIDRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{11}$`)
)

// ExtractVideoID extracts a YouTube video ID from various URL formats or bare ID
func ExtractVideoID(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("invalid video URL or ID: empty string")
	}

	// Trim whitespace
	input = strings.TrimSpace(input)

	// Check if it's already a valid video ID
	if IsValidVideoID(input) {
		return input, nil
	}

	// Try to parse as URL
	parsedURL, err := url.Parse(input)
	if err != nil {
		return "", fmt.Errorf("invalid video URL: %w", err)
	}

	// Extract video ID based on URL format
	videoID := ""

	if strings.Contains(parsedURL.Host, "youtu.be") {
		// Short URL format: https://youtu.be/VIDEO_ID
		videoID = strings.TrimPrefix(parsedURL.Path, "/")
	} else if strings.Contains(parsedURL.Host, "youtube.com") {
		// Check for different YouTube URL formats
		if strings.Contains(parsedURL.Path, "/watch") {
			// Standard format: https://www.youtube.com/watch?v=VIDEO_ID
			videoID = parsedURL.Query().Get("v")
		} else if strings.Contains(parsedURL.Path, "/embed/") {
			// Embed format: https://www.youtube.com/embed/VIDEO_ID
			parts := strings.Split(parsedURL.Path, "/")
			for i, part := range parts {
				if part == "embed" && i+1 < len(parts) {
					videoID = parts[i+1]
					break
				}
			}
		} else if strings.Contains(parsedURL.Path, "/v/") {
			// Old format: https://www.youtube.com/v/VIDEO_ID
			parts := strings.Split(parsedURL.Path, "/")
			for i, part := range parts {
				if part == "v" && i+1 < len(parts) {
					videoID = parts[i+1]
					break
				}
			}
		}
	} else {
		return "", fmt.Errorf("invalid video URL: not a YouTube URL")
	}

	// Remove any query parameters from videoID (e.g., ?t=123)
	if idx := strings.Index(videoID, "?"); idx != -1 {
		videoID = videoID[:idx]
	}

	// Validate extracted video ID
	if !IsValidVideoID(videoID) {
		return "", fmt.Errorf("invalid video ID extracted: %s", videoID)
	}

	return videoID, nil
}

// ValidateVideoURL validates that a string is a valid YouTube video URL or ID
func ValidateVideoURL(input string) error {
	_, err := ExtractVideoID(input)
	return err
}

// IsValidVideoID checks if a string is a valid YouTube video ID
// YouTube video IDs are exactly 11 characters: alphanumeric, underscore, or hyphen
func IsValidVideoID(id string) bool {
	return videoIDRegex.MatchString(id)
}

// ExtractChannelID extracts a YouTube channel ID from various URL formats or bare ID
func ExtractChannelID(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("invalid channel URL or ID: empty string")
	}

	// Trim whitespace
	input = strings.TrimSpace(input)

	// Check if it's already a valid channel ID
	if IsValidChannelID(input) {
		return input, nil
	}

	// Try to parse as URL
	parsedURL, err := url.Parse(input)
	if err != nil {
		return "", fmt.Errorf("invalid channel URL: %w", err)
	}

	// Extract channel ID based on URL format
	channelID := ""

	if strings.Contains(parsedURL.Host, "youtube.com") {
		// Check for different YouTube channel URL formats
		if strings.Contains(parsedURL.Path, "/channel/") {
			// Format: https://www.youtube.com/channel/CHANNEL_ID
			parts := strings.Split(parsedURL.Path, "/")
			for i, part := range parts {
				if part == "channel" && i+1 < len(parts) {
					channelID = parts[i+1]
					break
				}
			}
		} else if strings.Contains(parsedURL.Path, "/c/") {
			// Format: https://www.youtube.com/c/CustomName
			parts := strings.Split(parsedURL.Path, "/")
			for i, part := range parts {
				if part == "c" && i+1 < len(parts) {
					channelID = parts[i+1]
					break
				}
			}
		} else if strings.Contains(parsedURL.Path, "/user/") {
			// Format: https://www.youtube.com/user/Username
			parts := strings.Split(parsedURL.Path, "/")
			for i, part := range parts {
				if part == "user" && i+1 < len(parts) {
					channelID = parts[i+1]
					break
				}
			}
		} else if strings.HasPrefix(parsedURL.Path, "/@") {
			// Format: https://www.youtube.com/@Handle
			channelID = strings.TrimPrefix(parsedURL.Path, "/")
		}
	} else {
		return "", fmt.Errorf("invalid channel URL: not a YouTube URL")
	}

	// Validate extracted channel ID
	if !IsValidChannelID(channelID) {
		return "", fmt.Errorf("invalid channel ID extracted: %s", channelID)
	}

	return channelID, nil
}

// IsValidChannelID checks if a string is a valid YouTube channel ID or custom name
// Channel IDs typically start with "UC" and are 24 characters
// Custom names and handles can be shorter and contain alphanumeric, underscore, hyphen, or @
func IsValidChannelID(id string) bool {
	if id == "" || len(id) < 3 {
		return false
	}

	// Allow UC channel IDs (24 chars), custom names, and @ handles
	validPattern := regexp.MustCompile(`^(UC[a-zA-Z0-9_-]{22}|@?[a-zA-Z0-9_-]{3,})$`)
	return validPattern.MatchString(id)
}

// ExtractPlaylistID extracts a YouTube playlist ID from various URL formats or bare ID
func ExtractPlaylistID(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("invalid playlist URL or ID: empty string")
	}

	// Trim whitespace
	input = strings.TrimSpace(input)

	// Check if it's already a valid playlist ID
	if IsValidPlaylistID(input) {
		return input, nil
	}

	// Try to parse as URL
	parsedURL, err := url.Parse(input)
	if err != nil {
		return "", fmt.Errorf("invalid playlist URL: %w", err)
	}

	// Extract playlist ID from URL
	playlistID := ""

	if strings.Contains(parsedURL.Host, "youtube.com") || strings.Contains(parsedURL.Host, "youtu.be") {
		// Check query parameter "list"
		playlistID = parsedURL.Query().Get("list")
	} else {
		return "", fmt.Errorf("invalid playlist URL: not a YouTube URL")
	}

	// Validate extracted playlist ID
	if !IsValidPlaylistID(playlistID) {
		return "", fmt.Errorf("invalid playlist ID extracted: %s", playlistID)
	}

	return playlistID, nil
}

// IsValidPlaylistID checks if a string is a valid YouTube playlist ID
// Playlist IDs typically start with PL, UU, LL, or FL and are 24+ characters
func IsValidPlaylistID(id string) bool {
	if id == "" || len(id) < 10 {
		return false
	}

	// Allow PL, UU, LL, FL, RD, OL prefixes followed by alphanumeric, underscore, or hyphen
	validPattern := regexp.MustCompile(`^(PL|UU|LL|FL|RD|OL)[a-zA-Z0-9_-]{10,}$`)
	return validPattern.MatchString(id)
}

// GenerateSlug generates a URL-friendly slug from a title
func GenerateSlug(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)

	// Replace spaces and special characters with hyphens
	slug = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(slug, "-")

	// Remove leading/trailing hyphens
	slug = strings.Trim(slug, "-")

	// Remove duplicate hyphens
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")

	return slug
}
