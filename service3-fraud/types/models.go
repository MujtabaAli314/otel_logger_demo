package types

import "time"

// Severity classifies how strongly a rule indicates fraud.
type Severity string

const (
	SeverityLow    Severity = "low"
	SeverityMedium Severity = "medium"
	SeverityHigh   Severity = "high"
)

// RiskLevel classifies the overall assessment outcome.
type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)

// TransactionInput is the inbound representation of a transaction as
// received by the fraud service. It is intentionally independent of any
// other service's internal model so this service owns its own contract.
type TransactionInput struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Type        string    `json:"type"`
	Merchant    string    `json:"merchant"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// Rule describes a single fraud-detection heuristic: its metadata, the
// severity it implies, and the weight (points) it contributes to the
// overall risk score when triggered.
type Rule struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Severity    Severity `json:"severity"`
	Weight      int      `json:"weight"`
}
