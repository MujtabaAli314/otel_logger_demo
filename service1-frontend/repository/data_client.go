package repository

import (
	"context"
	"fmt"
	"net/http"

	"github.com/oteldemo/service1-frontend/types"
)

// DataClient abstracts the calls the frontend makes to the data service
// (service2) for users and transactions.
type DataClient interface {
	GetUser(ctx context.Context, userID uint) (*types.User, error)
	GetTransactions(ctx context.Context, userID uint) ([]types.Transaction, error)
}

type dataClient struct {
	baseURL string
}

func NewDataClient(baseURL string) DataClient {
	return &dataClient{baseURL: baseURL}
}

func (c *dataClient) GetUser(ctx context.Context, userID uint) (*types.User, error) {
	var user types.User
	url := fmt.Sprintf("%s/api/v1/users/%d", c.baseURL, userID)
	if err := doJSON(ctx, http.MethodGet, url, nil, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *dataClient) GetTransactions(ctx context.Context, userID uint) ([]types.Transaction, error) {
	var txs []types.Transaction
	url := fmt.Sprintf("%s/api/v1/users/%d/transactions", c.baseURL, userID)
	if err := doJSON(ctx, http.MethodGet, url, nil, &txs); err != nil {
		return nil, err
	}
	return txs, nil
}
