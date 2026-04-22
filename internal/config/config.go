package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config represents application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Logger   LoggerConfig
	Refinitiv RefinitivConfig
	Auth     AuthConfig
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	Environment     string        `mapstructure:"environment"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Driver          string        `mapstructure:"driver"`
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// LoggerConfig represents logger configuration
type LoggerConfig struct {
	Level      string `mapstructure:"level"`
	Encoding   string `mapstructure:"encoding"`
	OutputPath string `mapstructure:"output_path"`
}

// RefinitivConfig represents Refinitiv API configuration
type RefinitivConfig struct {
	BaseURL         string        `mapstructure:"base_url"`
	Username        string        `mapstructure:"username"`
	Password        string        `mapstructure:"password"`
	Timeout         time.Duration `mapstructure:"timeout"`
	MaxRetries      int           `mapstructure:"max_retries"`
	RetryBackoff    time.Duration `mapstructure:"retry_backoff"`
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	APIKeyHeader string `mapstructure:"api_key_header"`
	JWTSecret    string `mapstructure:"jwt_secret"`
	TokenExpiry  time.Duration `mapstructure:"token_expiry"`
}

// Load loads configuration from files and environment
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Set config file
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("../../") // support running from cmd/server/
	}
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Read environment variables
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", 30*time.Second)
	v.SetDefault("server.write_timeout", 30*time.Second)
	v.SetDefault("server.shutdown_timeout", 15*time.Second)
	v.SetDefault("server.environment", "development")

	v.SetDefault("database.driver", "postgres")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 5)
	v.SetDefault("database.conn_max_lifetime", 5*time.Minute)
	v.SetDefault("database.sslmode", "disable")

	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.encoding", "json")

	v.SetDefault("refinitiv.timeout", 30*time.Second)
	v.SetDefault("refinitiv.max_retries", 3)
	v.SetDefault("refinitiv.retry_backoff", 1*time.Second)

	v.SetDefault("auth.api_key_header", "X-API-Key")
	v.SetDefault("auth.token_expiry", 24*time.Hour)
}

func validate(cfg *Config) error {
	if cfg.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if cfg.Refinitiv.BaseURL == "" {
		return fmt.Errorf("refinitiv base URL is required")
	}

	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", cfg.Server.Port)
	}

	return nil
}
