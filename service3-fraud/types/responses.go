package types

import "time"

// FlaggedRule is the per-rule outcome of an assessment, whether or not
// the rule was triggered.
type FlaggedRule struct {
	RuleID      string   `json:"rule_id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Severity    Severity `json:"severity"`
	Weight      int      `json:"weight"`
	Triggered   bool     `json:"triggered"`
	Detail      string   `json:"detail,omitempty"`
}

// FraudAssessmentResponse is the full result of evaluating a set of
// transactions against the fraud rule catalog.
type FraudAssessmentResponse struct {
	UserID       uint          `json:"user_id"`
	RiskScore    int           `json:"risk_score"`
	RiskLevel    RiskLevel     `json:"risk_level"`
	FlaggedRules []FlaggedRule `json:"flagged_rules"`
	Summary      string        `json:"summary"`
	EvaluatedAt  time.Time     `json:"evaluated_at"`
}

// ErrorResponse is the canonical error body returned by every endpoint.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
