package controller

import (
	"errors"
	"strconv"

	"github.com/kataras/iris/v12"
	"github.com/oteldemo/service2-data/types"
	"github.com/oteldemo/service2-data/usecase"
)

type Controller struct {
	users usecase.UserUsecase
	txs   usecase.TransactionUsecase
}

func NewController(users usecase.UserUsecase, txs usecase.TransactionUsecase) *Controller {
	return &Controller{users: users, txs: txs}
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

	limit, _ := strconv.Atoi(ctx.URLParamDefault("limit", "0"))
	offset, _ := strconv.Atoi(ctx.URLParamDefault("offset", "0"))

	txs, err := c.txs.ListTransactions(types.GetTransactionsParams{
		UserID: uint(id),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		writeError(ctx, err)
		return
	}
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
