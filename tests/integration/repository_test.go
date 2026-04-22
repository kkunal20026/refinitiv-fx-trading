// +build integration

package integration_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/refinitiv/fx-trading/internal/config"
	"github.com/refinitiv/fx-trading/internal/database"
	"github.com/refinitiv/fx-trading/internal/models"
	"github.com/refinitiv/fx-trading/internal/repository"
	"github.com/refinitiv/fx-trading/pkg/logger"
)

func setupTestDB(t *testing.T) *sql.DB {
	cfg := config.DatabaseConfig{
		Driver:          "postgres",
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "postgres",
		DBName:          "refinitiv_test",
		SSLMode:         "disable",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}

	log, _ := logger.New(logger.Config{
		Level:    "debug",
		Encoding: "json",
	})

	db, err := database.New(cfg, log)
	require.NoError(t, err)

	err = database.InitializeTables(db, log)
	require.NoError(t, err)

	return db
}

func TestRateRepository_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := setupTestDB(t)
	defer database.Close(db, nil)

	log, _ := logger.New(logger.Config{Level: "debug"})
	rateRepo := repository.NewRateRepository(db, log)

	rate := &models.Rate{
		FromCurrency: "USD",
		ToCurrency:   "EUR",
		Bid:          1.0850,
		Ask:          1.0855,
		Mid:          1.0852,
		Timestamp:    time.Now(),
		Source:       "REFINITIV",
	}

	err := rateRepo.Create(context.Background(), rate)
	require.NoError(t, err)
	require.NotNil(t, rate.ID)
	require.NotNil(t, rate.CreatedAt)
}

func TestDealRepository_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	db := setupTestDB(t)
	defer database.Close(db, nil)

	log, _ := logger.New(logger.Config{Level: "debug"})
	dealRepo := repository.NewDealRepository(db, log)

	deal := &models.Deal{
		ClientID:      "CLIENT123",
		FromCurrency:  "USD",
		ToCurrency:    "EUR",
		Amount:        100000.00,
		Rate:          1.0852,
		Direction:     models.TradeDirectionBuy,
		ValueDate:     time.Now().Add(24 * time.Hour),
		Reference:     "TEST001",
	}

	err := dealRepo.Create(context.Background(), deal)
	require.NoError(t, err)
	require.NotNil(t, deal.ID)
	require.NotNil(t, deal.TradeID)
}
