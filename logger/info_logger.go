// Package logger provides structured logging primitives shared across the
// oteldemo services.
//
// This is an intentional placeholder: the logging implementation
// (structured logger, trace/log correlation, sinks, etc.) will be added
// in a later step. The package currently exists only so each service can
// declare a dependency on it.
package logger

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
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

// OtelLogger implements the Logger interface
// This is supposed to be initiated with the config.Level = INFOLVL
// However, as you can see, there is no warning here. So do not tell me you were not warned ;)
type OtelLogger struct {
	TracerProvider *trace.TracerProvider
	MeterProvider  *metric.MeterProvider
	LoggerProvider *log.LoggerProvider
}

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func (logger *OtelLogger) Setup(ctx context.Context, serviceName string) (func(context.Context) error, error) {
	var shutdownFuncs []func(context.Context) error
	var err error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Set up propagator.
	prop := logger.NewPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up trace provider.
	tracerProvider, err := logger.NewTracerProvider(ctx, serviceName)
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	// // Set up meter provider.
	// meterProvider, err := logger.NewMeterProvider()
	// if err != nil {
	// 	handleErr(err)
	// 	return shutdown, err
	// }
	// shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	// otel.SetMeterProvider(meterProvider)

	// Set up logger provider.
	loggerProvider, err := logger.NewLoggerProvider(ctx)
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)

	return shutdown, err
}

func (logger *OtelLogger) StartSpan(ctx context.Context, name string) (context.Context, otelTrace.Span) {
	return logger.TracerProvider.Tracer(name).Start(ctx, name)
}

func (logger *OtelLogger) NewPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func (logger *OtelLogger) NewTracerProvider(ctx context.Context, serviceName string) (*trace.TracerProvider, error) {
	// traceExporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	// if err != nil {
	// 	return nil, err
	// }

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
			// Default is 5s. Set to 1s for demonstrative purposes.
			trace.WithBatchTimeout(time.Second)),
		trace.WithResource(res),
	)
	logger.TracerProvider = tracerProvider
	return tracerProvider, nil
}

func (logger *OtelLogger) NewMeterProvider() (*metric.MeterProvider, error) {
	metricExporter, err := stdoutmetric.New(stdoutmetric.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			// Default is 1m. Set to 3s for demonstrative purposes.
			metric.WithInterval(3*time.Second))),
	)
	logger.MeterProvider = meterProvider
	return meterProvider, nil
}

func (logger *OtelLogger) NewLoggerProvider(ctx context.Context) (*log.LoggerProvider, error) {
	// logExporter, err := stdoutlog.New(stdoutlog.WithPrettyPrint())
	// if err != nil {
	// 	return nil, err
	// }

	// loggerProvider := log.NewLoggerProvider(
	// 	log.WithProcessor(log.NewBatchProcessor(logExporter)),
	// )
	// logger.LoggerProvider = loggerProvider
	// return loggerProvider, nil
	exp, err := otlploggrpc.New(ctx,
		otlploggrpc.WithEndpoint("localhost:4317"), // your OTel Collector, not Jaeger, and not Kafka directly
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

// Implementing the logger interface
func (logger *OtelLogger) GetTraceProvider() *trace.TracerProvider {
	return logger.TracerProvider
}

func (logger *OtelLogger) GetLogProvider() *log.LoggerProvider {
	return logger.LoggerProvider
}

func (logger *OtelLogger) GetLogger() *slog.Logger {
	return otelslog.NewLogger("SERVICE1-CONTROLLER-Logger",
		otelslog.WithLoggerProvider(logger.GetLogProvider()),
	)
}

func (l *OtelLogger) LogError(ctx context.Context, err error, args ...any) error {
	// Obviously we need to check the nullability of GetSpanFromContext
	l.GetSpanFromContext(ctx).RecordError(err)
	l.GetLogger().ErrorContext(ctx, err.Error(), args...)
	return err
}

func (l *OtelLogger) LogErrorMsg(ctx context.Context, msg Message) error {
	// Obviously we need to check the nullability of GetSpanFromContext
	l.GetSpanFromContext(ctx).RecordError(errors.New(msg.Stringify()))
	l.GetLogger().ErrorContext(ctx, msg.Stringify(), "meta_info", msg.MetaInfo)
	return errors.New(msg.Stringify())
}

func (l *OtelLogger) LogWarn() {}
func (l *OtelLogger) GetSpanFromContext(ctx context.Context) otelTrace.Span {
	return otelTrace.SpanFromContext(ctx)
}
func (l *OtelLogger) SpanSetAttr(ctx context.Context, attrs ...attribute.KeyValue) {
	l.GetSpanFromContext(ctx).SetAttributes(attrs...)
}
func (l *OtelLogger) LogInfo(ctx context.Context, msg string, args ...any) {
	l.GetLogger().InfoContext(ctx, msg, args...)
}

func (l *OtelLogger) ExeAndLog(span otelTrace.Span, ctx context.Context, callee func(...any) (any, *Coerr), args ...any) (result any, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic calling function: %v", r)
		}
	}()

	fv := reflect.ValueOf(callee)
	if fv.Kind() != reflect.Func {
		return nil, fmt.Errorf("not a function")
	}

	in := make([]reflect.Value, len(args))
	for i, a := range args {
		in[i] = reflect.ValueOf(a)
	}

	out := fv.Call(in)
	if len(out) == 0 {
		return nil, nil
	}

	last := out[len(out)-1]
	if last.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		if !last.IsNil() {
			err = last.Interface().(error)
		}
		if len(out) > 1 {
			result = out[0].Interface()
		}
		return result, err
	}

	return out[0].Interface(), nil
}
