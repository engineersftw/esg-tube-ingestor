package parser_test

import (
	"testing"

	"github.com/engineersftw/youtube-sync/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// T052: Unit test for playlist ID extraction from URLs
func TestExtractPlaylistID_ValidURLs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard playlist URL",
			input:    "https://www.youtube.com/playlist?list=PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf",
			expected: "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf",
		},
		{
			name:     "playlist URL with additional params",
			input:    "https://www.youtube.com/playlist?list=PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf&index=1",
			expected: "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf",
		},
		{
			name:     "watch URL with playlist",
			input:    "https://www.youtube.com/watch?v=dQw4w9WgXcQ&list=PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf",
			expected: "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf",
		},
		{
			name:     "bare playlist ID",
			input:    "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf",
			expected: "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf",
		},
		{
			name:     "playlist URL without www",
			input:    "https://youtube.com/playlist?list=PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf",
			expected: "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf",
		},
		{
			name:     "playlist with different prefix (UU)",
			input:    "UUuAXFkgsw1L7xaCfnd5JJOw",
			expected: "UUuAXFkgsw1L7xaCfnd5JJOw",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			playlistID, err := parser.ExtractPlaylistID(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, playlistID)
		})
	}
}

func TestExtractPlaylistID_InvalidURLs(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "empty string", input: ""},
		{name: "video URL without playlist", input: "https://www.youtube.com/watch?v=dQw4w9WgXcQ"},
		{name: "channel URL", input: "https://www.youtube.com/channel/UCuAXFkgsw1L7xaCfnd5JJOw"},
		{name: "invalid URL", input: "https://example.com"},
		{name: "too short ID", input: "PL123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.ExtractPlaylistID(tt.input)
			assert.Error(t, err)
		})
	}
}

func TestIsValidPlaylistID(t *testing.T) {
	tests := []struct {
		name       string
		playlistID string
		expected   bool
	}{
		{name: "valid PL playlist ID", playlistID: "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf", expected: true},
		{name: "valid UU playlist ID", playlistID: "UUuAXFkgsw1L7xaCfnd5JJOw", expected: true},
		{name: "valid LL playlist ID", playlistID: "LLuAXFkgsw1L7xaCfnd5JJOw", expected: true},
		{name: "valid FL playlist ID", playlistID: "FLuAXFkgsw1L7xaCfnd5JJOw", expected: true},
		{name: "too short", playlistID: "PL12", expected: false},
		{name: "empty", playlistID: "", expected: false},
		{name: "invalid prefix", playlistID: "XXuAXFkgsw1L7xaCfnd5JJOw", expected: false},
		{name: "invalid chars", playlistID: "PL@#$%^&*()", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.IsValidPlaylistID(tt.playlistID)
			assert.Equal(t, tt.expected, result)
		})
	}
}
