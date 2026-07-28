package logger

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	otelTrace "go.opentelemetry.io/otel/trace"
)

type ErrorLvlLoger struct {
	TracerProvider *trace.TracerProvider
	MeterProvider  *metric.MeterProvider
	LoggerProvider *log.LoggerProvider
}

func (logger *ErrorLvlLoger) SetupOTelSDK(ctx context.Context, serviceName string) (func(context.Context) error, error) {
	var shutdownFuncs []func(context.Context) error
	var err error

	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	prop := logger.NewPropagator()
	otel.SetTextMapPropagator(prop)

	tracerProvider, err := logger.NewTracerProvider(ctx, serviceName)
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	loggerProvider, err := logger.NewLoggerProvider(ctx)
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)

	return shutdown, err
}

func (logger *ErrorLvlLoger) NewPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func (logger *ErrorLvlLoger) NewTracerProvider(ctx context.Context, serviceName string) (*trace.TracerProvider, error) {

	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint("localhost:4317"),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(attribute.String("service.name", serviceName)),
	)
	if err != nil {
		return nil, err
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter,
			trace.WithBatchTimeout(time.Second)),
		trace.WithResource(res),
	)
	logger.TracerProvider = tracerProvider
	return tracerProvider, nil
}

func (logger *ErrorLvlLoger) NewMeterProvider() (*metric.MeterProvider, error) {
	metricExporter, err := stdoutmetric.New(stdoutmetric.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			metric.WithInterval(3*time.Second))),
	)
	logger.MeterProvider = meterProvider
	return meterProvider, nil
}

func (logger *ErrorLvlLoger) NewLoggerProvider(ctx context.Context) (*log.LoggerProvider, error) {

	exp, err := otlploggrpc.New(ctx,
		otlploggrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(attribute.String("service.name", "BFF-LOGGER")),
	)
	if err != nil {
		return nil, err
	}

	logProvider := log.NewLoggerProvider(
		log.WithResource(res),
		log.WithProcessor(log.NewBatchProcessor(exp)),
	)

	logger.LoggerProvider = logProvider

	slog.SetDefault(otelslog.NewLogger("BFF-LOGGER"))

	return logProvider, nil
}

func (logger *ErrorLvlLoger) GetTraceProvider() *trace.TracerProvider {
	return logger.TracerProvider
}

func (logger *ErrorLvlLoger) GetLogProvider() *log.LoggerProvider {
	return logger.LoggerProvider
}

func (l *ErrorLvlLoger) LogError(span otelTrace.Span, logger *slog.Logger, ctx context.Context, err error, args ...any) error {
	span.RecordError(err)
	logger.ErrorContext(ctx, err.Error(), args...)
	return err
}
func (l *ErrorLvlLoger) LogWarn(span otelTrace.Span, logger *slog.Logger) {
	return
}
func (l *ErrorLvlLoger) LogInfo(logger *slog.Logger, ctx context.Context, err error, args ...any) {
	return
}
