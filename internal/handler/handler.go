package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/refinitiv/fx-trading/internal/models"
	"github.com/refinitiv/fx-trading/internal/service"
	appErrors "github.com/refinitiv/fx-trading/pkg/errors"
	"go.uber.org/zap"
)

const errMissingParams = "Missing query parameters"

// Handler represents HTTP request handlers
type Handler struct {
	rateService *service.RateService
	dealService *service.DealService
	logger      *zap.Logger
}

// New creates a new handler instance
func New(rateService *service.RateService, dealService *service.DealService, logger *zap.Logger) *Handler {
	return &Handler{
		rateService: rateService,
		dealService: dealService,
		logger:      logger,
	}
}

// RateResponse represents API response
type Response struct {
	Success bool         `json:"success"`
	Data    interface{}  `json:"data,omitempty"`
	Error   *ErrorDetail `json:"error,omitempty"`
}

// ErrorDetail represents error details
type ErrorDetail struct {
	Type       string      `json:"type"`
	Message    string      `json:"message"`
	StatusCode int         `json:"status_code"`
	Details    interface{} `json:"details,omitempty"`
}

// GetRate retrieves current FX rate
func (h *Handler) GetRate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	fromCurrency := r.URL.Query().Get("from")
	toCurrency := r.URL.Query().Get("to")

	if fromCurrency == "" || toCurrency == "" {
		h.writeError(w, appErrors.NewValidationError(
			errMissingParams,
			map[string]string{"from": "required", "to": "required"},
			nil,
		))
		return
	}

	rate, err := h.rateService.FetchAndStoreRate(ctx, fromCurrency, toCurrency)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response := models.RateResponse{
		ID:           rate.ID,
		FromCurrency: rate.FromCurrency,
		ToCurrency:   rate.ToCurrency,
		Bid:          rate.Bid,
		Ask:          rate.Ask,
		Mid:          rate.Mid,
		Timestamp:    rate.Timestamp,
		Source:       rate.Source,
	}

	h.writeJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    response,
	})
}

// GetRateByID retrieves a rate by ID
func (h *Handler) GetRateByID(w http.ResponseWriter, r *http.Request) {
	rateID := chi.URLParam(r, "id")

	if rateID == "" {
		h.writeError(w, appErrors.NewValidationError(
			errMissingParams,
			map[string]string{"id": "required"},
			nil,
		))
		return
	}

	// In production, you'd parse the UUID
	h.logger.Info("fetching rate", zap.String("rate_id", rateID))

	h.writeJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    nil,
	})
}

// BookDeal creates a new FX deal
func (h *Handler) BookDeal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req models.DealRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, appErrors.NewValidationError(
			"Invalid request body",
			map[string]string{"body": "must be valid JSON"},
			err,
		))
		return
	}

	deal, err := h.dealService.BookDeal(ctx, &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response := models.DealResponse{
		ID:           deal.ID,
		ClientID:     deal.ClientID,
		TradeID:      deal.TradeID,
		FromCurrency: deal.FromCurrency,
		ToCurrency:   deal.ToCurrency,
		Amount:       deal.Amount,
		Rate:         deal.Rate,
		Status:       deal.Status,
		Direction:    deal.Direction,
		ValueDate:    deal.ValueDate,
		Reference:    deal.Reference,
	}

	h.logger.Info("deal booked", zap.String("deal_id", deal.ID.String()))
	h.writeJSON(w, http.StatusCreated, Response{
		Success: true,
		Data:    response,
	})
}

// GetDeal retrieves a deal by ID
func (h *Handler) GetDeal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dealID := chi.URLParam(r, "id")

	if dealID == "" {
		h.writeError(w, appErrors.NewValidationError(
			errMissingParams,
			map[string]string{"id": "required"},
			nil,
		))
		return
	}

	deal, err := h.dealService.GetDeal(ctx, dealID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response := models.DealResponse{
		ID:             deal.ID,
		ClientID:       deal.ClientID,
		TradeID:        deal.TradeID,
		FromCurrency:   deal.FromCurrency,
		ToCurrency:     deal.ToCurrency,
		Amount:         deal.Amount,
		Rate:           deal.Rate,
		Status:         deal.Status,
		Direction:      deal.Direction,
		ValueDate:      deal.ValueDate,
		SettlementDate: deal.SettlementDate,
		Reference:      deal.Reference,
	}

	h.writeJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    response,
	})
}

