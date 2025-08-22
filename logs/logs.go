package logs

import (
	"context"
	"log/slog"
	"os"
)

type Logger interface {
	Info(message string, args ...interface{})
	Debug(message string, args ...interface{})
	Error(message string, args ...interface{})
	Warn(message string, args ...interface{})
}

type slogLogger struct {
	Logger
	ctx    context.Context
	logger *slog.Logger
}

func FromContext(ctx context.Context) Logger {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug, // Asegúrate de establecer el nivel adecuado
	})
	logger := slog.New(handler)
	return &slogLogger{logger: logger, ctx: ctx}
}

func (log *slogLogger) Debug(message string, args ...interface{}) {
	log.logger.Log(log.ctx, slog.LevelDebug.Level(), message, args...)
}

func (log *slogLogger) Error(message string, args ...interface{}) {
	log.logger.Log(log.ctx, slog.LevelError.Level(), message, args...)
}

func (log *slogLogger) Info(message string, args ...interface{}) {
	log.logger.Log(log.ctx, slog.LevelInfo.Level(), message, args...)
}

func (log *slogLogger) Warn(message string, args ...interface{}) {
	log.logger.Log(log.ctx, slog.LevelWarn.Level(), message, args...)
}
