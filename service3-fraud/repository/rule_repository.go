package repository

import "github.com/oteldemo/service3-fraud/types"

// RuleRepository abstracts access to the fraud rule catalog. The default
// implementation is in-memory, but it could later be backed by a
// database or a remote configuration service without changing the
// usecase.
type RuleRepository interface {
	ListRules() []types.Rule
}

type ruleRepository struct {
	rules []types.Rule
}

func NewRuleRepository() RuleRepository {
	return &ruleRepository{
		rules: []types.Rule{
			{
				ID:          "large_transaction",
				Name:        "Large Transaction",
				Description: "Flags any single transaction above 10,000",
				Severity:    types.SeverityHigh,
				Weight:      40,
			},
			{
				ID:          "high_velocity",
				Name:        "High Velocity",
				Description: "Flags when 3 or more transactions occur within a 1-hour window",
				Severity:    types.SeverityMedium,
				Weight:      25,
			},
			{
				ID:          "international_transfer",
				Name:        "International Transfer",
				Description: "Flags transfers to known international wire merchants",
				Severity:    types.SeverityMedium,
				Weight:      20,
			},
			{
				ID:          "night_time",
				Name:        "Night-Time Activity",
				Description: "Flags transactions made between 02:00 and 05:00 UTC",
				Severity:    types.SeverityLow,
				Weight:      15,
			},
		},
	}
}

func (r *ruleRepository) ListRules() []types.Rule {
	out := make([]types.Rule, len(r.rules))
	copy(out, r.rules)
	return out
}
