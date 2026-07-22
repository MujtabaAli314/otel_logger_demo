package types

import "time"

// UserResponse is the transport representation of a user returned by the
// data service. It deliberately omits internal/audit fields such as
// UpdatedAt so the public contract stays small.
type UserResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func NewUserResponse(u User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	}
}

// TransactionResponse is the transport representation of a transaction.
type TransactionResponse struct {
	ID          uint            `json:"id"`
	UserID      uint            `json:"user_id"`
	Amount      float64         `json:"amount"`
	Currency    string          `json:"currency"`
	Type        TransactionType `json:"type"`
	Merchant    string          `json:"merchant"`
	Description string          `json:"description"`
	CreatedAt   time.Time       `json:"created_at"`
}

func NewTransactionResponse(t Transaction) TransactionResponse {
	return TransactionResponse{
		ID:          t.ID,
		UserID:      t.UserID,
		Amount:      t.Amount,
		Currency:    t.Currency,
		Type:        t.Type,
		Merchant:    t.Merchant,
		Description: t.Description,
		CreatedAt:   t.CreatedAt,
	}
}

// ErrorResponse is the canonical error body returned by every endpoint.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
