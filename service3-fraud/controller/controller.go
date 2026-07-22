package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/oteldemo/service3-fraud/types"
	"github.com/oteldemo/service3-fraud/usecase"
)

type Controller struct {
	fraud usecase.FraudUsecase
}

func New(fraud usecase.FraudUsecase) *Controller {
	return &Controller{fraud: fraud}
}

func (c *Controller) Register(app *iris.Application) {
	api := app.Party("/api/v1")
	api.Post("/fraud/assess", c.AssessFraud)
}

// assessRequest is the JSON body expected by POST /fraud/assess.
type assessRequest struct {
	UserID       uint                      `json:"user_id"`
	Transactions []types.TransactionInput  `json:"transactions"`
}

func (c *Controller) AssessFraud(ctx iris.Context) {
	var req assessRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		_ = ctx.JSON(types.ErrorResponse{Error: "bad_request", Message: err.Error()})
		return
	}

	resp, err := c.fraud.AssessFraud(types.AssessFraudParams{
		UserID:       req.UserID,
		Transactions: req.Transactions,
	})
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		_ = ctx.JSON(types.ErrorResponse{Error: "internal_error", Message: err.Error()})
		return
	}
	ctx.JSON(resp)
}
