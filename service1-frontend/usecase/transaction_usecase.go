package usecase

import (
	"context"
	"log/slog"

	"github.com/oteldemo/service1-frontend/repository"
	"github.com/oteldemo/service1-frontend/types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// TransactionUsecase contains the business rules around creating
// transactions through the data service.
type TransactionUsecase interface {
	CreateTransaction(ctx context.Context, params types.CreateTransactionParams) (*types.Transaction, error)
}

type transactionUsecase struct {
	data   repository.DataClient
	logger slog.Logger
}

func NewTransactionUsecase(data repository.DataClient, logger slog.Logger) TransactionUsecase {
	return &transactionUsecase{data: data, logger: logger}
}

func (u *transactionUsecase) CreateTransaction(ctx context.Context, params types.CreateTransactionParams) (*types.Transaction, error) {
	// The span is created by the controller and threaded down via the
	// context. The usecase records to that same span and forwards the
	// context to the repository.
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Int("usecase.user_id", int(params.UserID)))
	u.logger.InfoContext(ctx, "create transaction usecase", "user_id", params.UserID, "amount", params.Amount)

	tx, err := u.data.CreateTransaction(ctx, params.UserID, params)
	if err != nil {
		span.RecordError(err)
		u.logger.ErrorContext(ctx, "usecase create transaction failed", "err", err.Error())
		return nil, err
	}
	span.SetAttributes(attribute.Int("usecase.tx_id", int(tx.ID)))
	u.logger.InfoContext(ctx, "transaction created", "tx_id", tx.ID)
	return tx, nil
}
