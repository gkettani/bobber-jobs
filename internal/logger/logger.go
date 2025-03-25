package logger

import (
	"log/slog"
	"os"
)

// var logger *slog.Logger = slog.New(slog.NewJSONHandler(
// 	func() *os.File {
// 		f, err := os.OpenFile("file.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
// 		if err != nil {
// 			panic(err)
// 		}
// 		return f
// 	}(),
// 	&slog.HandlerOptions{
// 		Level: slog.LevelDebug,
// 	}))

var logger *slog.Logger = slog.New(slog.NewJSONHandler(
	os.Stdout,
	&slog.HandlerOptions{
		Level: slog.LevelInfo,
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
