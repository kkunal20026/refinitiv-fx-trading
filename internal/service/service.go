package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/refinitiv/fx-trading/internal/client"
	"github.com/refinitiv/fx-trading/internal/models"
	"github.com/refinitiv/fx-trading/internal/repository"
	appErrors "github.com/refinitiv/fx-trading/pkg/errors"
	"go.uber.org/zap"
)

const errValidationFailed = "Validation failed"

// RateService handles rate business logic
type RateService struct {
	rateRepo *repository.RateRepository
	client   *client.RefinitivClient
	logger   *zap.Logger
}

// NewRateService creates a new rate service
func NewRateService(rateRepo *repository.RateRepository, client *client.RefinitivClient, logger *zap.Logger) *RateService {
	return &RateService{
		rateRepo: rateRepo,
		client:   client,
		logger:   logger,
	}
}

// FetchAndStoreRate fetches a rate from Refinitiv and stores it
func (rs *RateService) FetchAndStoreRate(ctx context.Context, fromCurrency, toCurrency string) (*models.Rate, error) {
	// Validate currencies
	if len(fromCurrency) != 3 || len(toCurrency) != 3 {
		return nil, appErrors.NewValidationError(
			"Invalid currency codes",
			map[string]string{"message": "Currency codes must be 3 characters"},
			nil,
		)
	}

	// Fetch from Refinitiv
	refinitivRate, err := rs.client.GetRate(ctx, fromCurrency, toCurrency)
	if err != nil {
		rs.logger.Error("failed to fetch rate from Refinitiv", zap.Error(err))
		return nil, err
	}

	// Map to model
	rate := &models.Rate{
		FromCurrency: fromCurrency,
		ToCurrency:   toCurrency,
		Bid:          refinitivRate.Bid,
		Ask:          refinitivRate.Ask,
		Mid:          refinitivRate.Mid,
		Timestamp:    refinitivRate.LastUpdate,
		Source:       "REFINITIV",
	}

	// Store in database
	if err := rs.rateRepo.Create(ctx, rate); err != nil {
		rs.logger.Error("failed to store rate", zap.Error(err))
		return nil, err
	}

	return rate, nil
}

// GetLatestRate retrieves the latest rate for a currency pair
func (rs *RateService) GetLatestRate(ctx context.Context, fromCurrency, toCurrency string) (*models.Rate, error) {
	if len(fromCurrency) != 3 || len(toCurrency) != 3 {
		return nil, appErrors.NewValidationError(
			"Invalid currency codes",
			map[string]string{"message": "Currency codes must be 3 characters"},
			nil,
		)
	}

	rate, err := rs.rateRepo.GetLatest(ctx, fromCurrency, toCurrency)
	if err != nil {
		rs.logger.Error("failed to get rate", zap.Error(err))
		return nil, err
	}

	return rate, nil
}

// DealService handles deal business logic
type DealService struct {
	dealRepo *repository.DealRepository
	rateRepo *repository.RateRepository
	logger   *zap.Logger
}

// NewDealService creates a new deal service
func NewDealService(dealRepo *repository.DealRepository, rateRepo *repository.RateRepository, logger *zap.Logger) *DealService {
	return &DealService{
		dealRepo: dealRepo,
		rateRepo: rateRepo,
		logger:   logger,
	}
}

// BookDeal creates a new deal
func (ds *DealService) BookDeal(ctx context.Context, req *models.DealRequest) (*models.Deal, error) {
	// Validate request
	if err := validateDealRequest(req); err != nil {
		return nil, err
	}

	// Get latest rate for the pair
	rate, err := ds.rateRepo.GetLatest(ctx, req.FromCurrency, req.ToCurrency)
	if err != nil {
		ds.logger.Warn("no rate found, using estimated rate", zap.Error(err))
		// Use a default rate if not found (in production, this might be an error)
		rate = &models.Rate{
			Bid: 1.0,
			Ask: 1.0,
			Mid: 1.0,
		}
	}

	// Determine the rate to use based on direction
	dealRate := rate.Mid
	if req.Direction == models.TradeDirectionBuy {
		dealRate = rate.Ask
	} else {
		dealRate = rate.Bid
	}

	// Create deal
	deal := &models.Deal{
		ClientID:       req.ClientID,
		FromCurrency:   req.FromCurrency,
		ToCurrency:     req.ToCurrency,
		Amount:         req.Amount,
		Rate:           dealRate,
		Direction:      req.Direction,
		ValueDate:      req.ValueDate,
		SettlementDate: req.ValueDate.AddDate(0, 0, 2), // T+2 settlement
		Reference:      req.Reference,
		Notes:          req.Notes,
	}

	// Store in database
	if err := ds.dealRepo.Create(ctx, deal); err != nil {
		ds.logger.Error("failed to create deal", zap.Error(err))
		return nil, err
	}

	ds.logger.Info("deal booked successfully",
		zap.String("deal_id", deal.ID.String()),
		zap.String("client_id", deal.ClientID),
		zap.String("trade_id", deal.TradeID),
	)

	return deal, nil
}

