package types

// GetUserParams is the input required to fetch a single user.
type GetUserParams struct {
	ID uint
}

// GetTransactionsParams is the input required to list a user's
// transactions. Limit and Offset are optional and defaulted by the
// usecase when left zero.
type GetTransactionsParams struct {
	UserID uint
	Limit  int
	Offset int
}
