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
)

func main() {
	dataURL := envOrDefault("DATA_SERVICE_URL", "http://localhost:8081")
	fraudURL := envOrDefault("FRAUD_SERVICE_URL", "http://localhost:8082")

	cfg := logger.Config{
		Level: logger.ERRORLVL,
	}

	otelLogger := logger.NewLogger(&cfg)
	otelShutdown, err := otelLogger.Setup(context.Background(), "BFF")
	if err != nil {
		return
	}
	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	dataClient := repository.NewDataClient(dataURL, otelLogger)
	fraudClient := repository.NewFraudClient(fraudURL, otelLogger)

	dashboardUC := usecase.NewDashboardUsecase(dataClient, fraudClient, otelLogger)
	txUC := usecase.NewTransactionUsecase(dataClient, otelLogger)

	app := iris.New()
	// otelhttp.NewHandler(app, "/")
	controller.New(dashboardUC, txUC, otelLogger).Register(app)

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
