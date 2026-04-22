package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/refinitiv/fx-trading/internal/models"
	appErrors "github.com/refinitiv/fx-trading/pkg/errors"
	"go.uber.org/zap"
)

const (
	errQueryRate    = "failed to query rate"
	errDealNotFound = "deal not found"
	errQueryDeal    = "failed to query deal"
)

// RateRepository handles rate data persistence
type RateRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewRateRepository creates a new rate repository
func NewRateRepository(db *sql.DB, logger *zap.Logger) *RateRepository {
	return &RateRepository{
		db:     db,
		logger: logger,
	}
}

// Create inserts a new rate record
func (r *RateRepository) Create(ctx context.Context, rate *models.Rate) error {
	query := `
		INSERT INTO rates (id, from_currency, to_currency, bid, ask, mid, timestamp, source, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	rate.ID = uuid.New()
	rate.CreatedAt = time.Now()
	rate.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		rate.ID, rate.FromCurrency, rate.ToCurrency, rate.Bid, rate.Ask, rate.Mid,
		rate.Timestamp, rate.Source, rate.CreatedAt, rate.UpdatedAt,
	)
	if err != nil {
		r.logger.Error("failed to insert rate", zap.Error(err))
		return appErrors.NewInternalError("failed to insert rate", err)
	}

	return nil
}

// GetLatest retrieves the latest rate for a currency pair
func (r *RateRepository) GetLatest(ctx context.Context, fromCurrency, toCurrency string) (*models.Rate, error) {
	query := `
		SELECT id, from_currency, to_currency, bid, ask, mid, timestamp, source, created_at, updated_at
		FROM rates
		WHERE from_currency = $1 AND to_currency = $2
		ORDER BY timestamp DESC
		LIMIT 1
	`

	rate := &models.Rate{}
	err := r.db.QueryRowContext(ctx, query, fromCurrency, toCurrency).Scan(
		&rate.ID, &rate.FromCurrency, &rate.ToCurrency, &rate.Bid, &rate.Ask, &rate.Mid,
		&rate.Timestamp, &rate.Source, &rate.CreatedAt, &rate.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, appErrors.NewNotFoundError(
			fmt.Sprintf("no rate found for %s/%s", fromCurrency, toCurrency),
			err,
		)
	}
	if err != nil {
		r.logger.Error(errQueryRate, zap.Error(err))
		return nil, appErrors.NewInternalError(errQueryRate, err)
	}

	return rate, nil
}

// GetByID retrieves a rate by ID
func (r *RateRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Rate, error) {
	query := `
		SELECT id, from_currency, to_currency, bid, ask, mid, timestamp, source, created_at, updated_at
		FROM rates
		WHERE id = $1
	`

	rate := &models.Rate{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rate.ID, &rate.FromCurrency, &rate.ToCurrency, &rate.Bid, &rate.Ask, &rate.Mid,
		&rate.Timestamp, &rate.Source, &rate.CreatedAt, &rate.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, appErrors.NewNotFoundError("rate not found", err)
	}
	if err != nil {
		r.logger.Error(errQueryRate, zap.Error(err))
		return nil, appErrors.NewInternalError(errQueryRate, err)
	}

	return rate, nil
}

// DealRepository handles deal data persistence
type DealRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewDealRepository creates a new deal repository
func NewDealRepository(db *sql.DB, logger *zap.Logger) *DealRepository {
	return &DealRepository{
		db:     db,
		logger: logger,
	}
}

// Create inserts a new deal record
func (d *DealRepository) Create(ctx context.Context, deal *models.Deal) error {
	query := `
		INSERT INTO deals (id, client_id, trade_id, from_currency, to_currency, amount, rate, status, 
			direction, value_date, settlement_date, reference, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	deal.ID = uuid.New()
	deal.TradeID = fmt.Sprintf("TRD-%d-%s", time.Now().Unix(), deal.ID.String()[:8])
	deal.Status = models.DealPending
	deal.CreatedAt = time.Now()
	deal.UpdatedAt = time.Now()

	_, err := d.db.ExecContext(ctx, query,
		deal.ID, deal.ClientID, deal.TradeID, deal.FromCurrency, deal.ToCurrency, deal.Amount, deal.Rate,
		deal.Status, deal.Direction, deal.ValueDate, deal.SettlementDate, deal.Reference, deal.Notes,
		deal.CreatedAt, deal.UpdatedAt,
	)
	if err != nil {
		d.logger.Error("failed to insert deal", zap.Error(err))
		return appErrors.NewInternalError("failed to insert deal", err)
	}

	return nil
}

// GetByID retrieves a deal by ID
func (d *DealRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Deal, error) {
	query := `
		SELECT id, client_id, trade_id, from_currency, to_currency, amount, rate, status,
			direction, value_date, settlement_date, reference, notes, created_at, updated_at
		FROM deals
		WHERE id = $1
	`

	deal := &models.Deal{}
	err := d.db.QueryRowContext(ctx, query, id).Scan(
		&deal.ID, &deal.ClientID, &deal.TradeID, &deal.FromCurrency, &deal.ToCurrency, &deal.Amount, &deal.Rate,
		&deal.Status, &deal.Direction, &deal.ValueDate, &deal.SettlementDate, &deal.Reference, &deal.Notes,
		&deal.CreatedAt, &deal.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, appErrors.NewNotFoundError(errDealNotFound, err)
	}
	if err != nil {
		d.logger.Error(errQueryDeal, zap.Error(err))
		return nil, appErrors.NewInternalError(errQueryDeal, err)
	}

	return deal, nil
}

// GetByTradeID retrieves a deal by trade ID
func (d *DealRepository) GetByTradeID(ctx context.Context, tradeID string) (*models.Deal, error) {
	query := `
		SELECT id, client_id, trade_id, from_currency, to_currency, amount, rate, status,
			direction, value_date, settlement_date, reference, notes, created_at, updated_at
		FROM deals
		WHERE trade_id = $1
	`

	deal := &models.Deal{}
	err := d.db.QueryRowContext(ctx, query, tradeID).Scan(
		&deal.ID, &deal.ClientID, &deal.TradeID, &deal.FromCurrency, &deal.ToCurrency, &deal.Amount, &deal.Rate,
		&deal.Status, &deal.Direction, &deal.ValueDate, &deal.SettlementDate, &deal.Reference, &deal.Notes,
		&deal.CreatedAt, &deal.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, appErrors.NewNotFoundError(errDealNotFound, err)
	}
	if err != nil {
		d.logger.Error(errQueryDeal, zap.Error(err))
		return nil, appErrors.NewInternalError(errQueryDeal, err)
	}

	return deal, nil
}

// ListByClient retrieves all deals for a client with pagination
func (d *DealRepository) ListByClient(ctx context.Context, clientID string, offset, limit int) ([]*models.Deal, int64, error) {
	countQuery := `SELECT COUNT(*) FROM deals WHERE client_id = $1`
	var total int64
	err := d.db.QueryRowContext(ctx, countQuery, clientID).Scan(&total)
	if err != nil {
		d.logger.Error("failed to count deals", zap.Error(err))
		return nil, 0, appErrors.NewInternalError("failed to count deals", err)
	}

	query := `
		SELECT id, client_id, trade_id, from_currency, to_currency, amount, rate, status,
			direction, value_date, settlement_date, reference, notes, created_at, updated_at
		FROM deals
		WHERE client_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := d.db.QueryContext(ctx, query, clientID, limit, offset)
	if err != nil {
		d.logger.Error("failed to query deals", zap.Error(err))
		return nil, 0, appErrors.NewInternalError("failed to query deals", err)
	}
	defer rows.Close()

	var deals []*models.Deal
	for rows.Next() {
		deal := &models.Deal{}
		err := rows.Scan(
			&deal.ID, &deal.ClientID, &deal.TradeID, &deal.FromCurrency, &deal.ToCurrency, &deal.Amount, &deal.Rate,
			&deal.Status, &deal.Direction, &deal.ValueDate, &deal.SettlementDate, &deal.Reference, &deal.Notes,
			&deal.CreatedAt, &deal.UpdatedAt,
		)
		if err != nil {
			d.logger.Error("failed to scan deal", zap.Error(err))
			return nil, 0, appErrors.NewInternalError("failed to scan deal", err)
		}
		deals = append(deals, deal)
	}

	if err = rows.Err(); err != nil {
		d.logger.Error("row iteration error", zap.Error(err))
		return nil, 0, appErrors.NewInternalError("failed to iterate deals", err)
	}

	return deals, total, nil
}

// Update updates a deal status
func (d *DealRepository) Update(ctx context.Context, deal *models.Deal) error {
	query := `
		UPDATE deals
		SET status = $2, settlement_date = $3, notes = $4, updated_at = $5
		WHERE id = $1
	`

	deal.UpdatedAt = time.Now()
	result, err := d.db.ExecContext(ctx, query, deal.ID, deal.Status, deal.SettlementDate, deal.Notes, deal.UpdatedAt)
	if err != nil {
		d.logger.Error("failed to update deal", zap.Error(err))
		return appErrors.NewInternalError("failed to update deal", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return appErrors.NewInternalError("failed to get rows affected", err)
	}

	if rowsAffected == 0 {
		return appErrors.NewNotFoundError(errDealNotFound, nil)
	}

	return nil
}
