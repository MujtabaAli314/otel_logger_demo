package types

// AssessFraudParams is the input to a fraud assessment. The fraud
// service is stateless, so callers must supply the transactions to
// evaluate.
type AssessFraudParams struct {
	UserID       uint
	Transactions []TransactionInput
}
