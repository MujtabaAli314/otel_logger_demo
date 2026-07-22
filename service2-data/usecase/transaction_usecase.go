package usecase

import (
	"github.com/oteldemo/service2-data/repository"
	"github.com/oteldemo/service2-data/types"
)

const (
	defaultTxLimit = 50
	maxTxLimit     = 200
)

// TransactionUsecase contains the business rules around fetching a
// user's transaction history.
type TransactionUsecase interface {
	ListTransactions(params types.GetTransactionsParams) ([]types.TransactionResponse, error)
}

type transactionUsecase struct {
	txs repository.TransactionRepository
}

func NewTransactionUsecase(txs repository.TransactionRepository) TransactionUsecase {
	return &transactionUsecase{txs: txs}
}

func (u *transactionUsecase) ListTransactions(params types.GetTransactionsParams) ([]types.TransactionResponse, error) {
	limit := params.Limit
	if limit <= 0 {
		limit = defaultTxLimit
	}
	if limit > maxTxLimit {
		limit = maxTxLimit
	}

	txs, err := u.txs.ListByUserID(params.UserID, limit, params.Offset)
	if err != nil {
		return nil, err
	}

	resps := make([]types.TransactionResponse, 0, len(txs))
	for i := range txs {
		resps = append(resps, types.NewTransactionResponse(txs[i]))
	}
	return resps, nil
}
