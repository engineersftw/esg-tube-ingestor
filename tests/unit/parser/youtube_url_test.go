package parser_test

import (
	"testing"

	"github.com/engineersftw/youtube-sync/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T021: Unit test for YouTube video ID extraction from URLs
func TestExtractVideoID_ValidURLs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard watch URL",
			input:    "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			expected: "dQw4w9WgXcQ",
		},
		{
			name:     "short youtu.be URL",
			input:    "https://youtu.be/dQw4w9WgXcQ",
			expected: "dQw4w9WgXcQ",
		},
		{
			name:     "embed URL",
			input:    "https://www.youtube.com/embed/dQw4w9WgXcQ",
			expected: "dQw4w9WgXcQ",
		},
		{
			name:     "watch URL with timestamp",
			input:    "https://www.youtube.com/watch?v=dQw4w9WgXcQ&t=42s",
			expected: "dQw4w9WgXcQ",
		},
		{
			name:     "short URL with query params",
			input:    "https://youtu.be/dQw4w9WgXcQ?t=123",
			expected: "dQw4w9WgXcQ",
		},
		{
			name:     "bare video ID",
			input:    "dQw4w9WgXcQ",
			expected: "dQw4w9WgXcQ",
		},
		{
			name:     "watch URL without www",
			input:    "https://youtube.com/watch?v=dQw4w9WgXcQ",
			expected: "dQw4w9WgXcQ",
		},
		{
			name:     "mobile URL",
			input:    "https://m.youtube.com/watch?v=dQw4w9WgXcQ",
			expected: "dQw4w9WgXcQ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			videoID, err := parser.ExtractVideoID(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, videoID)
		})
	}
}

func TestExtractVideoID_InvalidURLs(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "empty string", input: ""},
		{name: "invalid URL", input: "https://example.com"},
		{name: "YouTube homepage", input: "https://www.youtube.com"},
		{name: "channel URL", input: "https://www.youtube.com/channel/UCuAXFkgsw1L7xaCfnd5JJOw"},
		{name: "playlist URL", input: "https://www.youtube.com/playlist?list=PLxxx"},
		{name: "too short ID", input: "abc"},
		{name: "too long ID", input: "dQw4w9WgXcQdQw4w9WgXcQ"},
		{name: "invalid characters", input: "dQw4w9Wg@cQ"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.ExtractVideoID(tt.input)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid", "Error should mention 'invalid'")
		})
	}
}

// T022: Unit test for YouTube URL validation
func TestValidateVideoURL_Valid(t *testing.T) {
	validURLs := []string{
		"https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		"https://youtu.be/dQw4w9WgXcQ",
		"https://www.youtube.com/embed/dQw4w9WgXcQ",
		"dQw4w9WgXcQ",
	}

	for _, url := range validURLs {
		t.Run(url, func(t *testing.T) {
			err := parser.ValidateVideoURL(url)
			assert.NoError(t, err)
		})
	}
}

func TestValidateVideoURL_Invalid(t *testing.T) {
	invalidURLs := []string{
		"",
		"https://example.com",
		"not-a-url",
		"https://www.youtube.com/channel/UCxxx",
		"https://www.youtube.com/playlist?list=PLxxx",
	}

	for _, url := range invalidURLs {
		t.Run(url, func(t *testing.T) {
			err := parser.ValidateVideoURL(url)
			assert.Error(t, err)
		})
	}
}

func TestIsValidVideoID(t *testing.T) {
	tests := []struct {
		name     string
		videoID  string
		expected bool
	}{
		{name: "valid 11 char ID", videoID: "dQw4w9WgXcQ", expected: true},
		{name: "valid with underscore", videoID: "abc_1234567", expected: true},
		{name: "valid with hyphen", videoID: "abc-1234567", expected: true},
		{name: "too short", videoID: "abc123", expected: false},
		{name: "too long", videoID: "dQw4w9WgXcQdQw4w9WgXcQ", expected: false},
		{name: "invalid chars", videoID: "dQw4w9Wg@cQ", expected: false},
		{name: "empty", videoID: "", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.IsValidVideoID(tt.videoID)
			assert.Equal(t, tt.expected, result)
		})
	}
}
