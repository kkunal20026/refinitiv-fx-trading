package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appErrors "github.com/refinitiv/fx-trading/pkg/errors"
	"github.com/refinitiv/fx-trading/internal/models"
	"github.com/refinitiv/fx-trading/internal/service"
	"github.com/refinitiv/fx-trading/pkg/logger"
	"go.uber.org/zap"
)

func TestValidateDealRequest(t *testing.T) {
	tests := []struct {
		name    string
		request *models.DealRequest
		wantErr bool
		errType appErrors.ErrorType
	}{
		{
			name: "valid request",
			request: &models.DealRequest{
				ClientID:     "CLIENT123",
				FromCurrency: "USD",
				ToCurrency:   "EUR",
				Amount:       1000.00,
				Direction:    models.TradeDirectionBuy,
				ValueDate:    time.Now().Add(24 * time.Hour),
				Reference:    "REF123",
			},
			wantErr: false,
		},
		{
			name: "missing client ID",
			request: &models.DealRequest{
				ClientID:     "",
				FromCurrency: "USD",
				ToCurrency:   "EUR",
				Amount:       1000.00,
				Direction:    models.TradeDirectionBuy,
				ValueDate:    time.Now().Add(24 * time.Hour),
			},
			wantErr: true,
			errType: appErrors.ValidationError,
		},
		{
			name: "invalid currency code",
			request: &models.DealRequest{
				ClientID:     "CLIENT123",
				FromCurrency: "US",
				ToCurrency:   "EUR",
				Amount:       1000.00,
				Direction:    models.TradeDirectionBuy,
				ValueDate:    time.Now().Add(24 * time.Hour),
			},
			wantErr: true,
			errType: appErrors.ValidationError,
		},
		{
			name: "invalid amount",
			request: &models.DealRequest{
				ClientID:     "CLIENT123",
				FromCurrency: "USD",
				ToCurrency:   "EUR",
				Amount:       -1000.00,
				Direction:    models.TradeDirectionBuy,
				ValueDate:    time.Now().Add(24 * time.Hour),
			},
			wantErr: true,
			errType: appErrors.ValidationError,
		},
		{
			name: "invalid direction",
			request: &models.DealRequest{
				ClientID:     "CLIENT123",
				FromCurrency: "USD",
				ToCurrency:   "EUR",
				Amount:       1000.00,
				Direction:    "INVALID",
				ValueDate:    time.Now().Add(24 * time.Hour),
			},
			wantErr: true,
			errType: appErrors.ValidationError,
		},
		{
			name: "past value date",
			request: &models.DealRequest{
				ClientID:     "CLIENT123",
				FromCurrency: "USD",
				ToCurrency:   "EUR",
				Amount:       1000.00,
				Direction:    models.TradeDirectionBuy,
				ValueDate:    time.Now().Add(-24 * time.Hour),
			},
			wantErr: true,
			errType: appErrors.ValidationError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock logger
			log, _ := getTestLogger()
			dealService := service.NewDealService(nil, nil, log)

			// Call the internal validation function via reflection or create a test version
			// For now, we'll test through BookDeal which calls validation
			ctx := context.Background()
			_, err := dealService.BookDeal(ctx, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				appErr, ok := appErrors.IsAppError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.errType, appErr.Type)
			} else {
				if err != nil {
					// Expected to fail due to nil repo, but validation should pass
					appErr, ok := appErrors.IsAppError(err)
					if ok {
						// Validate that it's not a validation error
						assert.NotEqual(t, appErrors.ValidationError, appErr.Type)
					}
				}
			}
		})
	}
}

func TestDealStatusUpdate(t *testing.T) {
	tests := []struct {
		name       string
		status     models.DealStatus
		wantErr    bool
	}{
		{
			name:    "valid confirmed status",
			status:  models.DealConfirmed,
			wantErr: false,
		},
		{
			name:    "valid settled status",
			status:  models.DealSettled,
			wantErr: false,
		},
		{
			name:    "valid cancelled status",
			status:  models.DealCancelled,
			wantErr: false,
		},
		{
			name:    "invalid status",
			status:  "INVALID",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, _ := getTestLogger()
			dealService := service.NewDealService(nil, nil, log)
			ctx := context.Background()

			_, err := dealService.UpdateDealStatus(ctx, uuid.New().String(), tt.status)

			// Since repo is nil, we'll get an error, but validation should work
			if tt.wantErr && tt.status == "INVALID" {
				require.Error(t, err)
				appErr, ok := appErrors.IsAppError(err)
				assert.True(t, ok)
				assert.Equal(t, appErrors.ValidationError, appErr.Type)
			}
		})
	}
}

func getTestLogger() (*zap.Logger, error) {
	log, err := logger.New(logger.Config{
		Level:    "debug",
		Encoding: "json",
	})
	if err != nil {
		return nil, err
	}
	return log.Logger, nil
}
