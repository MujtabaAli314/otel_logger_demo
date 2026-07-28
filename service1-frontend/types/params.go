package types

// GetDashboardParams is the input required to build a user dashboard.
type GetDashboardParams struct {
	UserID uint
}

// CreateTransactionParams is the input required to create a transaction
// via the data service.
type CreateTransactionParams struct {
	UserID      uint
	Amount      float64
	Currency    string
	Type        string
	Merchant    string
	Description string
}
