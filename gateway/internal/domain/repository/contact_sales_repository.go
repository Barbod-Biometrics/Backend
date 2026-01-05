package repository

import (
	"context"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
)

type ContactSalesFilter struct {
	Status   *entity.ContactSalesStatus
	Search   *string
	HasEmail *bool
	FromDate *time.Time
	ToDate   *time.Time
}

type ContactSalesSort struct {
	Field string // created_at, status
	Order string // asc, desc
}

type ContactSalesPaginatedResult struct {
	Items      []*entity.ContactSalesRequest
	TotalCount int64
	Page       int
	PageSize   int
	TotalPages int
}

type ContactSalesRepository interface {
	Create(ctx context.Context, req *entity.ContactSalesRequest) error
	List(ctx context.Context, filter ContactSalesFilter, sort ContactSalesSort, pagination Pagination) (*ContactSalesPaginatedResult, error)
	MarkRead(ctx context.Context, id uint64) error
	Delete(ctx context.Context, id uint64) error
}
