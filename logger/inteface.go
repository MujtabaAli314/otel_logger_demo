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
	Setup(context.Context, string) (func(context.Context) error, error)
	GetTraceProvider() *traceSdk.TracerProvider
	GetLogProvider() *log.LoggerProvider
	GetLogger() *slog.Logger
	StartSpan(ctx context.Context, name string) (context.Context, trace.Span)
	LogError(span trace.Span, ctx context.Context, err error, args ...any) error
	LogErrorMsg(span trace.Span, ctx context.Context, msg Message) error
	LogWarn(span trace.Span)
	LogInfo(ctx context.Context, msg string, args ...any)
}

func NewLogger(cfg *Config) Logger {
	if cfg == nil {
		return nil
	}
	if cfg.Level == INFOLVL {
		return new(OtelLogger)
	} else if cfg.Level == ERRORLVL {
		return new(ErrorLvlLoger)
	}
	return new(OtelLogger)
}

type MsgType int

const (
	ERROR_MSG_TYPE = iota
	WARN_MSG_TYPE
	INTO_MSG_TYPE
)

type Message struct {
	Type     MsgType        `json:"Type"`
	Msg      string         `json:"Message"`
	Code     string         `json:"Code"`
	Ref      string         `json:"Ref"`
	MetaInfo map[string]any `json:"MetaInfo"`
}

func (msg Message) Stringify() string {
	return "Code: " + msg.Code +
		"\tRef: " + msg.Ref +
		"\tMsg: " + msg.Msg
}
