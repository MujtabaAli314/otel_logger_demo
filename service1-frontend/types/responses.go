package types

// DashboardResponse is the aggregated result returned by the frontend
// service. It combines the user profile and transaction history from the
// data service with the fraud risk assessment from the fraud service.
type DashboardResponse struct {
	User            User            `json:"user"`
	Transactions    []Transaction   `json:"transactions"`
	FraudAssessment FraudAssessment `json:"fraud_assessment"`
}

// ErrorResponse is the canonical error body returned by every endpoint.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
