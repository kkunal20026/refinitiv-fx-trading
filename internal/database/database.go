package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	"github.com/refinitiv/fx-trading/internal/config"
	"go.uber.org/zap"
)

// New creates a new database connection
func New(cfg config.DatabaseConfig, logger *zap.Logger) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open(cfg.Driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Test connection
	if err := db.PingContext(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	logger.Info("database connected successfully")
	return db, nil
}

// InitializeTables creates required tables
func InitializeTables(db *sql.DB, logger *zap.Logger) error {
	schema := `
	-- Rates table
	CREATE TABLE IF NOT EXISTS rates (
		id UUID PRIMARY KEY,
		from_currency VARCHAR(3) NOT NULL,
		to_currency VARCHAR(3) NOT NULL,
		bid DECIMAL(18, 8) NOT NULL,
		ask DECIMAL(18, 8) NOT NULL,
		mid DECIMAL(18, 8) NOT NULL,
		timestamp TIMESTAMP NOT NULL,
		source VARCHAR(50) NOT NULL,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		UNIQUE(from_currency, to_currency, timestamp)
	);

	CREATE INDEX IF NOT EXISTS idx_rates_pair ON rates(from_currency, to_currency);
	CREATE INDEX IF NOT EXISTS idx_rates_timestamp ON rates(timestamp DESC);

	-- Deals table
	CREATE TABLE IF NOT EXISTS deals (
		id UUID PRIMARY KEY,
		client_id VARCHAR(100) NOT NULL,
		trade_id VARCHAR(100) NOT NULL UNIQUE,
		from_currency VARCHAR(3) NOT NULL,
		to_currency VARCHAR(3) NOT NULL,
		amount DECIMAL(18, 2) NOT NULL,
		rate DECIMAL(18, 8) NOT NULL,
		status VARCHAR(20) NOT NULL,
		direction VARCHAR(10) NOT NULL,
		value_date TIMESTAMP NOT NULL,
		settlement_date TIMESTAMP,
		reference VARCHAR(255),
		notes TEXT,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_deals_client_id ON deals(client_id);
	CREATE INDEX IF NOT EXISTS idx_deals_trade_id ON deals(trade_id);
	CREATE INDEX IF NOT EXISTS idx_deals_status ON deals(status);
	CREATE INDEX IF NOT EXISTS idx_deals_created_at ON deals(created_at DESC);
	`

	if _, err := db.Exec(schema); err != nil {
		// Check if it's a postgres error about existing objects
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "42P07" {
			// Table already exists, ignore
			logger.Info("tables already exist")
			return nil
		}
		return fmt.Errorf("failed to initialize tables: %w", err)
	}

	logger.Info("database tables initialized successfully")
	return nil
}

// Close closes the database connection
func Close(db *sql.DB, logger *zap.Logger) error {
	if err := db.Close(); err != nil {
		logger.Error("failed to close database", zap.Error(err))
		return err
	}
	logger.Info("database connection closed")
	return nil
}
