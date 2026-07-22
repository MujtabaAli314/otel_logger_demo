package types

import "time"

// User is the persistent representation of an application user.
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	Email     string    `gorm:"size:255;uniqueIndex;not null" json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (User) TableName() string { return "users" }

// TransactionType enumerates the direction of money movement.
type TransactionType string

const (
	TransactionTypeDebit  TransactionType = "debit"
	TransactionTypeCredit TransactionType = "credit"
)

// Transaction is the persistent representation of a single financial
// movement belonging to a user.
type Transaction struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	UserID      uint            `gorm:"index;not null" json:"user_id"`
	Amount      float64         `gorm:"not null" json:"amount"`
	Currency    string          `gorm:"size:3;not null" json:"currency"`
	Type        TransactionType `gorm:"type:varchar(10);not null" json:"type"`
	Merchant    string          `gorm:"size:255" json:"merchant"`
	Description string          `gorm:"size:512" json:"description"`
	CreatedAt   time.Time       `json:"created_at"`
}

func (Transaction) TableName() string { return "transactions" }
