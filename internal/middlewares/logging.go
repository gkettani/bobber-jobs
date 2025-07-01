package middlewares

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/gkettani/bobber-the-swe/internal/logger"
	"github.com/google/uuid"
)

// MiddlewareConfig holds configuration for the HTTP logging middleware
type MiddlewareConfig struct {
	LogRequestBody      bool  `env:"LOG_REQUEST_BODIES" envDefault:"true"`
	LogResponseBody     bool  `env:"LOG_RESPONSE_BODIES" envDefault:"true"`
	MaxRequestBodySize  int64 `env:"LOG_MAX_REQUEST_BODY_SIZE" envDefault:"2048"`
	MaxResponseBodySize int64 `env:"LOG_MAX_RESPONSE_BODY_SIZE" envDefault:"1024"`
	LogHeaders          bool  `env:"LOG_HEADERS" envDefault:"true"`
}

func LoadMiddlewareConfig() *MiddlewareConfig {
	config := &MiddlewareConfig{}
	if err := env.Parse(config); err != nil {
		logger.Error("Failed to parse middleware config", "error", err)
		panic(err)
	}
	return config
}

// RequestID key for context
type contextKey string

const RequestIDKey contextKey = "request_id"

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int64
	body       *bytes.Buffer
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.body != nil && rw.statusCode < 400 {
		// Only capture successful response bodies up to 1KB for logging
		if rw.body.Len() < 1024 {
			rw.body.Write(b)
		}
	}
	size, err := rw.ResponseWriter.Write(b)
	rw.size += int64(size)
	return size, err
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RequestLogData contains all the data we want to log about an HTTP request
type RequestLogData struct {
	RequestID     string            `json:"request_id"`
	Method        string            `json:"method"`
	URL           string            `json:"url"`
	Path          string            `json:"path"`
	RawQuery      string            `json:"raw_query,omitempty"`
	QueryParams   map[string]string `json:"query_params,omitempty"`
	Headers       map[string]string `json:"headers,omitempty"`
	UserAgent     string            `json:"user_agent,omitempty"`
	RemoteIP      string            `json:"remote_ip,omitempty"`
	ContentType   string            `json:"content_type,omitempty"`
	ContentLength int64             `json:"content_length,omitempty"`
	RequestBody   string            `json:"request_body,omitempty"`
	StatusCode    int               `json:"status_code"`
	ResponseSize  int64             `json:"response_size"`
	Duration      string            `json:"duration"`
	DurationMs    float64           `json:"duration_ms"`
	ResponseBody  string            `json:"response_body,omitempty"`
	Error         string            `json:"error,omitempty"`
	Timestamp     time.Time         `json:"timestamp"`
}

// HTTPLoggingMiddleware creates a middleware that logs all HTTP requests with comprehensive details
func HTTPLoggingMiddleware(next http.Handler) http.Handler {
	return HTTPLoggingMiddlewareWithConfig(next, LoadMiddlewareConfig())
}

