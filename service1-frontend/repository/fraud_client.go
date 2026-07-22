package repository

import (
	"context"
	"fmt"
	"net/http"

	"github.com/oteldemo/service1-frontend/types"
)

// FraudClient abstracts the call the frontend makes to the fraud service
// (service3) to assess a set of transactions.
type FraudClient interface {
	AssessFraud(ctx context.Context, userID uint, txs []types.Transaction) (*types.FraudAssessment, error)
}

type assessRequest struct {
	UserID       uint                `json:"user_id"`
	Transactions []types.Transaction `json:"transactions"`
}

type fraudClient struct {
	baseURL string
}

func NewFraudClient(baseURL string) FraudClient {
	return &fraudClient{baseURL: baseURL}
}

func (c *fraudClient) AssessFraud(ctx context.Context, userID uint, txs []types.Transaction) (*types.FraudAssessment, error) {
	var assessment types.FraudAssessment
	url := fmt.Sprintf("%s/api/v1/fraud/assess", c.baseURL)
	body := assessRequest{UserID: userID, Transactions: txs}
	if err := doJSON(ctx, http.MethodPost, url, body, &assessment); err != nil {
		return nil, err
	}
	return &assessment, nil
}
