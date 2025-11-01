package parser_test

import (
	"testing"

	"github.com/engineersftw/youtube-sync/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T036: Unit test for channel ID extraction from URLs
func TestExtractChannelID_ValidURLs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "channel URL",
			input:    "https://www.youtube.com/channel/UCuAXFkgsw1L7xaCfnd5JJOw",
			expected: "UCuAXFkgsw1L7xaCfnd5JJOw",
		},
		{
			name:     "custom URL with @",
			input:    "https://www.youtube.com/@RickAstleyYT",
			expected: "@RickAstleyYT",
		},
		{
			name:     "custom URL with /c/",
			input:    "https://www.youtube.com/c/RickAstleyYT",
			expected: "RickAstleyYT",
		},
		{
			name:     "user URL",
			input:    "https://www.youtube.com/user/RickAstleyVEVO",
			expected: "RickAstleyVEVO",
		},
		{
			name:     "bare channel ID",
			input:    "UCuAXFkgsw1L7xaCfnd5JJOw",
			expected: "UCuAXFkgsw1L7xaCfnd5JJOw",
		},
		{
			name:     "channel URL without www",
			input:    "https://youtube.com/channel/UCuAXFkgsw1L7xaCfnd5JJOw",
			expected: "UCuAXFkgsw1L7xaCfnd5JJOw",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			channelID, err := parser.ExtractChannelID(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, channelID)
		})
	}
}

func TestExtractChannelID_InvalidURLs(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "empty string", input: ""},
		{name: "video URL", input: "https://www.youtube.com/watch?v=dQw4w9WgXcQ"},
		{name: "playlist URL", input: "https://www.youtube.com/playlist?list=PLxxx"},
		{name: "invalid URL", input: "https://example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.ExtractChannelID(tt.input)
			assert.Error(t, err)
		})
	}
}

func TestIsValidChannelID(t *testing.T) {
	tests := []struct {
		name      string
		channelID string
		expected  bool
	}{
		{name: "valid UC channel ID", channelID: "UCuAXFkgsw1L7xaCfnd5JJOw", expected: true},
		{name: "valid custom name", channelID: "GoogleDevelopers", expected: true},
		{name: "valid @ handle", channelID: "@RickAstleyYT", expected: true},
		{name: "too short", channelID: "UC", expected: false},
		{name: "empty", channelID: "", expected: false},
		{name: "invalid chars", channelID: "UC@#$%", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.IsValidChannelID(tt.channelID)
			assert.Equal(t, tt.expected, result)
		})
	}
}