// HTTPLoggingMiddlewareWithConfig creates a configurable middleware that logs HTTP requests
func HTTPLoggingMiddlewareWithConfig(next http.Handler, config *MiddlewareConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()

		// Generate request ID
		requestID := uuid.New().String()

		// Add request ID to context
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		r = r.WithContext(ctx)

		// Create wrapped response writer
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     200, // Default status code
			body:           &bytes.Buffer{},
		}

		// Capture request body (for POST/PUT/PATCH requests)
		var requestBody string
		if config.LogRequestBody && (r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH") {
			if r.Body != nil {
				bodyBytes, err := io.ReadAll(r.Body)
				if err == nil {
					// Use configured request body size limit
					if int64(len(bodyBytes)) <= config.MaxRequestBodySize {
						requestBody = string(bodyBytes)
					} else {
						requestBody = fmt.Sprintf("[Request body too large: %d bytes]", len(bodyBytes))
					}
					// Restore the body for the actual handler
					r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				}
			}
		}

		// Extract query parameters
		queryParams := make(map[string]string)
		for key, values := range r.URL.Query() {
			if len(values) > 0 {
				queryParams[key] = values[0] // Take first value
			}
		}

		// Extract important headers (excluding sensitive ones) if configured
		var headers map[string]string
		if config.LogHeaders {
			headers = make(map[string]string)
			sensitiveHeaders := map[string]bool{
				"authorization": true,
				"cookie":        true,
				"x-api-key":     true,
				"x-auth-token":  true,
			}

			for name, values := range r.Header {
				lowerName := strings.ToLower(name)
				if !sensitiveHeaders[lowerName] && len(values) > 0 {
					headers[name] = values[0]
				}
			}
		}

		// Get client IP
		remoteIP := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			remoteIP = strings.Split(forwarded, ",")[0]
		} else if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
			remoteIP = realIP
		}

		// Set request ID in response header for tracing
		rw.Header().Set("X-Request-ID", requestID)

		defer func() {
			if err := recover(); err != nil {
				duration := time.Since(startTime)
				logData := RequestLogData{
					RequestID:     requestID,
					Method:        r.Method,
					URL:           r.URL.String(),
					Path:          r.URL.Path,
					RawQuery:      r.URL.RawQuery,
					QueryParams:   queryParams,
					Headers:       headers,
					UserAgent:     r.Header.Get("User-Agent"),
					RemoteIP:      remoteIP,
					ContentType:   r.Header.Get("Content-Type"),
					ContentLength: r.ContentLength,
					RequestBody:   requestBody,
					StatusCode:    500,
					ResponseSize:  rw.size,
					Duration:      duration.String(),
					DurationMs:    float64(duration.Nanoseconds()) / 1e6,
					Error:         fmt.Sprintf("Panic: %v", err),
					Timestamp:     startTime,
				}

				logger.Error("HTTP request panic", "request", logData)

				// Re-panic to let other middleware handle it
				panic(err)
			}
		}()

		next.ServeHTTP(rw, r)

		// Calculate request duration
		duration := time.Since(startTime)

		logData := RequestLogData{
			RequestID:     requestID,
			Method:        r.Method,
			URL:           r.URL.String(),
			Path:          r.URL.Path,
			RawQuery:      r.URL.RawQuery,
			QueryParams:   queryParams,
			Headers:       headers,
			UserAgent:     r.Header.Get("User-Agent"),
			RemoteIP:      remoteIP,
			ContentType:   r.Header.Get("Content-Type"),
			ContentLength: r.ContentLength,
			RequestBody:   requestBody,
			StatusCode:    rw.statusCode,
			ResponseSize:  rw.size,
			Duration:      duration.String(),
			DurationMs:    float64(duration.Nanoseconds()) / 1e6,
			Timestamp:     startTime,
		}

		// Add response body for successful API calls if configured
		if config.LogResponseBody && rw.statusCode < 400 && strings.HasPrefix(r.URL.Path, "/api/") && rw.body.Len() > 0 {
			responseBody := rw.body.String()
			if int64(len(responseBody)) <= config.MaxResponseBodySize {
				logData.ResponseBody = responseBody
			} else {
				logData.ResponseBody = "[Response body too large]"
			}
		}

		// Log the request with appropriate level based on status code
		if rw.statusCode >= 500 {
			logger.Error("HTTP request completed with server error", "request", logData)
		} else if rw.statusCode >= 400 {
			logger.Warn("HTTP request completed with client error", "request", logData)
		} else {
			logger.Info("HTTP request completed", "request", logData)
		}
	})
}

// WrapHandler wraps a single handler with logging middleware
func WrapHandler(handler http.HandlerFunc) http.HandlerFunc {
	return HTTPLoggingMiddleware(http.HandlerFunc(handler)).ServeHTTP
}
