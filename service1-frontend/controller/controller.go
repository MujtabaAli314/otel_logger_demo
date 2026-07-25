package controller

import (
	"errors"

	"log/slog"

	"github.com/kataras/iris/v12"
	"github.com/oteldemo/service1-frontend/types"
	"github.com/oteldemo/service1-frontend/usecase"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type Controller struct {
	dashboard usecase.DashboardUsecase
	tracer    trace.Tracer
	metric    metric.Meter
	logger    slog.Logger
}

func New(dashboard usecase.DashboardUsecase, tracer trace.Tracer, logger slog.Logger) *Controller {
	return &Controller{dashboard: dashboard,
		tracer: tracer,
		// metric: metricProvider.Meter("SERVICE1-CONTROLLER-Meter"),
		logger: logger}
}

func (c *Controller) Register(app *iris.Application) {
	api := app.Party("/api/v1")
	api.Get("/users/{id:uint}/dashboard", c.GetDashboard)
}

func (c *Controller) GetDashboard(ctx iris.Context) {
	spanCtx, span := c.tracer.Start(ctx, "GetDashboardSpan")
	defer span.End()
	id, err := ctx.Params().GetUint("id")
	if err != nil {
		ctx.StopWithStatus(iris.StatusInternalServerError)
		c.logger.ErrorContext(ctx, "No id provided")
		return
	}
	resp, err := c.dashboard.GetDashboard(spanCtx, types.GetDashboardParams{UserID: uint(id)})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrUserNotFound):
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(types.ErrorResponse{Error: "not_found", Message: err.Error()})
			c.logger.ErrorContext(spanCtx, "uesr not found")
		default:
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(types.ErrorResponse{Error: "internal_error", Message: err.Error()})
			c.logger.ErrorContext(spanCtx, "internal server error")
		}
		return
	}
	ctx.JSON(resp)
}
