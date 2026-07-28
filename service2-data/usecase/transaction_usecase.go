package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/oteldemo/service2-data/repository"
	"github.com/oteldemo/service2-data/tracing"
	"github.com/oteldemo/service2-data/types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	defaultTxLimit = 50
	maxTxLimit     = 200
)

// ErrInvalidTransaction is returned when a create request fails
// validation.
var ErrInvalidTransaction = errors.New("invalid transaction")

// TransactionUsecase contains the business rules around a user's
// transaction history.
type TransactionUsecase interface {
	ListTransactions(ctx context.Context, params types.GetTransactionsParams) ([]types.TransactionResponse, error)
	CreateTransaction(ctx context.Context, params types.CreateTransactionParams) (*types.TransactionResponse, error)
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
	// context to the repository. The parent span ID (service1's span) is
	// attached to every log record.
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Int("usecase.user_id", int(params.UserID)))
	log := u.logger.With("parent_span_id", tracing.ParentSpanID(ctx))
	log.InfoContext(ctx, "reached transaction usecase", "user_id", params.UserID)

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
		log.ErrorContext(ctx, "usecase listing transactions failed", "err", err.Error())
		return nil, err
	}

	span.SetAttributes(attribute.Int("usecase.tx_count", len(txs)))
	log.InfoContext(ctx, "usecase fetched transactions", "count", len(txs))

	resps := make([]types.TransactionResponse, 0, len(txs))
	for i := range txs {
		resps = append(resps, types.NewTransactionResponse(txs[i]))
	}
	return resps, nil
}

func (u *transactionUsecase) CreateTransaction(ctx context.Context, params types.CreateTransactionParams) (*types.TransactionResponse, error) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Int("usecase.user_id", int(params.UserID)))
	log := u.logger.With("parent_span_id", tracing.ParentSpanID(ctx))
	log.InfoContext(ctx, "create transaction usecase", "user_id", params.UserID, "amount", params.Amount)

	if err := validateCreate(params); err != nil {
		span.RecordError(err)
		log.ErrorContext(ctx, "validation failed", "err", err.Error())
		return nil, fmt.Errorf("%w: %s", ErrInvalidTransaction, err.Error())
	}

	tx := &types.Transaction{
		UserID:      params.UserID,
		Amount:      params.Amount,
		Currency:    params.Currency,
		Type:        params.Type,
		Merchant:    params.Merchant,
		Description: params.Description,
	}
	created, err := u.txs.Create(ctx, tx)
	if err != nil {
		span.RecordError(err)
		log.ErrorContext(ctx, "usecase create transaction failed", "err", err.Error())
		return nil, err
	}

	resp := types.NewTransactionResponse(*created)
	span.SetAttributes(attribute.Int("usecase.tx_id", int(created.ID)))
	log.InfoContext(ctx, "transaction created", "tx_id", created.ID)
	return &resp, nil
}

func validateCreate(p types.CreateTransactionParams) error {
	if p.UserID == 0 {
		return errors.New("user_id is required")
	}
	if p.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}
	if len(p.Currency) != 3 {
		return errors.New("currency must be a 3-letter code")
	}
	if p.Type != types.TransactionTypeDebit && p.Type != types.TransactionTypeCredit {
		return errors.New("type must be debit or credit")
	}
	return nil
}
