package usecase

import (
	"context"
	"errors"
	"sync"

	"github.com/oteldemo/logger"
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
	logger logger.Logger
}

func NewDashboardUsecase(data repository.DataClient, fraud repository.FraudClient, logger logger.Logger) DashboardUsecase {
	return &dashboardUsecase{data: data, fraud: fraud, logger: logger}
}

func (u *dashboardUsecase) GetDashboard(ctx context.Context, params types.GetDashboardParams) (*types.DashboardResponse, error) {
	// The span is created by the controller and threaded down via the
	// context. The usecase records to that same span (it does not start
	// its own) and forwards the context to the repository so the repo
	// records to the same span too.
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.Int("usecase.user_id", int(params.UserID)))
	u.logger.LogInfo(ctx, "Heeeey we reached the usecase!!!!!", "user_id", params.UserID)
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
			u.logger.LogError(span, ctx, errors.New("usecase calling repo isnotfound and got an error"), "user_id", params.UserID)
			span.SetAttributes(attribute.KeyValue{Key: "ERROR-01", Value: attribute.StringValue(uErr.Error())})
			return nil, ErrUserNotFound
		}
		// Here is the opposite case where there is no call to the repo so we should register the error here
		u.logger.LogError(span, ctx, errors.New("usecase throwing an error"+uErr.Error()))
		return nil, uErr
	}
	if tErr != nil {
		u.logger.LogError(span, ctx, errors.New("usecase throwing an error"+tErr.Error()))
		return nil, tErr
	}

	span.SetAttributes(attribute.Int("usecase.tx_count", len(txs)))
	u.logger.LogInfo(ctx, "fetched user and transactions", "user_id", params.UserID, "tx_count", len(txs))

	assessment, err := u.fraud.AssessFraud(ctx, params.UserID, txs)
	if err != nil {
		u.logger.LogError(span, ctx, errors.New("usecase throwing an error"+err.Error()))
		return nil, err
	}

	span.SetAttributes(
		attribute.String("usecase.fraud.risk_level", assessment.RiskLevel),
		attribute.Int("usecase.fraud.risk_score", assessment.RiskScore),
	)
	u.logger.LogInfo(ctx, "fraud assessment complete", "risk_level", assessment.RiskLevel, "risk_score", assessment.RiskScore)

	return &types.DashboardResponse{
		User:            *user,
		Transactions:    txs,
		FraudAssessment: *assessment,
	}, nil
}
