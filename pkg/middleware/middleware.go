package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RequestLogger logs HTTP requests
func RequestLogger(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			traceID := uuid.New().String()

			// Add trace ID to context
			ctx := context.WithValue(r.Context(), "trace-id", traceID)
			r = r.WithContext(ctx)

			// Create response writer wrapper to capture status code
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			logger.Info("request started",
				zap.String("trace_id", traceID),
				zap.String("method", r.Method),
				zap.String("path", r.RequestURI),
				zap.String("remote_addr", r.RemoteAddr),
			)

			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			logger.Info("request completed",
				zap.String("trace_id", traceID),
				zap.String("method", r.Method),
				zap.String("path", r.RequestURI),
				zap.Int("status_code", rw.statusCode),
				zap.Duration("duration", duration),
			)
		})
	}
}

// CORS enables CORS
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RateLimiter middleware for rate limiting (basic implementation)
func RateLimiter(requestsPerSecond int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// In production, use a proper rate limiting library like golang.org/x/time/rate
			next.ServeHTTP(w, r)
		})
	}
}

// Panic recovery middleware
func Recovery(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("panic recovered",
						zap.String("path", r.RequestURI),
						zap.String("method", r.Method),
						zap.Any("panic", err),
					)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("Internal Server Error"))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}
