package controller

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/kataras/iris/v12"
	"github.com/oteldemo/service2-data/tracing"
	"github.com/oteldemo/service2-data/types"
	"github.com/oteldemo/service2-data/usecase"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Controller struct {
	users  usecase.UserUsecase
	txs    usecase.TransactionUsecase
	tracer trace.Tracer
	logger slog.Logger
}

func NewController(users usecase.UserUsecase, txs usecase.TransactionUsecase, tracer trace.Tracer, logger slog.Logger) *Controller {
	return &Controller{users: users, txs: txs, tracer: tracer, logger: logger}
}

func (c *Controller) Register(app *iris.Application) {
	api := app.Party("/api/v1")
	{
		api.Get("/users/{id:uint}", c.GetUser)
		api.Get("/users/{id:uint}/transactions", c.ListTransactions)
		api.Post("/users/{id:uint}/transactions", c.CreateTransaction)
	}
}

// startSpanFromHeaders reconstructs a remote parent span context from the
// traceID/spanID headers injected by service1 and starts a new child span.
// If the headers are absent or invalid a root span is created. The
// incoming IDs are recorded as attributes for debugging the propagation.
func (c *Controller) startSpanFromHeaders(ctx iris.Context, name string) (context.Context, trace.Span) {
	traceIDStr := ctx.GetHeader("traceID")
	spanIDStr := ctx.GetHeader("spanID")
	reqCtx := ctx.Request().Context()
	if traceIDStr != "" && spanIDStr != "" {
		if tid, err := trace.TraceIDFromHex(traceIDStr); err == nil {
			if sid, err := trace.SpanIDFromHex(spanIDStr); err == nil {
				remote := trace.NewSpanContext(trace.SpanContextConfig{
					TraceID:    tid,
					SpanID:     sid,
					Remote:     true,
					TraceFlags: trace.FlagsSampled,
				})
				reqCtx = trace.ContextWithSpanContext(reqCtx, remote)
			}
		}
	}
	spanCtx, span := c.tracer.Start(reqCtx, name)
	span.SetAttributes(
		attribute.String("controller.incoming_trace_id", traceIDStr),
		attribute.String("controller.incoming_span_id", spanIDStr),
	)
	return spanCtx, span
}

func (c *Controller) GetUser(ctx iris.Context) {
	id, _ := ctx.Params().GetUint("id")

	user, err := c.users.GetUser(types.GetUserParams{ID: uint(id)})
	if err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(user)
}

func (c *Controller) ListTransactions(ctx iris.Context) {
	id, _ := ctx.Params().GetUint("id")
	spanCtx, span := c.startSpanFromHeaders(ctx, "ListTransactionsSpan")
	defer span.End()
	span.SetAttributes(attribute.Int("controller.user_id", int(id)))
	// Attach the parent (service1) span ID to every log record emitted in
	// this request so service2 logs can be correlated back to the caller.
	log := c.logger.With("parent_span_id", tracing.ParentSpanID(spanCtx))
	log.InfoContext(spanCtx, "list transactions request received", "user_id", id)

	limit, _ := strconv.Atoi(ctx.URLParamDefault("limit", "0"))
	offset, _ := strconv.Atoi(ctx.URLParamDefault("offset", "0"))

	txs, err := c.txs.ListTransactions(spanCtx, types.GetTransactionsParams{
		UserID: uint(id),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		span.RecordError(err)
		log.ErrorContext(spanCtx, "list transactions failed", "user_id", id, "err", err.Error())
		writeError(ctx, err)
		return
	}
	span.SetAttributes(attribute.Int("controller.tx_count", len(txs)))
	log.InfoContext(spanCtx, "list transactions completed", "user_id", id, "count", len(txs))
	ctx.JSON(txs)
}

func (c *Controller) CreateTransaction(ctx iris.Context) {
	id, _ := ctx.Params().GetUint("id")
	spanCtx, span := c.startSpanFromHeaders(ctx, "CreateTransactionSpan")
	defer span.End()
	span.SetAttributes(attribute.Int("controller.user_id", int(id)))
	log := c.logger.With("parent_span_id", tracing.ParentSpanID(spanCtx))

	var body struct {
		Amount      float64 `json:"amount"`
		Currency    string  `json:"currency"`
		Type        string  `json:"type"`
		Merchant    string  `json:"merchant"`
		Description string  `json:"description"`
	}
	if err := ctx.ReadJSON(&body); err != nil {
		span.RecordError(err)
		log.ErrorContext(spanCtx, "invalid request body", "err", err.Error())
		writeError(ctx, fmt.Errorf("%w: %s", usecase.ErrInvalidTransaction, err.Error()))
		return
	}
	log.InfoContext(spanCtx, "create transaction request received", "user_id", id, "amount", body.Amount)

	resp, err := c.txs.CreateTransaction(spanCtx, types.CreateTransactionParams{
		UserID:      uint(id),
		Amount:      body.Amount,
		Currency:    body.Currency,
		Type:        types.TransactionType(body.Type),
		Merchant:    body.Merchant,
		Description: body.Description,
	})
	if err != nil {
		span.RecordError(err)
		log.ErrorContext(spanCtx, "create transaction failed", "user_id", id, "err", err.Error())
		writeError(ctx, err)
		return
	}
	span.SetAttributes(attribute.Int("controller.tx_id", int(resp.ID)))
	log.InfoContext(spanCtx, "create transaction completed", "user_id", id, "tx_id", resp.ID)
	ctx.StatusCode(iris.StatusCreated)
	_ = ctx.JSON(resp)
}

func writeError(ctx iris.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrUserNotFound):
		ctx.StatusCode(iris.StatusNotFound)
	case errors.Is(err, usecase.ErrInvalidTransaction):
		ctx.StatusCode(iris.StatusBadRequest)
	default:
		ctx.StatusCode(iris.StatusInternalServerError)
	}
	_ = ctx.JSON(types.ErrorResponse{
		Error:   httpStatusReason(ctx.GetStatusCode()),
		Message: err.Error(),
	})
}

func httpStatusReason(code int) string {
	switch code {
	case iris.StatusNotFound:
		return "not_found"
	case iris.StatusBadRequest:
		return "bad_request"
	default:
		return "internal_error"
	}
}
