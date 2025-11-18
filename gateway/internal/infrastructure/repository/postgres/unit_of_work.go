package postgres

import (
	"context"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
	"gorm.io/gorm"
)

type gormUnitOfWork struct {
	db *gorm.DB
}

type ctxKey string

const dbTxKey ctxKey = "db_tx"

func NewGormUnitOfWork(db *gorm.DB) repository.UnitOfWork {
	return &gormUnitOfWork{db: db}
}

func (u *gormUnitOfWork) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, dbTxKey, tx)

		return fn(txCtx)
	})
}
