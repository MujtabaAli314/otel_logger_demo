package repository

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/oteldemo/service1-frontend/types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DataClient abstracts the calls the frontend makes to the data service
// (service2) for users and transactions.
type DataClient interface {
	GetUser(ctx context.Context, userID uint) (*types.User, error)
	GetTransactions(ctx context.Context, userID uint) ([]types.Transaction, error)
	CreateTransaction(ctx context.Context, userID uint, params types.CreateTransactionParams) (*types.Transaction, error)
}

type dataClient struct {
	baseURL string
	logger  slog.Logger
}

func NewDataClient(baseURL string, logger slog.Logger) DataClient {
	return &dataClient{baseURL: baseURL, logger: logger}
}

func (c *dataClient) GetUser(ctx context.Context, userID uint) (*types.User, error) {
	url := fmt.Sprintf("%s/api/v1/users/%d", c.baseURL, userID)
	// Record to the span created by the controller and threaded down via
	// the context. The repository does not start its own span.
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("data_service.get_user.url", url))
	c.logger.InfoContext(ctx, "calling data service", "op", "get_user", "url", url, "user_id", userID)
	var user types.User
	if err := doJSON(ctx, http.MethodGet, url, nil, &user); err != nil {
		span.RecordError(err)
		c.logger.ErrorContext(ctx, "data service get_user failed", "url", url, "err", err.Error())
		return nil, err
	}
	span.SetAttributes(attribute.String("data_service.get_user.email", user.Email))
	c.logger.InfoContext(ctx, "data service returned user", "user_id", user.ID, "email", user.Email)
	return &user, nil
}

func (c *dataClient) GetTransactions(ctx context.Context, userID uint) ([]types.Transaction, error) {
	url := fmt.Sprintf("%s/api/v1/users/%d/transactions", c.baseURL, userID)
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("data_service.get_transactions.url", url))
	c.logger.InfoContext(ctx, "calling data service", "op", "get_transactions", "url", url, "user_id", userID)
	var txs []types.Transaction
	if err := doJSON(ctx, http.MethodGet, url, nil, &txs); err != nil {
		span.RecordError(err)
		c.logger.ErrorContext(ctx, "data service get_transactions failed", "url", url, "err", err.Error())
		return nil, err
	}
	span.SetAttributes(attribute.Int("data_service.get_transactions.count", len(txs)))
	c.logger.InfoContext(ctx, "data service returned transactions", "user_id", userID, "count", len(txs))
	return txs, nil
}

func (c *dataClient) CreateTransaction(ctx context.Context, userID uint, params types.CreateTransactionParams) (*types.Transaction, error) {
	url := fmt.Sprintf("%s/api/v1/users/%d/transactions", c.baseURL, userID)
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("data_service.create_transaction.url", url),
		attribute.Float64("data_service.create_transaction.amount", params.Amount),
	)
	c.logger.InfoContext(ctx, "calling data service", "op", "create_transaction", "url", url, "user_id", userID, "amount", params.Amount)

	body := struct {
		Amount      float64 `json:"amount"`
		Currency    string  `json:"currency"`
		Type        string  `json:"type"`
		Merchant    string  `json:"merchant"`
		Description string  `json:"description"`
	}{
		Amount:      params.Amount,
		Currency:    params.Currency,
		Type:        params.Type,
		Merchant:    params.Merchant,
		Description: params.Description,
	}
	var tx types.Transaction
	if err := doJSON(ctx, http.MethodPost, url, body, &tx); err != nil {
		span.RecordError(err)
		c.logger.ErrorContext(ctx, "data service create_transaction failed", "url", url, "err", err.Error())
		return nil, err
	}
	span.SetAttributes(attribute.Int("data_service.create_transaction.tx_id", int(tx.ID)))
	c.logger.InfoContext(ctx, "data service created transaction", "tx_id", tx.ID)
	return &tx, nil
}
