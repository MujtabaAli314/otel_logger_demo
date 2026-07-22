package main

import (
	"log"
	"os"

	"github.com/kataras/iris/v12"
	"github.com/oteldemo/service3-fraud/controller"
	"github.com/oteldemo/service3-fraud/repository"
	"github.com/oteldemo/service3-fraud/usecase"
)

func main() {
	ruleRepo := repository.NewRuleRepository()
	fraudUC := usecase.NewFraudUsecase(ruleRepo)

	app := iris.New()
	controller.New(fraudUC).Register(app)

	addr := envOrDefault("PORT", "8082")
	log.Printf("service3-fraud listening on :%s", addr)
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
