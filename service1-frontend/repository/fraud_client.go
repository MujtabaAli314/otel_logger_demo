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
	logger  slog.Logger
}

func NewFraudClient(baseURL string, logger slog.Logger) FraudClient {
	return &fraudClient{baseURL: baseURL, logger: logger}
}

func (c *fraudClient) AssessFraud(ctx context.Context, userID uint, txs []types.Transaction) (*types.FraudAssessment, error) {
	url := fmt.Sprintf("%s/api/v1/fraud/assess", c.baseURL)
	// Record to the span created by the controller and threaded down via
	// the context. The repository does not start its own span.
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		attribute.String("fraud_service.assess.url", url),
		attribute.Int("fraud_service.assess.tx_count", len(txs)),
	)
	c.logger.InfoContext(ctx, "calling fraud service", "url", url, "user_id", userID, "tx_count", len(txs))
	var assessment types.FraudAssessment
	body := assessRequest{UserID: userID, Transactions: txs}
	if err := doJSON(ctx, http.MethodPost, url, body, &assessment); err != nil {
		span.RecordError(err)
		c.logger.ErrorContext(ctx, "fraud service assess failed", "url", url, "err", err.Error())
		return nil, err
	}
	span.SetAttributes(
		attribute.String("fraud_service.assess.risk_level", assessment.RiskLevel),
		attribute.Int("fraud_service.assess.risk_score", assessment.RiskScore),
	)
	c.logger.InfoContext(ctx, "fraud service returned assessment", "user_id", userID, "risk_level", assessment.RiskLevel, "risk_score", assessment.RiskScore)
	return &assessment, nil
}
