package postgres

import (
	"context"
	"fmt"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
	"gorm.io/gorm"
)

type ServiceRepository struct {
	db *gorm.DB
}

func NewServiceRepository(db *gorm.DB) repository.ServiceRepository {
	return &ServiceRepository{db: db}
}

func (r *ServiceRepository) GetByName(ctx context.Context, serviceName string) (*entity.Service, error) {
	var service entity.Service
	if err := r.db.WithContext(ctx).Where("service_name = ?", serviceName).First(&service).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("service not found: %s", serviceName)
		}
		return nil, fmt.Errorf("failed to get service by name: %w", err)
	}
	return &service, nil
}

func (r *ServiceRepository) GetByID(ctx context.Context, serviceID uint64) (*entity.Service, error) {
	var service entity.Service
	if err := r.db.WithContext(ctx).First(&service, serviceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("service not found: %d", serviceID)
		}
		return nil, fmt.Errorf("failed to get service by ID: %w", err)
	}
	return &service, nil
}

func (r *ServiceRepository) Create(ctx context.Context, service *entity.Service) error {
	if err := r.db.WithContext(ctx).Create(service).Error; err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	return nil
}

func (r *ServiceRepository) Update(ctx context.Context, service *entity.Service) error {
	if err := r.db.WithContext(ctx).Save(service).Error; err != nil {
		return fmt.Errorf("failed to update service: %w", err)
	}
	return nil
}

func (r *ServiceRepository) List(ctx context.Context) ([]*entity.Service, error) {
	var services []*entity.Service
	if err := r.db.WithContext(ctx).Find(&services).Error; err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}
	return services, nil
}
