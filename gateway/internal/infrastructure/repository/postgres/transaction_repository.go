package postgres

import (
	"context"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/profile"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/enum"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
	"gorm.io/gorm"
)

type TransactionRepository struct {
	db     *gorm.DB
	logger logger.Logger
}

var _ repository.TransactionRepository = &TransactionRepository{}

func NewTransactionRepository(db *gorm.DB, logger logger.Logger) *TransactionRepository {
	return &TransactionRepository{
		db:     db,
		logger: logger,
	}
}

func (r *TransactionRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value("db_tx").(*gorm.DB); ok {
		return tx
	}
	return r.db
}

func (r *TransactionRepository) Create(ctx context.Context, transaction *entity.Transaction) error {
	db := r.getDB(ctx)
	return db.WithContext(ctx).Create(transaction).Error
}

func (r *TransactionRepository) GetByID(ctx context.Context, transactionID uint64) (*entity.Transaction, error) {
	db := r.getDB(ctx)
	var transaction entity.Transaction

	err := db.WithContext(ctx).First(&transaction, transactionID).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *TransactionRepository) GetByProfileID(ctx context.Context, profileID uint64, pagination repository.TransactionPagination) (*repository.TransactionPaginatedResult, error) {
	db := r.getDB(ctx)
	query := db.WithContext(ctx).Model(&entity.Transaction{}).Where("profile_id = ?", profileID)

	// Count total
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	// Apply pagination
	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.PageSize < 1 {
		pagination.PageSize = 10
	}
	if pagination.PageSize > 100 {
		pagination.PageSize = 100
	}

	offset := (pagination.Page - 1) * pagination.PageSize
	query = query.Order("created_at DESC").Offset(offset).Limit(pagination.PageSize)

	// Execute query
	var transactions []*entity.Transaction
	if err := query.Find(&transactions).Error; err != nil {
		return nil, err
	}

	totalPages := int(totalCount) / pagination.PageSize
	if int(totalCount)%pagination.PageSize > 0 {
		totalPages++
	}

	return &repository.TransactionPaginatedResult{
		Transactions: transactions,
		TotalCount:   totalCount,
		Page:         pagination.Page,
		PageSize:     pagination.PageSize,
		TotalPages:   totalPages,
	}, nil
}

func (r *TransactionRepository) GetWalletSummary(ctx context.Context, profileID uint64) (*repository.WalletSummary, error) {
	db := r.getDB(ctx)

	var totalDeposits int64
	var totalWithdrawals int64
	var transactionCount int64

	// Get total deposits
	err := db.WithContext(ctx).
		Model(&entity.Transaction{}).
		Where("profile_id = ? AND transaction_type = ?", profileID, enum.TransactionTypeDeposit).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalDeposits).Error
	if err != nil {
		return nil, err
	}

	// Get total withdrawals (stored as negative, so we take absolute value)
	err = db.WithContext(ctx).
		Model(&entity.Transaction{}).
		Where("profile_id = ? AND transaction_type = ?", profileID, enum.TransactionTypeWithdrawal).
		Select("COALESCE(ABS(SUM(amount)), 0)").
		Scan(&totalWithdrawals).Error
	if err != nil {
		return nil, err
	}

	// Get transaction count
	err = db.WithContext(ctx).
		Model(&entity.Transaction{}).
		Where("profile_id = ?", profileID).
		Count(&transactionCount).Error
	if err != nil {
		return nil, err
	}

	return &repository.WalletSummary{
		TotalDeposits:    totalDeposits,
		TotalWithdrawals: totalWithdrawals,
		TransactionCount: transactionCount,
	}, nil
}

func (r *TransactionRepository) GetUsageSummary(ctx context.Context, profileID uint64) (*profile.UsageSummaryResponse, error) {
	db := r.getDB(ctx)

	r.logger.Info("starting usage summary calculation",
		logger.Field{Key: "profile_id", Value: profileID},
	)

	response := &profile.UsageSummaryResponse{
		TotalSpend:       0,
		ServiceBreakdown: []profile.ServiceUsageStats{},
	}

	// Query 1: The Grand Total (All time spend)
	err := db.WithContext(ctx).
		Model(&entity.Transaction{}).
		Where("profile_id = ? AND transaction_type = ?", profileID, "withdrawal").
		Select("COALESCE(ABS(SUM(amount)), 0)").
		Scan(&response.TotalSpend).Error

	if err != nil {
		r.logger.Error("failed to calculate total spend",
			logger.Field{Key: "profile_id", Value: profileID},
			logger.Field{Key: "error", Value: err},
		)
		return nil, err
	}

	// Query 2: The Breakdown (Grouped by Service)
	err = db.WithContext(ctx).
		Table("transactions").
		Select("services.service_name, count(transactions.transaction_id) as total_count, ABS(SUM(transactions.amount)) as total_cost").
		Joins("JOIN verification_jobs ON verification_jobs.job_id = transactions.job_id").
		Joins("JOIN services ON services.service_id = verification_jobs.service_id").
		Where("transactions.profile_id = ? AND transactions.transaction_type = ?", profileID, "withdrawal").
		Group("services.service_name").
		Scan(&response.ServiceBreakdown).Error

	if err != nil {
		r.logger.Error("failed to fetch service breakdown",
			logger.Field{Key: "profile_id", Value: profileID},
			logger.Field{Key: "error", Value: err},
		)
		return nil, err
	}

	r.logger.Debug("usage summary calculated successfully",
		logger.Field{Key: "profile_id", Value: profileID},
		logger.Field{Key: "total_spend", Value: response.TotalSpend},
		logger.Field{Key: "service_count", Value: len(response.ServiceBreakdown)},
	)

	return response, nil

}