// GetDealByTradeID retrieves a deal by trade ID
func (h *Handler) GetDealByTradeID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tradeID := chi.URLParam(r, "trade_id")

	if tradeID == "" {
		h.writeError(w, appErrors.NewValidationError(
			errMissingParams,
			map[string]string{"trade_id": "required"},
			nil,
		))
		return
	}

	deal, err := h.dealService.GetDealByTradeID(ctx, tradeID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response := models.DealResponse{
		ID:             deal.ID,
		ClientID:       deal.ClientID,
		TradeID:        deal.TradeID,
		FromCurrency:   deal.FromCurrency,
		ToCurrency:     deal.ToCurrency,
		Amount:         deal.Amount,
		Rate:           deal.Rate,
		Status:         deal.Status,
		Direction:      deal.Direction,
		ValueDate:      deal.ValueDate,
		SettlementDate: deal.SettlementDate,
		Reference:      deal.Reference,
	}

	h.writeJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    response,
	})
}

// ListDeals retrieves deals for a client
func (h *Handler) ListDeals(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	clientID := r.URL.Query().Get("client_id")

	if clientID == "" {
		h.writeError(w, appErrors.NewValidationError(
			errMissingParams,
			map[string]string{"client_id": "required"},
			nil,
		))
		return
	}

	page := 1
	limit := 20

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	deals, total, err := h.dealService.ListDeals(ctx, clientID, page, limit)
	if err != nil {
		h.handleError(w, err)
		return
	}

	var responses []models.DealResponse
	for _, deal := range deals {
		responses = append(responses, models.DealResponse{
			ID:             deal.ID,
			ClientID:       deal.ClientID,
			TradeID:        deal.TradeID,
			FromCurrency:   deal.FromCurrency,
			ToCurrency:     deal.ToCurrency,
			Amount:         deal.Amount,
			Rate:           deal.Rate,
			Status:         deal.Status,
			Direction:      deal.Direction,
			ValueDate:      deal.ValueDate,
			SettlementDate: deal.SettlementDate,
			Reference:      deal.Reference,
		})
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)

	h.writeJSON(w, http.StatusOK, Response{
		Success: true,
		Data: models.PaginatedResponse{
			Data:       responses,
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// UpdateDealStatus updates deal status
func (h *Handler) UpdateDealStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	dealID := chi.URLParam(r, "id")
	status := r.URL.Query().Get("status")

	if dealID == "" || status == "" {
		h.writeError(w, appErrors.NewValidationError(
			errMissingParams,
			map[string]string{"id": "required", "status": "required"},
			nil,
		))
		return
	}

	deal, err := h.dealService.UpdateDealStatus(ctx, dealID, models.DealStatus(status))
	if err != nil {
		h.handleError(w, err)
		return
	}

	response := models.DealResponse{
		ID:             deal.ID,
		ClientID:       deal.ClientID,
		TradeID:        deal.TradeID,
		FromCurrency:   deal.FromCurrency,
		ToCurrency:     deal.ToCurrency,
		Amount:         deal.Amount,
		Rate:           deal.Rate,
		Status:         deal.Status,
		Direction:      deal.Direction,
		ValueDate:      deal.ValueDate,
		SettlementDate: deal.SettlementDate,
		Reference:      deal.Reference,
	}

	h.writeJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    response,
	})
}

// HealthCheck returns service health status
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	services := map[string]string{
		"database":  "ok",
		"refinitiv": "ok",
	}

	response := models.HealthCheckResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Services:  services,
	}

	h.writeJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    response,
	})
}

// Helper functions

func (h *Handler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	var statusCode int
	var errorDetail *ErrorDetail

	appErr, isAppError := appErrors.IsAppError(err)
	if !isAppError {
		statusCode = http.StatusInternalServerError
		errorDetail = &ErrorDetail{
			Type:       "INTERNAL_ERROR",
			Message:    "An unexpected error occurred",
			StatusCode: statusCode,
		}
	} else {
		statusCode = appErr.StatusCode
		errorDetail = &ErrorDetail{
			Type:       string(appErr.Type),
			Message:    appErr.Message,
			StatusCode: statusCode,
			Details:    appErr.Details,
		}
	}

	h.logger.Error("request error",
		zap.String("error_type", errorDetail.Type),
		zap.String("message", errorDetail.Message),
		zap.Int("status_code", statusCode),
		zap.Error(err),
	)

	h.writeJSON(w, statusCode, Response{
		Success: false,
		Error:   errorDetail,
	})
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	appErr, isAppError := appErrors.IsAppError(err)
	if !isAppError {
		h.writeError(w, appErrors.NewInternalError("An unexpected error occurred", err))
		return
	}

	h.writeError(w, appErr)
}
