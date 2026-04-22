package config

import (
	"time"
)

// DefaultConfig provides default configuration values
var DefaultConfig = Config{
	Server: ServerConfig{
		Port:            8080,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		ShutdownTimeout: 15 * time.Second,
		Environment:     "development",
	},
	Database: DatabaseConfig{
		Driver:          "postgres",
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "postgres",
		DBName:          "refinitiv_db",
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	},
	Logger: LoggerConfig{
		Level:      "info",
		Encoding:   "json",
		OutputPath: "stdout",
	},
	Refinitiv: RefinitivConfig{
		BaseURL:      "https://api.refinitiv.com",
		Timeout:      30 * time.Second,
		MaxRetries:   3,
		RetryBackoff: 1 * time.Second,
	},
	Auth: AuthConfig{
		APIKeyHeader: "X-API-Key",
		TokenExpiry:  24 * time.Hour,
	},
}
