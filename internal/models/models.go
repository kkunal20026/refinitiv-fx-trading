package models

import (
	"time"

	"github.com/google/uuid"
)

// Rate represents an FX rate from Refinitiv
type Rate struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	FromCurrency string   `json:"from_currency" db:"from_currency"`
	ToCurrency   string   `json:"to_currency" db:"to_currency"`
	Bid         float64   `json:"bid" db:"bid"`
	Ask         float64   `json:"ask" db:"ask"`
	Mid         float64   `json:"mid" db:"mid"`
	Timestamp   time.Time `json:"timestamp" db:"timestamp"`
	Source      string    `json:"source" db:"source"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Deal represents a FX deal
type Deal struct {
	ID            uuid.UUID   `json:"id" db:"id"`
	ClientID      string      `json:"client_id" db:"client_id"`
	TradeID       string      `json:"trade_id" db:"trade_id"`
	FromCurrency  string      `json:"from_currency" db:"from_currency"`
	ToCurrency    string      `json:"to_currency" db:"to_currency"`
	Amount        float64     `json:"amount" db:"amount"`
	Rate          float64     `json:"rate" db:"rate"`
	Status        DealStatus  `json:"status" db:"status"`
	Direction     TradeDirection `json:"direction" db:"direction"`
	ValueDate     time.Time   `json:"value_date" db:"value_date"`
	SettlementDate time.Time  `json:"settlement_date" db:"settlement_date"`
	Reference     string      `json:"reference" db:"reference"`
	Notes         string      `json:"notes" db:"notes"`
	CreatedAt     time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at" db:"updated_at"`
}

// DealStatus represents deal status
type DealStatus string

const (
	DealPending    DealStatus = "PENDING"
	DealConfirmed  DealStatus = "CONFIRMED"
	DealSettled    DealStatus = "SETTLED"
	DealCancelled  DealStatus = "CANCELLED"
	DealRejected   DealStatus = "REJECTED"
)

// TradeDirection represents trade direction (BUY or SELL)
type TradeDirection string

const (
	TradeDirectionBuy  TradeDirection = "BUY"
	TradeDirectionSell TradeDirection = "SELL"
)

// RateRequest represents a request to fetch rates
type RateRequest struct {
	FromCurrency string `json:"from_currency" validate:"required,len=3"`
	ToCurrency   string `json:"to_currency" validate:"required,len=3"`
}

// DealRequest represents a request to book a deal
type DealRequest struct {
	ClientID      string         `json:"client_id" validate:"required"`
	FromCurrency  string         `json:"from_currency" validate:"required,len=3"`
	ToCurrency    string         `json:"to_currency" validate:"required,len=3"`
	Amount        float64        `json:"amount" validate:"required,gt=0"`
	Direction     TradeDirection `json:"direction" validate:"required,oneof=BUY SELL"`
	ValueDate     time.Time      `json:"value_date" validate:"required"`
	Reference     string         `json:"reference"`
	Notes         string         `json:"notes"`
}

// RateResponse represents a rate response
type RateResponse struct {
	ID        uuid.UUID  `json:"id"`
	FromCurrency string   `json:"from_currency"`
	ToCurrency   string   `json:"to_currency"`
	Bid         float64   `json:"bid"`
	Ask         float64   `json:"ask"`
	Mid         float64   `json:"mid"`
	Timestamp   time.Time `json:"timestamp"`
	Source      string    `json:"source"`
}

// DealResponse represents a deal response
type DealResponse struct {
	ID            uuid.UUID      `json:"id"`
	ClientID      string         `json:"client_id"`
	TradeID       string         `json:"trade_id"`
	FromCurrency  string         `json:"from_currency"`
	ToCurrency    string         `json:"to_currency"`
	Amount        float64        `json:"amount"`
	Rate          float64        `json:"rate"`
	Status        DealStatus     `json:"status"`
	Direction     TradeDirection `json:"direction"`
	ValueDate     time.Time      `json:"value_date"`
	SettlementDate time.Time     `json:"settlement_date"`
	Reference     string         `json:"reference"`
}

// PaginationRequest represents pagination parameters
type PaginationRequest struct {
	Page  int `json:"page" validate:"required,min=1"`
	Limit int `json:"limit" validate:"required,min=1,max=100"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	Total      int64       `json:"total"`
	TotalPages int64       `json:"total_pages"`
}

// HealthCheckResponse represents health check response
type HealthCheckResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
}
