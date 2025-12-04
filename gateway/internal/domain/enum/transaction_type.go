package enum

type TransactionType string

const (
	TransactionTypeDeposit    TransactionType = "deposit"
	TransactionTypeWithdrawal TransactionType = "withdrawal"
)

func (t TransactionType) IsValid() bool {
	switch t {
	case TransactionTypeDeposit, TransactionTypeWithdrawal:
		return true
	}
	return false
}

func (t TransactionType) String() string {
	return string(t)
}
