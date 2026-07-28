package controller

import (
	"errors"
	"log/slog"
	"strconv"

	"github.com/kataras/iris/v12"
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
	}
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
	traceIDStr := ctx.GetHeader("traceID")
	spanIDStr := ctx.GetHeader("spanID")

	// Reconstruct the remote parent span context from the headers injected
	// by service1 so this service's span joins the same trace.
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

	spanCtx, span := c.tracer.Start(reqCtx, "ListTransactionsSpan")
	defer span.End()
	span.SetAttributes(
		attribute.Int("controller.user_id", int(id)),
		attribute.String("controller.incoming_trace_id", traceIDStr),
		attribute.String("controller.incoming_span_id", spanIDStr),
	)
	c.logger.InfoContext(spanCtx, "list transactions request received", "user_id", id, "trace_id", traceIDStr, "span_id", spanIDStr)

	limit, _ := strconv.Atoi(ctx.URLParamDefault("limit", "0"))
	offset, _ := strconv.Atoi(ctx.URLParamDefault("offset", "0"))

	txs, err := c.txs.ListTransactions(spanCtx, types.GetTransactionsParams{
		UserID: uint(id),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		span.RecordError(err)
		c.logger.ErrorContext(spanCtx, "list transactions failed", "user_id", id, "err", err.Error())
		writeError(ctx, err)
		return
	}
	span.SetAttributes(attribute.Int("controller.tx_count", len(txs)))
	c.logger.InfoContext(spanCtx, "list transactions completed", "user_id", id, "count", len(txs))
	ctx.JSON(txs)
}

func writeError(ctx iris.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrUserNotFound):
		ctx.StatusCode(iris.StatusNotFound)
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
