package usecase

import (
	"context"

	"github.com/oteldemo/logger"
	"github.com/oteldemo/service1-frontend/repository"
	"github.com/oteldemo/service1-frontend/types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// TransactionUsecase contains the business rules around creating
// transactions through the data service.
type TransactionUsecase interface {
	CreateTransaction(ctx context.Context, params types.CreateTransactionParams) (*types.Transaction, *logger.Coerr)
}

type transactionUsecase struct {
	data   repository.DataClient
	logger logger.Logger
}

func NewTransactionUsecase(data repository.DataClient, logger logger.Logger) TransactionUsecase {
	return &transactionUsecase{data: data, logger: logger}
}

func (u *transactionUsecase) CreateTransaction(ctx context.Context, params types.CreateTransactionParams) (*types.Transaction, *logger.Coerr) {
	// The span is created by the controller and threaded down via the
	// context. The usecase records to that same span and forwards the
	// context to the repository.
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Int("usecase.user_id", int(params.UserID)))
	u.logger.LogInfo(ctx, "create transaction usecase", "user_id", params.UserID, "amount", params.Amount)

	tx, err := u.data.CreateTransaction(ctx, params.UserID, params)
	if err != nil {
		// no need to log the error here since it is triggered by the repo and the usecase simple returns it
		return nil, &logger.Coerr{
			Msg:  err.Error(),
			Code: "CREATE-TRANSACTION-00",
			Ref:  "CREATE-TRANSACTION-00",
		}
	}
	// the following setattr may be included in the logs only, no need for the span. tx.id can be included in the metadata of
	// the error message in the case of errors.
	span.SetAttributes(attribute.Int("usecase.tx_id", int(tx.ID)))
	u.logger.LogInfo(ctx, "transaction created", "tx_id", tx.ID)
	return tx, nil
}
