package controller

import (
	"errors"
	"log/slog"

	"github.com/kataras/iris/v12"
	"github.com/oteldemo/service1-frontend/repository"
	"github.com/oteldemo/service1-frontend/types"
	"github.com/oteldemo/service1-frontend/usecase"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type Controller struct {
	dashboard usecase.DashboardUsecase
	txs       usecase.TransactionUsecase
	tracer    trace.Tracer
	metric    metric.Meter
	logger    slog.Logger
}

func New(dashboard usecase.DashboardUsecase, txs usecase.TransactionUsecase, tracer trace.Tracer, logger slog.Logger) *Controller {
	return &Controller{dashboard: dashboard,
		txs:    txs,
		tracer: tracer,
		// metric: metricProvider.Meter("SERVICE1-CONTROLLER-Meter"),
		logger: logger}
}

func (c *Controller) Register(app *iris.Application) {
	api := app.Party("/api/v1")
	api.Get("/users/{id:uint}/dashboard", c.GetDashboard)
	api.Post("/users/{id:uint}/transactions", c.CreateTransaction)
}

func (c *Controller) GetDashboard(ctx iris.Context) {
	spanCtx, span := c.tracer.Start(ctx, "GetDashboardSpan")
	defer span.End()
	id, err := ctx.Params().GetUint("id")
	if err != nil {
		ctx.StopWithStatus(iris.StatusInternalServerError)
		c.logger.ErrorContext(spanCtx, "No id provided", "err", err.Error())
		return
	}
	span.SetAttributes(
		attribute.Int("controller.user_id", int(id)),
		attribute.String("controller.route", "/api/v1/users/{id}/dashboard"),
	)
	c.logger.InfoContext(spanCtx, "dashboard request received", "user_id", id)

	resp, err := c.dashboard.GetDashboard(spanCtx, types.GetDashboardParams{UserID: uint(id)})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrUserNotFound):
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(types.ErrorResponse{Error: "not_found", Message: err.Error()})
			c.logger.ErrorContext(spanCtx, "uesr not found", "user_id", id)
		default:
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(types.ErrorResponse{Error: "internal_error", Message: err.Error()})
			c.logger.ErrorContext(spanCtx, "internal server error", "user_id", id, "err", err.Error())
		}
		return
	}
	c.logger.InfoContext(spanCtx, "dashboard request completed", "user_id", id)
	ctx.JSON(resp)
}

func (c *Controller) CreateTransaction(ctx iris.Context) {
	spanCtx, span := c.tracer.Start(ctx, "CreateTransactionSpan")
	defer span.End()
	id, err := ctx.Params().GetUint("id")
	if err != nil {
		ctx.StopWithStatus(iris.StatusInternalServerError)
		c.logger.ErrorContext(spanCtx, "No id provided", "err", err.Error())
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
		c.logger.ErrorContext(spanCtx, "invalid request body", "err", err.Error())
		return
	}
	c.logger.InfoContext(spanCtx, "create transaction request received", "user_id", id, "amount", body.Amount)

	tx, err := c.txs.CreateTransaction(spanCtx, types.CreateTransactionParams{
		UserID:      uint(id),
		Amount:      body.Amount,
		Currency:    body.Currency,
		Type:        body.Type,
		Merchant:    body.Merchant,
		Description: body.Description,
	})
	if err != nil {
		status := iris.StatusInternalServerError
		if sc := repository.StatusCode(err); sc >= 400 && sc < 500 {
			status = sc
		}
		ctx.StatusCode(status)
		_ = ctx.JSON(types.ErrorResponse{Error: "internal_error", Message: err.Error()})
		c.logger.ErrorContext(spanCtx, "create transaction failed", "user_id", id, "err", err.Error())
		return
	}
	span.SetAttributes(attribute.Int("controller.tx_id", int(tx.ID)))
	c.logger.InfoContext(spanCtx, "create transaction completed", "user_id", id, "tx_id", tx.ID)
	ctx.StatusCode(iris.StatusCreated)
	_ = ctx.JSON(tx)
}
