package main

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/kataras/iris/v12"
	"github.com/oteldemo/logger"
	"github.com/oteldemo/service1-frontend/controller"
	"github.com/oteldemo/service1-frontend/repository"
	"github.com/oteldemo/service1-frontend/usecase"
	"go.opentelemetry.io/contrib/bridges/otelslog"
)

func main() {
	dataURL := envOrDefault("DATA_SERVICE_URL", "http://localhost:8081")
	fraudURL := envOrDefault("FRAUD_SERVICE_URL", "http://localhost:8082")

	otelLogger := new(logger.OtelLogger)
	otelShutdown, err := otelLogger.SetupOTelSDK(context.Background(), "BFF")
	if err != nil {
		return
	}
	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()
	traceProvider := otelLogger.TracerProvider
	tracer := traceProvider.Tracer("SERVICE1-CONTROLLER-Tracer")
	logger := otelslog.NewLogger("SERVICE1-CONTROLLER-Logger",
		otelslog.WithLoggerProvider(otelLogger.LoggerProvider),
	)

	dataClient := repository.NewDataClient(dataURL, *logger)
	fraudClient := repository.NewFraudClient(fraudURL, *logger)

	dashboardUC := usecase.NewDashboardUsecase(dataClient, fraudClient, *logger)
	txUC := usecase.NewTransactionUsecase(dataClient, *logger)

	app := iris.New()
	// otelhttp.NewHandler(app, "/")
	controller.New(dashboardUC, txUC, tracer, *logger).Register(app)

	addr := envOrDefault("PORT", "8080")
	log.Printf("service1-frontend listening on :%s (data=%s fraud=%s)", addr, dataURL, fraudURL)
	if err := app.Listen(":" + addr); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
