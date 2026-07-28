package logger

// Here we define the abstract entity of the logger
// The actual logger is to be initiated during registry based on the config
// i.e. Logger_1 (the lowest level logging the errors only)
// i.e. Logger_2 (Warnings)
// i.e. Logger_3 (Infos)
// Probably you can come up with better names (or better structure). This is, however, how far my brain helped me reach
// Sorry for that (or you are welcome)

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/sdk/log"
	traceSdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type Logger interface {
	GetTraceProvider() *traceSdk.TracerProvider
	GetLogProvider() *log.LoggerProvider
	LogError(span trace.Span, logger *slog.Logger, ctx context.Context, err error, args ...any) error
	LogWarn(span trace.Span, logger *slog.Logger)
	LogInfo(logger *slog.Logger, ctx context.Context, msg string, args ...any)
}

func NewLogger(cfg *Config) Logger {
	return new(OtelLogger)
}
