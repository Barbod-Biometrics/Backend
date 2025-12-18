package repository

import (
	"context"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
)

type ServiceRepository interface {
	GetByName(ctx context.Context, serviceName string) (*entity.Service, error)
	GetByID(ctx context.Context, serviceID uint64) (*entity.Service, error)
	Create(ctx context.Context, service *entity.Service) error
	Update(ctx context.Context, service *entity.Service) error
	List(ctx context.Context) ([]*entity.Service, error)
}
