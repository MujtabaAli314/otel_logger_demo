package controller

import (
	"errors"

	"github.com/kataras/iris/v12"
	"github.com/oteldemo/service1-frontend/types"
	"github.com/oteldemo/service1-frontend/usecase"
)

type Controller struct {
	dashboard usecase.DashboardUsecase
}

func New(dashboard usecase.DashboardUsecase) *Controller {
	return &Controller{dashboard: dashboard}
}

func (c *Controller) Register(app *iris.Application) {
	api := app.Party("/api/v1")
	api.Get("/users/{id:uint}/dashboard", c.GetDashboard)
}

func (c *Controller) GetDashboard(ctx iris.Context) {
	id, _ := ctx.Params().GetUint("id")

	resp, err := c.dashboard.GetDashboard(ctx.Request().Context(), types.GetDashboardParams{UserID: uint(id)})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrUserNotFound):
			ctx.StatusCode(iris.StatusNotFound)
			_ = ctx.JSON(types.ErrorResponse{Error: "not_found", Message: err.Error()})
		default:
			ctx.StatusCode(iris.StatusInternalServerError)
			_ = ctx.JSON(types.ErrorResponse{Error: "internal_error", Message: err.Error()})
		}
		return
	}
	ctx.JSON(resp)
}
