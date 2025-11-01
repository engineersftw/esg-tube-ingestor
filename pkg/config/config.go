package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	YouTube   YouTubeConfig   `mapstructure:"youtube"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Logging   LoggingConfig   `mapstructure:"logging"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	Retry     RetryConfig     `mapstructure:"retry"`
}

// YouTubeConfig holds YouTube API configuration
type YouTubeConfig struct {
	APIKey string `mapstructure:"api_key"`
}

// DatabaseConfig holds PostgreSQL database configuration
type DatabaseConfig struct {
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	User           string `mapstructure:"user"`
	Password       string `mapstructure:"password"`
	Database       string `mapstructure:"database"`
	SSLMode        string `mapstructure:"sslmode"`
	MaxConnections int    `mapstructure:"max_connections"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level"`  // debug, info, warn, error
	Format string `mapstructure:"format"` // text or json
	File   string `mapstructure:"file"`   // empty = stdout
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int `mapstructure:"requests_per_minute"`
}

// RetryConfig holds retry logic configuration
type RetryConfig struct {
	MaxAttempts     int    `mapstructure:"max_attempts"`
	InitialBackoff  string `mapstructure:"initial_backoff"`  // duration string (e.g., "1s")
	MaxBackoff      string `mapstructure:"max_backoff"`      // duration string (e.g., "60s")
}

// Load loads configuration from file and environment variables
// Environment variables take precedence over config file values
func Load(configPath string) (*Config, error) {
	// Try to load .env file from current directory or parent directories
	loadEnvFile()

	v := viper.New()

	// Set default values
	setDefaults(v)

	// Load config file if provided
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Enable environment variable overrides
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Bind specific environment variables
	v.BindEnv("youtube.api_key", "YOUTUBE_API_KEY")
	v.BindEnv("database.host", "POSTGRES_HOST")
	v.BindEnv("database.port", "POSTGRES_PORT")
	v.BindEnv("database.user", "POSTGRES_USER")
	v.BindEnv("database.password", "POSTGRES_PASSWORD")
	v.BindEnv("database.database", "POSTGRES_DATABASE")
	v.BindEnv("database.sslmode", "POSTGRES_SSLMODE")
	v.BindEnv("database.max_connections", "POSTGRES_MAX_CONNECTIONS")
	v.BindEnv("logging.level", "LOG_LEVEL")
	v.BindEnv("logging.format", "LOG_FORMAT")

	// Unmarshal into Config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate required fields
	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Database defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.database", "esg_tube")
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("database.max_connections", 10)

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "text")
	v.SetDefault("logging.file", "")

	// Rate limiting defaults
	v.SetDefault("rate_limit.requests_per_minute", 100)

	// Retry defaults
	v.SetDefault("retry.max_attempts", 3)
	v.SetDefault("retry.initial_backoff", "1s")
	v.SetDefault("retry.max_backoff", "60s")
}

// validate checks that required configuration values are present
func validate(cfg *Config) error {
	if cfg.YouTube.APIKey == "" {
		return fmt.Errorf("YOUTUBE_API_KEY is required but not set")
	}

	if cfg.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if cfg.Database.User == "" {
		return fmt.Errorf("database user is required")
	}

	if cfg.Database.Database == "" {
		return fmt.Errorf("database name is required")
	}

	return nil
}

// GetDSN returns the PostgreSQL connection string (DSN)
func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Database,
		c.Database.SSLMode,
	)
}

// loadEnvFile attempts to load .env file from current directory or parent directories
func loadEnvFile() {
	// Try to find .env file in current directory or up to 3 parent directories
	currentDir, err := os.Getwd()
	if err != nil {
		return // Silently ignore error, environment variables may still be set
	}

	// Try current directory first
	envPath := filepath.Join(currentDir, ".env")
	if _, err := os.Stat(envPath); err == nil {
		_ = godotenv.Load(envPath)
		return
	}

	// Try parent directories
	for i := 0; i < 3; i++ {
		currentDir = filepath.Dir(currentDir)
		envPath = filepath.Join(currentDir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			_ = godotenv.Load(envPath)
			return
		}
	}

	// If no .env file found, that's okay - environment variables may be set directly
}
