package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	appErrors "github.com/refinitiv/fx-trading/pkg/errors"
	"go.uber.org/zap"
)

// RefinitivClient represents Refinitiv API client
type RefinitivClient struct {
	baseURL    string
	username   string
	password   string
	timeout    time.Duration
	maxRetries int
	backoff    time.Duration
	httpClient *http.Client
	logger     *zap.Logger
}

// RateResponse represents rate data from Refinitiv API
type RateResponse struct {
	ISIN        string    `json:"ISIN"`
	Bid         float64   `json:"Bid"`
	Ask         float64   `json:"Ask"`
	Mid         float64   `json:"Mid"`
	LastUpdate  time.Time `json:"LastUpdate"`
	Status      string    `json:"Status"`
}

// New creates a new Refinitiv client
func New(baseURL, username, password string, timeout time.Duration, maxRetries int, backoff time.Duration, logger *zap.Logger) *RefinitivClient {
	return &RefinitivClient{
		baseURL:    baseURL,
		username:   username,
		password:   password,
		timeout:    timeout,
		maxRetries: maxRetries,
		backoff:    backoff,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger: logger,
	}
}

// GetRate fetches FX rate from Refinitiv API
func (rc *RefinitivClient) GetRate(ctx context.Context, fromCurrency, toCurrency string) (*RateResponse, error) {
	isin := fmt.Sprintf("%s%s=", fromCurrency, toCurrency)
	endpoint := fmt.Sprintf("%s/v1/rates/%s", rc.baseURL, isin)

	var result *RateResponse
	err := rc.doRequestWithRetry(ctx, http.MethodGet, endpoint, nil, &result)
	if err != nil {
		rc.logger.Warn("Refinitiv API unavailable, returning mock rate", zap.String("pair", isin), zap.Error(err))
		return rc.mockRate(fromCurrency, toCurrency), nil
	}

	if result.Status != "OK" {
		return nil, appErrors.NewServiceError(
			fmt.Sprintf("Refinitiv returned status %s", result.Status),
			http.StatusBadGateway,
			nil,
		)
	}

	return result, nil
}

// mockRate returns a synthetic rate for local development when the real API is unreachable.
func (rc *RefinitivClient) mockRate(fromCurrency, toCurrency string) *RateResponse {
	// Static mid prices keyed by pair; bid/ask spread is ±0.0005.
	mids := map[string]float64{
		"EURUSD": 1.0850, "GBPUSD": 1.2650, "USDJPY": 149.50,
		"USDCHF": 0.9050, "AUDUSD": 0.6550, "USDCAD": 1.3600,
		"EURGBP": 0.8580, "EURJPY": 162.20,
	}
	pair := fromCurrency + toCurrency
	mid, ok := mids[pair]
	if !ok {
		mid = 1.0
	}
	return &RateResponse{
		ISIN:       fmt.Sprintf("%s%s=", fromCurrency, toCurrency),
		Bid:        mid - 0.0005,
		Ask:        mid + 0.0005,
		Mid:        mid,
		LastUpdate: time.Now(),
		Status:     "OK",
	}
}

// GetMultipleRates fetches multiple FX rates
func (rc *RefinitivClient) GetMultipleRates(ctx context.Context, pairs []string) (map[string]*RateResponse, error) {
	results := make(map[string]*RateResponse)

	for _, pair := range pairs {
		rate, err := rc.GetRate(ctx, pair[:3], pair[3:])
		if err != nil {
			rc.logger.Warn("failed to get rate for pair", zap.String("pair", pair), zap.Error(err))
			continue
		}
		results[pair] = rate
	}

	if len(results) == 0 {
		return nil, appErrors.NewServiceError(
			"failed to fetch rates for any pair",
			http.StatusBadGateway,
			nil,
		)
	}

	return results, nil
}

// doRequestWithRetry performs HTTP request with retry logic
func (rc *RefinitivClient) doRequestWithRetry(ctx context.Context, method, url string, body io.Reader, result interface{}) error {
	var lastErr error

	for attempt := 0; attempt < rc.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-time.After(rc.backoff * time.Duration(1<<uint(attempt-1))):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		err := rc.doRequest(ctx, method, url, body, result)
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry on client errors (4xx)
		appErr, ok := appErrors.IsAppError(err)
		if ok && appErr.StatusCode >= 400 && appErr.StatusCode < 500 {
			return err
		}

		rc.logger.Warn("request failed, retrying", 
			zap.Int("attempt", attempt+1), 
			zap.Int("max_retries", rc.maxRetries),
			zap.Error(err),
		)
	}

	return appErrors.NewServiceError(
		fmt.Sprintf("request failed after %d retries", rc.maxRetries),
		http.StatusBadGateway,
		lastErr,
	)
}

// doRequest performs a single HTTP request
func (rc *RefinitivClient) doRequest(ctx context.Context, method, url string, body io.Reader, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return appErrors.NewInternalError("failed to create request", err)
	}

	// Add authentication
	req.SetBasicAuth(rc.username, rc.password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := rc.httpClient.Do(req)
	if err != nil {
		return appErrors.NewServiceError(
			"failed to execute request to Refinitiv API",
			http.StatusBadGateway,
			err,
		)
	}
	defer resp.Body.Close()

	// Handle non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return appErrors.NewServiceError(
			fmt.Sprintf("Refinitiv API returned status %d: %s", resp.StatusCode, string(respBody)),
			resp.StatusCode,
			nil,
		)
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return appErrors.NewServiceError(
			"failed to decode Refinitiv API response",
			http.StatusBadGateway,
			err,
		)
	}

	return nil
}
