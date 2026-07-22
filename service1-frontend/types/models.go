package types

import "time"

// User is the frontend's view of a user, decoded from the data
// service's JSON contract. Each service owns its own DTOs.
type User struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// Transaction is the frontend's view of a transaction, decoded from the
// data service. It is also reused as the outbound payload to the fraud
// service since both contracts share the same shape.
type Transaction struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Type        string    `json:"type"`
	Merchant    string    `json:"merchant"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// FlaggedRule is the frontend's view of a fraud rule outcome.
type FlaggedRule struct {
	RuleID      string `json:"rule_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Weight      int    `json:"weight"`
	Triggered   bool   `json:"triggered"`
	Detail      string `json:"detail,omitempty"`
}

// FraudAssessment is the frontend's view of a fraud risk assessment,
// decoded from the fraud service.
type FraudAssessment struct {
	UserID       uint          `json:"user_id"`
	RiskScore    int           `json:"risk_score"`
	RiskLevel    string        `json:"risk_level"`
	FlaggedRules []FlaggedRule `json:"flagged_rules"`
	Summary      string        `json:"summary"`
	EvaluatedAt  time.Time     `json:"evaluated_at"`
}
