package postgres

import (
	"context"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
	"gorm.io/gorm"
)

type GormUnitOfWork struct {
	db *gorm.DB
}

func NewGormUnitOfWork(db *gorm.DB) repository.UnitOfWork {
	return &GormUnitOfWork{db: db}
}

func (u *GormUnitOfWork) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, "db_tx", tx)

		return fn(txCtx)
	})
}
