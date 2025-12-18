package enum

type TransactionType string

const (
	TransactionTypeDeposit          TransactionType = "deposit"
	TransactionTypeWithdrawal       TransactionType = "withdrawal"
	TransactionTypeFaceVerification TransactionType = "face_verification"
)

func (t TransactionType) IsValid() bool {
	switch t {
	case TransactionTypeDeposit, TransactionTypeWithdrawal, TransactionTypeFaceVerification:
		return true
	}
	return false
}

func (t TransactionType) String() string {
	return string(t)
}
