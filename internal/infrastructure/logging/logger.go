package logging

import (
	"log/slog"
	"os"

	"github.com/tungnt127/scb-core-v2/internal/usecase/ports"
)

type Logger struct {
	logger *slog.Logger
}

func New() ports.Logger {
	return &Logger{
		logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})),
	}
}

func (l *Logger) Info(msg string, fields map[string]any) {
	l.logger.Info(msg, attrs(fields)...)
}

func (l *Logger) Error(msg string, fields map[string]any) {
	l.logger.Error(msg, attrs(fields)...)
}

func attrs(fields map[string]any) []any {
	if len(fields) == 0 {
		return nil
	}
	args := make([]any, 0, len(fields)*2)
	for key, value := range fields {
		args = append(args, key, value)
	}
	return args
}
