package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

func getLogLevel() slog.Level {
	levelStr := strings.ToUpper(os.Getenv("LOG_LEVEL"))

	switch levelStr {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

var logger *slog.Logger = slog.New(slog.NewJSONHandler(
	os.Stdout,
	&slog.HandlerOptions{
		Level: getLogLevel(),
	},
))

func Info(msg string, args ...any) {
	logger.Info(msg, args...)
}

func Error(msg string, args ...any) {
	logger.Error(msg, args...)
}

func Debug(msg string, args ...any) {
	logger.Debug(msg, args...)
}

func Warn(msg string, args ...any) {
	logger.Warn(msg, args...)
}

func Fatal(msg string, args ...any) {
	logger.Error(msg, args...)
	os.Exit(1)
}

const RequestIDKey = "request_id"

// GetRequestID extracts the request ID from the context in the context of an http request
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// LogWithRequestID logs a message with the request ID from context in the context of an http request
func LogWithRequestID(ctx context.Context, level string, msg string, args ...any) {
	requestID := GetRequestID(ctx)
	if requestID != "" {
		args = append(args, "request_id", requestID)
	}

	switch strings.ToLower(level) {
	case "debug":
		logger.Debug(msg, args...)
	case "info":
		logger.Info(msg, args...)
	case "warn":
		logger.Warn(msg, args...)
	case "error":
		Error(msg, args...)
	default:
		Info(msg, args...)
	}
}
