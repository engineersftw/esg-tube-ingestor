package config_test

import (
	"os"
	"testing"

	"github.com/engineersftw/youtube-sync/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_FromEnvironment(t *testing.T) {
	// Set environment variables
	os.Setenv("YOUTUBE_API_KEY", "test-api-key")
	os.Setenv("POSTGRES_HOST", "testhost")
	os.Setenv("POSTGRES_PORT", "5433")
	os.Setenv("POSTGRES_USER", "testuser")
	os.Setenv("POSTGRES_PASSWORD", "testpass")
	os.Setenv("POSTGRES_DATABASE", "testdb")
	defer func() {
		os.Unsetenv("YOUTUBE_API_KEY")
		os.Unsetenv("POSTGRES_HOST")
		os.Unsetenv("POSTGRES_PORT")
		os.Unsetenv("POSTGRES_USER")
		os.Unsetenv("POSTGRES_PASSWORD")
		os.Unsetenv("POSTGRES_DATABASE")
	}()

	cfg, err := config.Load("")
	require.NoError(t, err)

	assert.Equal(t, "test-api-key", cfg.YouTube.APIKey)
	assert.Equal(t, "testhost", cfg.Database.Host)
	assert.Equal(t, 5433, cfg.Database.Port)
	assert.Equal(t, "testuser", cfg.Database.User)
	assert.Equal(t, "testpass", cfg.Database.Password)
	assert.Equal(t, "testdb", cfg.Database.Database)
}

func TestLoadConfig_MissingAPIKey(t *testing.T) {
	os.Unsetenv("YOUTUBE_API_KEY")

	_, err := config.Load("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "YOUTUBE_API_KEY")
}

func TestLoadConfig_DefaultValues(t *testing.T) {
	os.Setenv("YOUTUBE_API_KEY", "test-key")
	defer os.Unsetenv("YOUTUBE_API_KEY")

	cfg, err := config.Load("")
	require.NoError(t, err)

	// Check default values
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "text", cfg.Logging.Format)
	assert.Equal(t, 100, cfg.RateLimit.RequestsPerMinute)
}

func TestLoadConfig_GetDSN(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "secret",
			Database: "testdb",
			SSLMode:  "disable",
		},
	}

	dsn := cfg.GetDSN()
	assert.Contains(t, dsn, "host=localhost")
	assert.Contains(t, dsn, "port=5432")
	assert.Contains(t, dsn, "user=postgres")
	assert.Contains(t, dsn, "password=secret")
	assert.Contains(t, dsn, "dbname=testdb")
	assert.Contains(t, dsn, "sslmode=disable")
}

func TestLoadConfig_FromFile(t *testing.T) {
	// Create a temporary config file
	configContent := `
youtube:
  api_key: file-api-key

database:
  host: filehost
  port: 5433
  user: fileuser
  password: filepass
  database: filedb

logging:
  level: debug
  format: json
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	cfg, err := config.Load(tmpFile.Name())
	require.NoError(t, err)

	assert.Equal(t, "file-api-key", cfg.YouTube.APIKey)
	assert.Equal(t, "filehost", cfg.Database.Host)
	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
}

func TestLoadConfig_EnvironmentOverridesFile(t *testing.T) {
	// Create config file
	configContent := `
youtube:
  api_key: file-api-key
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	// Set environment variable (should override file)
	os.Setenv("YOUTUBE_API_KEY", "env-api-key")
	defer os.Unsetenv("YOUTUBE_API_KEY")

	cfg, err := config.Load(tmpFile.Name())
	require.NoError(t, err)

	assert.Equal(t, "env-api-key", cfg.YouTube.APIKey, "Environment variable should override file")
}
