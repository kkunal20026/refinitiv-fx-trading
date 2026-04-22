package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/refinitiv/fx-trading/internal/client"
	"github.com/refinitiv/fx-trading/internal/config"
	"github.com/refinitiv/fx-trading/internal/database"
	"github.com/refinitiv/fx-trading/internal/handler"
	"github.com/refinitiv/fx-trading/internal/repository"
	"github.com/refinitiv/fx-trading/internal/service"
	"github.com/refinitiv/fx-trading/pkg/logger"
	"github.com/refinitiv/fx-trading/pkg/middleware"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := logger.New(logger.Config{
		Level:      cfg.Logger.Level,
		Encoding:   cfg.Logger.Encoding,
		OutputPath: cfg.Logger.OutputPath,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	zapLog := log.Logger
	zapLog.Info("Starting Refinitiv FX Trading Service",
		zap.String("version", "1.0.0"),
		zap.String("environment", cfg.Server.Environment),
	)

	// Initialize database
	db, err := database.New(cfg.Database, zapLog)
	if err != nil {
		zapLog.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer database.Close(db, zapLog)

	// Initialize database tables
	if err := database.InitializeTables(db, zapLog); err != nil {
		zapLog.Fatal("Failed to initialize database tables", zap.Error(err))
	}

	// Initialize Refinitiv client
	refinitivClient := client.New(
		cfg.Refinitiv.BaseURL,
		cfg.Refinitiv.Username,
		cfg.Refinitiv.Password,
		cfg.Refinitiv.Timeout,
		cfg.Refinitiv.MaxRetries,
		cfg.Refinitiv.RetryBackoff,
		zapLog,
	)

	// Initialize repositories
	rateRepo := repository.NewRateRepository(db, zapLog)
	dealRepo := repository.NewDealRepository(db, zapLog)

	// Initialize services
	rateService := service.NewRateService(rateRepo, refinitivClient, zapLog)
	dealService := service.NewDealService(dealRepo, rateRepo, zapLog)

	// Initialize handlers
	h := handler.New(rateService, dealService, zapLog)

	// Setup router
	router := chi.NewRouter()

	// Middleware
	router.Use(middleware.Recovery(zapLog))
	router.Use(middleware.RequestLogger(zapLog))
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	router.Get("/health", h.HealthCheck)

	// Rate endpoints
	router.Get("/api/v1/rates", h.GetRate)
	router.Get("/api/v1/rates/{id}", h.GetRateByID)

	// Deal endpoints
	router.Post("/api/v1/deals", h.BookDeal)
	router.Get("/api/v1/deals", h.ListDeals)
	router.Get("/api/v1/deals/{id}", h.GetDeal)
	router.Get("/api/v1/deals/trade/{trade_id}", h.GetDealByTradeID)
	router.Put("/api/v1/deals/{id}/status", h.UpdateDealStatus)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Info("Server starting",
			zap.String("address", server.Addr),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server error", zap.Error(err))
		}
	}()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	log.Info("Shutdown signal received, initiating graceful shutdown")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("Server shutdown error", zap.Error(err))
	}

	log.Info("Server stopped gracefully")
}

// Helper functions for logging
func WithString(key, value string) interface{} {
	return fmt.Sprintf("%s=%s", key, value)
}

func WithError(err error) interface{} {
	return err
}
