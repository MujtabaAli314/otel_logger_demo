package usecase

import (
	"context"
	"errors"
	"sync"

	"github.com/oteldemo/service1-frontend/repository"
	"github.com/oteldemo/service1-frontend/types"
)

// ErrUserNotFound is returned when the data service reports the user
// does not exist.
var ErrUserNotFound = errors.New("user not found")

// DashboardUsecase contains the business rules for assembling a user
// dashboard by orchestrating calls to the data and fraud services.
type DashboardUsecase interface {
	GetDashboard(ctx context.Context, params types.GetDashboardParams) (*types.DashboardResponse, error)
}

type dashboardUsecase struct {
	data  repository.DataClient
	fraud repository.FraudClient
}

func NewDashboardUsecase(data repository.DataClient, fraud repository.FraudClient) DashboardUsecase {
	return &dashboardUsecase{data: data, fraud: fraud}
}

func (u *dashboardUsecase) GetDashboard(ctx context.Context, params types.GetDashboardParams) (*types.DashboardResponse, error) {
	// Fan out to the data service for the user profile and transaction
	// history concurrently. The fraud assessment depends on the
	// transactions, so it runs afterwards.
	var (
		user  *types.User
		txs   []types.Transaction
		uErr  error
		tErr  error
	)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		user, uErr = u.data.GetUser(ctx, params.UserID)
	}()
	go func() {
		defer wg.Done()
		txs, tErr = u.data.GetTransactions(ctx, params.UserID)
	}()
	wg.Wait()

	if uErr != nil {
		if repository.IsNotFound(uErr) {
			return nil, ErrUserNotFound
		}
		return nil, uErr
	}
	if tErr != nil {
		return nil, tErr
	}

	assessment, err := u.fraud.AssessFraud(ctx, params.UserID, txs)
	if err != nil {
		return nil, err
	}

	return &types.DashboardResponse{
		User:            *user,
		Transactions:    txs,
		FraudAssessment: *assessment,
	}, nil
}
