package repository

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/oteldemo/logger"
	"github.com/oteldemo/service1-frontend/types"
	"go.opentelemetry.io/otel/attribute"
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
	logger  logger.Logger
}

func NewDataClient(baseURL string, logger logger.Logger) DataClient {
	return &dataClient{baseURL: baseURL, logger: logger}
}

func (c *dataClient) GetUser(ctx context.Context, userID uint) (*types.User, error) {
	url := fmt.Sprintf("%s/api/v1/users/%d", c.baseURL, userID)
	// Record to the span created by the controller and threaded down via
	// the context. The repository does not start its own span.
	c.logger.SpanSetAttr(ctx, attribute.String("data_service.get_user.url", url))
	c.logger.LogInfo(ctx, "calling data service", "op", "get_user", "url", url, "user_id", userID)
	var user types.User
	if err := doJSON(ctx, http.MethodGet, url, nil, &user); err != nil {
		return nil, c.logger.LogError(ctx, errors.New("data service get_user failed"), "url", url, "err", err.Error())
	}
	c.logger.SpanSetAttr(ctx, attribute.String("data_service.get_user.email", user.Email))
	c.logger.LogInfo(ctx, "data service returned user", "user_id", user.ID, "email", user.Email)
	return &user, nil
}

func (c *dataClient) GetTransactions(ctx context.Context, userID uint) ([]types.Transaction, error) {
	url := fmt.Sprintf("%s/api/v1/users/%d/transactions", c.baseURL, userID)
	c.logger.SpanSetAttr(ctx, attribute.String("data_service.get_transactions.url", url))
	c.logger.LogInfo(ctx, "calling data service", "op", "get_transactions", "url", url, "user_id", userID)
	var txs []types.Transaction
	if err := doJSON(ctx, http.MethodGet, url, nil, &txs); err != nil {
		return nil, c.logger.LogError(ctx, errors.New("data service get_transactions failed"), "url", url, "err", err.Error())
	}
	c.logger.SpanSetAttr(ctx, attribute.Int("data_service.get_transactions.count", len(txs)))
	c.logger.LogInfo(ctx, "data service returned transactions", "user_id", userID, "count", len(txs))
	return txs, nil
}

func (c *dataClient) CreateTransaction(ctx context.Context, userID uint, params types.CreateTransactionParams) (*types.Transaction, error) {
	url := fmt.Sprintf("%s/api/v1/users/%d/transactions", c.baseURL, userID)
	c.logger.SpanSetAttr(ctx,
		attribute.String("data_service.create_transaction.url", url),
		attribute.Float64("data_service.create_transaction.amount", params.Amount),
	)
	c.logger.LogInfo(ctx, "calling data service", "op", "create_transaction", "url", url, "user_id", userID, "amount", params.Amount)

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
		// return nil, c.logger.LogError( ctx, errors.New("data service create_transaction failed"), "url", url, "err", err.Error())
		return nil, err
	}
	c.logger.SpanSetAttr(ctx, attribute.Int("data_service.create_transaction.tx_id", int(tx.ID)))
	c.logger.LogInfo(ctx, "data service created transaction", "tx_id", tx.ID)
	return &tx, nil
}
