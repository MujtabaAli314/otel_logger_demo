package controller

import (
	"errors"

	"github.com/kataras/iris/v12"
	"github.com/oteldemo/logger"
	"github.com/oteldemo/service1-frontend/repository"
	"github.com/oteldemo/service1-frontend/types"
	"github.com/oteldemo/service1-frontend/usecase"
	"go.opentelemetry.io/otel/attribute"
)

type Controller struct {
	dashboard usecase.DashboardUsecase
	txs       usecase.TransactionUsecase
	logger    logger.Logger
}

func New(dashboard usecase.DashboardUsecase, txs usecase.TransactionUsecase, logger logger.Logger) *Controller {
	return &Controller{dashboard: dashboard,
		txs:    txs,
		logger: logger}
}

func (c *Controller) Register(app *iris.Application) {
	api := app.Party("/api/v1")
	api.Get("/users/{id:uint}/dashboard", c.GetDashboard)
	api.Post("/users/{id:uint}/transactions", c.CreateTransaction)
}

func (c *Controller) GetDashboard(ctx iris.Context) {
	spanCtx, span := c.logger.StartSpan(ctx, "GetDashboardSpan")
	defer span.End()
	id, err := ctx.Params().GetUint("id")
	if err != nil {
		ctx.StopWithStatus(iris.StatusInternalServerError)
		c.logger.LogError(span, spanCtx, errors.New("No id provided"), "err", err.Error())
		return
	}
	span.SetAttributes(
		attribute.Int("controller.user_id", int(id)),
		attribute.String("controller.route", "/api/v1/users/{id}/dashboard"),
	)
	c.logger.LogInfo(spanCtx, "dashboard request received", "user_id", id)

	resp, err := c.dashboard.GetDashboard(spanCtx, types.GetDashboardParams{UserID: uint(id)})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrUserNotFound):
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(types.ErrorResponse{Error: "not_found", Message: err.Error()})
			c.logger.LogError(span, spanCtx, errors.New("uesr not found"), "user_id", id)
		default:
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(types.ErrorResponse{Error: "internal_error", Message: err.Error()})
			c.logger.LogError(span, spanCtx, errors.New("internal server error"), "user_id", id, "err", err.Error())
		}
		return
	}
	c.logger.LogInfo(spanCtx, "dashboard request completed", "user_id", id)
	ctx.JSON(resp)
}

func (c *Controller) CreateTransaction(ctx iris.Context) {
	spanCtx, span := c.logger.StartSpan(ctx, "CreateTransactionSpan")
	defer span.End()
	id, err := ctx.Params().GetUint("id")
	if err != nil {
		ctx.StopWithStatus(iris.StatusInternalServerError)
		c.logger.LogError(span, spanCtx, errors.New("No id provided"), "err", err.Error())
		return
	}
	span.SetAttributes(
		attribute.Int("controller.user_id", int(id)),
		attribute.String("controller.route", "/api/v1/users/{id}/transactions"),
	)

	var body struct {
		Amount      float64 `json:"amount"`
		Currency    string  `json:"currency"`
		Type        string  `json:"type"`
		Merchant    string  `json:"merchant"`
		Description string  `json:"description"`
	}
	if err := ctx.ReadJSON(&body); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(types.ErrorResponse{Error: "bad_request", Message: err.Error()})
		c.logger.LogError(span, spanCtx, errors.New("invalid request body"), "err", err.Error())
		return
	}
	c.logger.LogInfo(spanCtx, "create transaction request received", "user_id", id, "amount", body.Amount)

	// tx, cerr := c.logger.ExeAndLog(span, spanCtx, c.txs.CreateTransaction, uint(id), body.Amount, body.Currency, body.Type, body.Merchant, body.Description)
	tx, cerr := logger.ExeAngLog(c.logger, span, spanCtx, c.txs.CreateTransaction, types.CreateTransactionParams{
		UserID:      uint(id),
		Amount:      body.Amount,
		Currency:    body.Currency,
		Type:        body.Type,
		Merchant:    body.Merchant,
		Description: body.Description,
	})
	// tx, err := c.txs.CreateTransaction(spanCtx, types.CreateTransactionParams{
	// 	UserID:      uint(id),
	// 	Amount:      body.Amount,
	// 	Currency:    body.Currency,
	// 	Type:        body.Type,
	// 	Merchant:    body.Merchant,
	// 	Description: body.Description,
	// })
	if cerr != nil {
		status := iris.StatusInternalServerError
		if sc := repository.StatusCode(err); sc >= 400 && sc < 500 {
			status = sc
		}
		ctx.StatusCode(status)
		_ = ctx.JSON(types.ErrorResponse{Error: "internal_error", Message: cerr.Error()})
		c.logger.LogError(span, spanCtx, errors.New("create transaction failed"), "user_id", id, "err", err.Error())
		return
	}
	span.SetAttributes(attribute.Int("controller.tx_id", int(tx.ID)))
	c.logger.LogInfo(spanCtx, "create transaction completed", "user_id", id, "tx_id", tx.ID)
	ctx.StatusCode(iris.StatusCreated)
	_ = ctx.JSON(tx)
}
