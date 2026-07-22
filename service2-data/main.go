package main

import (
	"log"
	"os"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/oteldemo/service2-data/controller"
	"github.com/oteldemo/service2-data/repository"
	"github.com/oteldemo/service2-data/types"
	"github.com/oteldemo/service2-data/usecase"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	db := mustConnectDB()
	migrate(db)
	seed(db)

	userRepo := repository.NewUserRepository(db)
	txRepo := repository.NewTransactionRepository(db)

	userUC := usecase.NewUserUsecase(userRepo)
	txUC := usecase.NewTransactionUsecase(txRepo)

	app := iris.New()
	controller.NewController(userUC, txUC).Register(app)

	addr := envOrDefault("PORT", "8081")
	log.Printf("service2-data listening on :%s", addr)
	if err := app.Listen(":" + addr); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func mustConnectDB() *gorm.DB {
	host := envOrDefault("DB_HOST", "localhost")
	port := envOrDefault("DB_PORT", "5433")
	user := envOrDefault("DB_USER", "oteldemo")
	pass := envOrDefault("DB_PASS", "oteldemo")
	name := envOrDefault("DB_NAME", "oteldemo")

	dsn := "host=" + host + " port=" + port + " user=" + user +
		" password=" + pass + " dbname=" + name + " sslmode=disable TimeZone=UTC"

	gormLogLevel := logger.Silent
	if os.Getenv("DB_LOG") == "1" {
		gormLogLevel = logger.Info
	}

	var db *gorm.DB
	var err error
	for attempt := 1; attempt <= 30; attempt++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(gormLogLevel)})
		if err == nil {
			if sqlDB, pingErr := db.DB(); pingErr == nil {
				if pErr := sqlDB.Ping(); pErr == nil {
					log.Println("connected to postgres")
					return db
				}
				_ = sqlDB.Close()
			}
		}
		log.Printf("waiting for postgres (%d/30): %v", attempt, err)
		time.Sleep(time.Second)
	}
	log.Fatalf("could not connect to postgres: %v", err)
	return nil
}

func migrate(db *gorm.DB) {
	if err := db.AutoMigrate(&types.User{}, &types.Transaction{}); err != nil {
		log.Fatalf("automigrate failed: %v", err)
	}
	log.Println("migration complete")
}

func seed(db *gorm.DB) {
	var userCount int64
	if err := db.Model(&types.User{}).Count(&userCount).Error; err != nil {
		log.Fatalf("seed check failed: %v", err)
	}
	if userCount > 0 {
		return
	}

	users := []types.User{
		{Name: "Alice Anderson", Email: "alice@example.com"},
		{Name: "Bob Brown", Email: "bob@example.com"},
		{Name: "Carol Clarke", Email: "carol@example.com"},
	}
	if err := db.Create(&users).Error; err != nil {
		log.Fatalf("seed users failed: %v", err)
	}

	now := time.Now().UTC()
	txs := []types.Transaction{
		{UserID: users[0].ID, Amount: 1200.00, Currency: "USD", Type: types.TransactionTypeCredit, Merchant: "Acme Payroll", Description: "Monthly salary", CreatedAt: now.Add(-48 * time.Hour)},
		{UserID: users[0].ID, Amount: 49.99, Currency: "USD", Type: types.TransactionTypeDebit, Merchant: "StreamFlix", Description: "Subscription", CreatedAt: now.Add(-24 * time.Hour)},
		{UserID: users[0].ID, Amount: 9800.00, Currency: "USD", Type: types.TransactionTypeDebit, Merchant: "Wire Transfer LTD", Description: "International transfer", CreatedAt: now.Add(-2 * time.Hour)},
		{UserID: users[0].ID, Amount: 12.50, Currency: "USD", Type: types.TransactionTypeDebit, Merchant: "Coffee Corner", Description: "Latte", CreatedAt: now.Add(-1 * time.Hour)},
		{UserID: users[0].ID, Amount: 15000.00, Currency: "USD", Type: types.TransactionTypeDebit, Merchant: "Wire Transfer LTD", Description: "Suspicious large transfer", CreatedAt: now.Add(-30 * time.Minute)},

		{UserID: users[1].ID, Amount: 3200.00, Currency: "USD", Type: types.TransactionTypeCredit, Merchant: "Acme Payroll", Description: "Monthly salary", CreatedAt: now.Add(-48 * time.Hour)},
		{UserID: users[1].ID, Amount: 89.00, Currency: "USD", Type: types.TransactionTypeDebit, Merchant: "PowerGrid Co", Description: "Electricity bill", CreatedAt: now.Add(-20 * time.Hour)},

		{UserID: users[2].ID, Amount: 2500.00, Currency: "USD", Type: types.TransactionTypeCredit, Merchant: "Acme Payroll", Description: "Monthly salary", CreatedAt: now.Add(-72 * time.Hour)},
		{UserID: users[2].ID, Amount: 7.99, Currency: "USD", Type: types.TransactionTypeDebit, Merchant: "Tuneify", Description: "Music subscription", CreatedAt: now.Add(-10 * time.Hour)},
	}
	if err := db.Create(&txs).Error; err != nil {
		log.Fatalf("seed transactions failed: %v", err)
	}
	log.Printf("seeded %d users and %d transactions", len(users), len(txs))
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
