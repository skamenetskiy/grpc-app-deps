package log

import (
	"context"
	"io"
	"log/slog"
	"os"
)

func Debug(ctx context.Context, msg string, args ...any) {
	logger.DebugContext(ctx, msg, args...)
}

func Info(ctx context.Context, msg string, args ...any) {
	logger.InfoContext(ctx, msg, args...)
}

func Warn(ctx context.Context, msg string, args ...any) {
	logger.WarnContext(ctx, msg, args...)
}

func Error(ctx context.Context, msg string, args ...any) {
	logger.ErrorContext(ctx, msg, args...)
}

func Fatal(ctx context.Context, msg string, args ...any) {
	logger.ErrorContext(ctx, msg, args...)
	os.Exit(1)
}

func SetLogLevel(level slog.Level) {
	leveler.Set(level)
}

func GetLogLevel() slog.Level {
	return leveler.Level()
}

var logger, leveler = initLogger(os.Stderr)

func initLogger(w io.Writer) (*slog.Logger, *slog.LevelVar) {
	lvl := new(slog.LevelVar)
	lvl.Set(slog.LevelInfo)
	log := slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
		AddSource:   false,
		Level:       lvl,
		ReplaceAttr: nil,
	}))
	return log, lvl
}
