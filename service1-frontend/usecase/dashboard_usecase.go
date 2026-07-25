package usecase

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/oteldemo/service1-frontend/repository"
	"github.com/oteldemo/service1-frontend/types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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
	data   repository.DataClient
	fraud  repository.FraudClient
	tracer trace.Tracer
	logger slog.Logger
}

func NewDashboardUsecase(data repository.DataClient, fraud repository.FraudClient, tracer trace.Tracer, logger slog.Logger) DashboardUsecase {
	return &dashboardUsecase{data: data, fraud: fraud, tracer: tracer, logger: logger}
}

func (u *dashboardUsecase) GetDashboard(ctx context.Context, params types.GetDashboardParams) (*types.DashboardResponse, error) {
	spanCtx, span := u.tracer.Start(ctx, "GetDashboardSpanUsecase")
	defer span.End()
	u.logger.InfoContext(spanCtx, "Heeeey we reached the usecase!!!!!")
	span.SetAttributes(attribute.KeyValue{Key: "UNIQUEKEY", Value: attribute.StringValue("UNIQUEVALUE")})
	// Fan out to the data service for the user profile and transaction
	// history concurrently. The fraud assessment depends on the
	// transactions, so it runs afterwards.
	var (
		user *types.User
		txs  []types.Transaction
		uErr error
		tErr error
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
			// For the following error I think it is better recorded within the repo (See the comment below)
			u.logger.ErrorContext(spanCtx, "usecase calling repo isnotfound and got an error")
			span.SetAttributes(attribute.KeyValue{Key: "ERROR-01", Value: attribute.StringValue(uErr.Error())})
			return nil, ErrUserNotFound
		}
		// Here is the opposite case where there is no call to the repo so we should register the error here
		u.logger.ErrorContext(spanCtx, "usecase throwing an error"+uErr.Error())
		return nil, uErr
	}
	if tErr != nil {
		u.logger.ErrorContext(spanCtx, "usecase throwing an error"+tErr.Error())
		return nil, tErr
	}

	assessment, err := u.fraud.AssessFraud(ctx, params.UserID, txs)
	if err != nil {
		u.logger.ErrorContext(spanCtx, "usecase throwing an error"+err.Error())
		return nil, err
	}

	return &types.DashboardResponse{
		User:            *user,
		Transactions:    txs,
		FraudAssessment: *assessment,
	}, nil
}
