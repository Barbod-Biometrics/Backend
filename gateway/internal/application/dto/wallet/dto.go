package wallet

import "time"

// Request DTOs

type DepositRequest struct {
	Amount      uint64 `json:"amount" binding:"required,min=1"`
	Description string `json:"description" binding:"max=500"`
}

// Response DTOs

type WalletSummaryResponse struct {
	Success bool              `json:"success"`
	Data    WalletSummaryData `json:"data"`
}

type WalletSummaryData struct {
	Balance          uint64 `json:"balance"`
	TotalDeposits    int64  `json:"total_deposits"`
	TotalWithdrawals int64  `json:"total_withdrawals"`
	TransactionCount int64  `json:"transaction_count"`
	LastUpdated      string `json:"last_updated"`
}

type TransactionsResponse struct {
	Success bool             `json:"success"`
	Data    TransactionsData `json:"data"`
}

type TransactionsData struct {
	Transactions []TransactionDTO `json:"transactions"`
	TotalCount   int64            `json:"total_count"`
	Page         int              `json:"page"`
	PageSize     int              `json:"page_size"`
	TotalPages   int              `json:"total_pages"`
}

type TransactionDTO struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Amount      int64     `json:"amount"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Date        time.Time `json:"date"`
}

type DepositResponse struct {
	Success bool           `json:"success"`
	Data    DepositDataDTO `json:"data"`
}

type DepositDataDTO struct {
	TransactionID string `json:"transaction_id"`
	NewBalance    uint64 `json:"new_balance"`
	Message       string `json:"message"`
}
