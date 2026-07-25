package repository

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/oteldemo/service1-frontend/types"
	"go.opentelemetry.io/otel/trace"
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
	tracer  trace.Tracer
	logger  slog.Logger
}

func NewFraudClient(baseURL string, tracer trace.Tracer, logger slog.Logger) FraudClient {
	return &fraudClient{baseURL: baseURL, tracer: tracer, logger: logger}
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
