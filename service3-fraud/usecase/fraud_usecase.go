package usecase

import (
	"fmt"
	"strings"
	"time"

	"github.com/oteldemo/service3-fraud/repository"
	"github.com/oteldemo/service3-fraud/types"
)

// FraudUsecase contains the business rules around fraud assessment.
type FraudUsecase interface {
	AssessFraud(params types.AssessFraudParams) (*types.FraudAssessmentResponse, error)
}

// ruleEvaluator decides whether a rule is triggered given the full set
// of transactions being assessed. It returns a human-readable detail
// string explaining why it triggered.
type ruleEvaluator func(all []types.TransactionInput) (triggered bool, detail string)

type fraudUsecase struct {
	rules      repository.RuleRepository
	evaluators map[string]ruleEvaluator
}

func NewFraudUsecase(rules repository.RuleRepository) FraudUsecase {
	u := &fraudUsecase{rules: rules}
	u.evaluators = map[string]ruleEvaluator{
		"large_transaction":      evalLargeTransaction,
		"high_velocity":          evalHighVelocity,
		"international_transfer": evalInternationalTransfer,
		"night_time":             evalNightTime,
	}
	return u
}

func (u *fraudUsecase) AssessFraud(params types.AssessFraudParams) (*types.FraudAssessmentResponse, error) {
	rules := u.rules.ListRules()
	flagged := make([]types.FlaggedRule, 0, len(rules))
	score := 0

	for _, rule := range rules {
		fr := types.FlaggedRule{
			RuleID:      rule.ID,
			Name:        rule.Name,
			Description: rule.Description,
			Severity:    rule.Severity,
			Weight:      rule.Weight,
		}
		if eval, ok := u.evaluators[rule.ID]; ok {
			if triggered, detail := eval(params.Transactions); triggered {
				fr.Triggered = true
				fr.Detail = detail
				score += rule.Weight
			}
		}
		flagged = append(flagged, fr)
	}

	if score > 100 {
		score = 100
	}

	resp := &types.FraudAssessmentResponse{
		UserID:       params.UserID,
		RiskScore:    score,
		RiskLevel:    riskLevel(score),
		FlaggedRules: flagged,
		EvaluatedAt:  time.Now().UTC(),
	}
	resp.Summary = fmt.Sprintf(
		"Assessed %d transactions across %d rules; risk score %d (%s).",
		len(params.Transactions), len(rules), score, resp.RiskLevel,
	)
	return resp, nil
}

func riskLevel(score int) types.RiskLevel {
	switch {
	case score >= 70:
		return types.RiskLevelCritical
	case score >= 40:
		return types.RiskLevelHigh
	case score >= 20:
		return types.RiskLevelMedium
	default:
		return types.RiskLevelLow
	}
}

func evalLargeTransaction(all []types.TransactionInput) (bool, string) {
	for _, tx := range all {
		if tx.Amount > 10000 {
			return true, fmt.Sprintf(
				"transaction #%d to %q for %.2f %s exceeds the 10,000 threshold",
				tx.ID, tx.Merchant, tx.Amount, tx.Currency,
			)
		}
	}
	return false, ""
}

func evalHighVelocity(all []types.TransactionInput) (bool, string) {
	if len(all) < 3 {
		return false, ""
	}
	for i := range all {
		start := all[i].CreatedAt
		count := 0
		for j := range all {
			diff := all[j].CreatedAt.Sub(start)
			if diff >= 0 && diff <= time.Hour {
				count++
			}
		}
		if count >= 3 {
			return true, fmt.Sprintf(
				"%d transactions within a 1-hour window starting %s",
				count, start.UTC().Format(time.RFC3339),
			)
		}
	}
	return false, ""
}

func evalInternationalTransfer(all []types.TransactionInput) (bool, string) {
	for _, tx := range all {
		merchant := strings.ToLower(tx.Merchant)
		desc := strings.ToLower(tx.Description)
		if strings.Contains(merchant, "wire transfer") || strings.Contains(desc, "international") {
			return true, fmt.Sprintf(
				"transaction #%d to %q looks like an international transfer",
				tx.ID, tx.Merchant,
			)
		}
	}
	return false, ""
}

func evalNightTime(all []types.TransactionInput) (bool, string) {
	for _, tx := range all {
		h := tx.CreatedAt.UTC().Hour()
		if h >= 2 && h < 5 {
			return true, fmt.Sprintf(
				"transaction #%d occurred at %s (within 02:00-05:00 UTC window)",
				tx.ID, tx.CreatedAt.UTC().Format("15:04:05"),
			)
		}
	}
	return false, ""
}
