package main

import (
	"log"
	"os"

	"github.com/kataras/iris/v12"
	"github.com/oteldemo/service1-frontend/controller"
	"github.com/oteldemo/service1-frontend/repository"
	"github.com/oteldemo/service1-frontend/usecase"
)

func main() {
	dataURL := envOrDefault("DATA_SERVICE_URL", "http://localhost:8081")
	fraudURL := envOrDefault("FRAUD_SERVICE_URL", "http://localhost:8082")

	dataClient := repository.NewDataClient(dataURL)
	fraudClient := repository.NewFraudClient(fraudURL)

	dashboardUC := usecase.NewDashboardUsecase(dataClient, fraudClient)

	app := iris.New()
	controller.New(dashboardUC).Register(app)

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
