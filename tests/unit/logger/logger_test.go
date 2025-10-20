package logger_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/engineersftw/youtube-sync/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogger_TextFormat(t *testing.T) {
	var buf bytes.Buffer
	log, err := logger.New("info", "text", &buf)
	require.NoError(t, err)
	require.NotNil(t, log)

	log.Info("test message")
	output := buf.String()
	assert.Contains(t, output, "test message")
	assert.Contains(t, output, "INFO")
}

func TestNewLogger_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	log, err := logger.New("info", "json", &buf)
	require.NoError(t, err)
	require.NotNil(t, log)

	log.Info("test message")
	output := buf.String()

	// Verify it's valid JSON
	var jsonOutput map[string]interface{}
	err = json.Unmarshal([]byte(output), &jsonOutput)
	require.NoError(t, err, "Output should be valid JSON")

	assert.Equal(t, "INFO", jsonOutput["level"]) // zap uses uppercase levels
	assert.Equal(t, "test message", jsonOutput["msg"])
}

func TestLogger_WithCorrelationID(t *testing.T) {
	var buf bytes.Buffer
	log, err := logger.New("info", "json", &buf)
	require.NoError(t, err)

	correlationID := "test-correlation-123"
	logWithCorrelation := log.WithCorrelationID(correlationID)
	logWithCorrelation.Info("test with correlation")

	output := buf.String()
	var jsonOutput map[string]interface{}
	err = json.Unmarshal([]byte(output), &jsonOutput)
	require.NoError(t, err)

	assert.Equal(t, correlationID, jsonOutput["correlation_id"])
}

func TestLogger_WithFields(t *testing.T) {
	var buf bytes.Buffer
	log, err := logger.New("info", "json", &buf)
	require.NoError(t, err)

	logWithFields := log.WithFields(map[string]interface{}{
		"video_id": "abc123",
		"channel":  "test_channel",
		"count":    42,
	})
	logWithFields.Info("test with fields")

	output := buf.String()
	var jsonOutput map[string]interface{}
	err = json.Unmarshal([]byte(output), &jsonOutput)
	require.NoError(t, err)

	assert.Equal(t, "abc123", jsonOutput["video_id"])
	assert.Equal(t, "test_channel", jsonOutput["channel"])
	assert.Equal(t, float64(42), jsonOutput["count"]) // JSON numbers are float64
}

func TestLogger_LogLevels(t *testing.T) {
	tests := []struct {
		name      string
		logLevel  string
		logFunc   string
		shouldLog bool
	}{
		{"debug logged at debug level", "debug", "Debug", true},
		{"info logged at debug level", "debug", "Info", true},
		{"debug not logged at info level", "info", "Debug", false},
		{"info logged at info level", "info", "Info", true},
		{"warn logged at info level", "info", "Warn", true},
		{"error logged at info level", "info", "Error", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			log, err := logger.New(tt.logLevel, "text", &buf)
			require.NoError(t, err)

			// Call the appropriate log function
			switch tt.logFunc {
			case "Debug":
				log.Debug("test message")
			case "Info":
				log.Info("test message")
			case "Warn":
				log.Warn("test message")
			case "Error":
				log.Error("test message")
			}

			output := buf.String()
			if tt.shouldLog {
				assert.Contains(t, output, "test message")
			} else {
				assert.Empty(t, output, "Message should not be logged at this level")
			}
		})
	}
}

func TestLogger_InvalidLogLevel(t *testing.T) {
	var buf bytes.Buffer
	_, err := logger.New("invalid_level", "text", &buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid log level")
}

func TestLogger_InvalidFormat(t *testing.T) {
	var buf bytes.Buffer
	_, err := logger.New("info", "invalid_format", &buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid log format")
}

func TestLogger_Sync(t *testing.T) {
	var buf bytes.Buffer
	log, err := logger.New("info", "text", &buf)
	require.NoError(t, err)

	log.Info("test message")
	err = log.Sync()
	assert.NoError(t, err, "Sync should not error")
}
