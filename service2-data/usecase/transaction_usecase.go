package usecase

import (
	"context"
	"log/slog"

	"github.com/oteldemo/service2-data/repository"
	"github.com/oteldemo/service2-data/types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	defaultTxLimit = 50
	maxTxLimit     = 200
)

// TransactionUsecase contains the business rules around fetching a
// user's transaction history.
type TransactionUsecase interface {
	ListTransactions(ctx context.Context, params types.GetTransactionsParams) ([]types.TransactionResponse, error)
}

type transactionUsecase struct {
	txs    repository.TransactionRepository
	logger slog.Logger
}

func NewTransactionUsecase(txs repository.TransactionRepository, logger slog.Logger) TransactionUsecase {
	return &transactionUsecase{txs: txs, logger: logger}
}

func (u *transactionUsecase) ListTransactions(ctx context.Context, params types.GetTransactionsParams) ([]types.TransactionResponse, error) {
	// The span is created by the controller and threaded down via the
	// context. The usecase records to that same span and forwards the
	// context to the repository.
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Int("usecase.user_id", int(params.UserID)))
	u.logger.InfoContext(ctx, "reached transaction usecase", "user_id", params.UserID)

	limit := params.Limit
	if limit <= 0 {
		limit = defaultTxLimit
	}
	if limit > maxTxLimit {
		limit = maxTxLimit
	}
	span.SetAttributes(attribute.Int("usecase.limit", limit), attribute.Int("usecase.offset", params.Offset))

	txs, err := u.txs.ListByUserID(ctx, params.UserID, limit, params.Offset)
	if err != nil {
		span.RecordError(err)
		u.logger.ErrorContext(ctx, "usecase listing transactions failed", "err", err.Error())
		return nil, err
	}

	span.SetAttributes(attribute.Int("usecase.tx_count", len(txs)))
	u.logger.InfoContext(ctx, "usecase fetched transactions", "count", len(txs))

	resps := make([]types.TransactionResponse, 0, len(txs))
	for i := range txs {
		resps = append(resps, types.NewTransactionResponse(txs[i]))
	}
	return resps, nil
}