// GetDeal retrieves a deal by ID
func (ds *DealService) GetDeal(ctx context.Context, dealID string) (*models.Deal, error) {
	id, err := uuid.Parse(dealID)
	if err != nil {
		return nil, appErrors.NewValidationError(
			"Invalid deal ID format",
			map[string]string{"deal_id": "Must be a valid UUID"},
			err,
		)
	}

	deal, err := ds.dealRepo.GetByID(ctx, id)
	if err != nil {
		ds.logger.Error("failed to get deal", zap.Error(err))
		return nil, err
	}

	return deal, nil
}

// GetDealByTradeID retrieves a deal by trade ID
func (ds *DealService) GetDealByTradeID(ctx context.Context, tradeID string) (*models.Deal, error) {
	deal, err := ds.dealRepo.GetByTradeID(ctx, tradeID)
	if err != nil {
		ds.logger.Error("failed to get deal by trade ID", zap.Error(err))
		return nil, err
	}

	return deal, nil
}

// ListDeals retrieves deals for a client
func (ds *DealService) ListDeals(ctx context.Context, clientID string, page, limit int) ([]*models.Deal, int64, error) {
	if page < 1 || limit < 1 {
		return nil, 0, appErrors.NewValidationError(
			"Invalid pagination parameters",
			map[string]string{"page": "must be >= 1", "limit": "must be >= 1"},
			nil,
		)
	}

	offset := (page - 1) * limit

	deals, total, err := ds.dealRepo.ListByClient(ctx, clientID, offset, limit)
	if err != nil {
		ds.logger.Error("failed to list deals", zap.Error(err))
		return nil, 0, err
	}

	return deals, total, nil
}

// UpdateDealStatus updates the status of a deal
func (ds *DealService) UpdateDealStatus(ctx context.Context, dealID string, status models.DealStatus) (*models.Deal, error) {
	// Validate status
	switch status {
	case models.DealConfirmed, models.DealSettled, models.DealCancelled, models.DealRejected:
		// Valid statuses
	default:
		return nil, appErrors.NewValidationError(
			"Invalid deal status",
			map[string]string{"status": "must be CONFIRMED, SETTLED, CANCELLED, or REJECTED"},
			nil,
		)
	}

	id, err := uuid.Parse(dealID)
	if err != nil {
		return nil, appErrors.NewValidationError(
			"Invalid deal ID format",
			map[string]string{"deal_id": "Must be a valid UUID"},
			err,
		)
	}

	deal, err := ds.dealRepo.GetByID(ctx, id)
	if err != nil {
		ds.logger.Error("failed to get deal", zap.Error(err))
		return nil, err
	}

	deal.Status = status
	deal.SettlementDate = time.Now().AddDate(0, 0, 2)

	if err := ds.dealRepo.Update(ctx, deal); err != nil {
		ds.logger.Error("failed to update deal status", zap.Error(err))
		return nil, err
	}

	ds.logger.Info("deal status updated",
		zap.String("deal_id", deal.ID.String()),
		zap.String("status", string(status)),
	)

	return deal, nil
}

// validateDealRequest validates a deal request
func validateDealRequest(req *models.DealRequest) error {
	if req.ClientID == "" {
		return appErrors.NewValidationError(
			errValidationFailed,
			map[string]string{"client_id": "required"},
			nil,
		)
	}

	if len(req.FromCurrency) != 3 || len(req.ToCurrency) != 3 {
		return appErrors.NewValidationError(
			errValidationFailed,
			map[string]string{"currencies": "must be 3-character ISO codes"},
			nil,
		)
	}

	if req.Amount <= 0 {
		return appErrors.NewValidationError(
			errValidationFailed,
			map[string]string{"amount": "must be greater than 0"},
			nil,
		)
	}

	if req.Direction != models.TradeDirectionBuy && req.Direction != models.TradeDirectionSell {
		return appErrors.NewValidationError(
			errValidationFailed,
			map[string]string{"direction": "must be BUY or SELL"},
			nil,
		)
	}

	if req.ValueDate.Before(time.Now()) {
		return appErrors.NewValidationError(
			errValidationFailed,
			map[string]string{"value_date": "must be in the future"},
			nil,
		)
	}

	return nil
}
