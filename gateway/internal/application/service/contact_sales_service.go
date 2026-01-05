package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/contact"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
)

type CaptchaVerifier interface {
	Verify(ctx context.Context, token string, remoteIP string) error
	IsEnabled() bool
}

type ContactSalesService struct {
	repo    repository.ContactSalesRepository
	captcha CaptchaVerifier
	logger  logger.Logger
}

func NewContactSalesService(repo repository.ContactSalesRepository, captcha CaptchaVerifier, l logger.Logger) usecase.ContactSalesUsecase {
	return &ContactSalesService{
		repo:    repo,
		captcha: captcha,
		logger:  l,
	}
}

func (s *ContactSalesService) Submit(ctx context.Context, req contact.SubmitContactSalesRequest, remoteIP string) (*contact.SubmitContactSalesResponse, error) {
	if s.captcha != nil && s.captcha.IsEnabled() {
		if req.CaptchaToken == "" {
			return nil, errors.New("captcha token is required")
		}
		if err := s.captcha.Verify(ctx, req.CaptchaToken, remoteIP); err != nil {
			return nil, fmt.Errorf("captcha verification failed: %w", err)
		}
	}

	entityReq := &entity.ContactSalesRequest{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Phone:        req.Phone,
		BusinessName: req.BusinessName,
		Description:  req.Description,
		Status:       entity.ContactSalesStatusNew,
	}

	if req.Email != "" {
		entityReq.Email = &req.Email
	}

	if err := s.repo.Create(ctx, entityReq); err != nil {
		return nil, fmt.Errorf("failed to save contact sales request: %w", err)
	}

	return &contact.SubmitContactSalesResponse{
		ID:      entityReq.ID,
		Message: "submitted",
	}, nil
}

func (s *ContactSalesService) AdminList(ctx context.Context, req contact.AdminListContactSalesRequest) (*contact.ContactSalesListResponse, error) {
	filter := repository.ContactSalesFilter{}
	if req.Status != "" {
		status := entity.ContactSalesStatus(req.Status)
		filter.Status = &status
	}
	if req.Search != "" {
		filter.Search = &req.Search
	}
	if req.HasEmail != nil {
		filter.HasEmail = req.HasEmail
	}
	if req.FromDate != "" {
		from, err := time.Parse("2006-01-02", req.FromDate)
		if err != nil {
			return nil, fmt.Errorf("invalid from_date: %w", err)
		}
		filter.FromDate = &from
	}
	if req.ToDate != "" {
		to, err := time.Parse("2006-01-02", req.ToDate)
		if err != nil {
			return nil, fmt.Errorf("invalid to_date: %w", err)
		}
		filter.ToDate = &to
	}

	sort := repository.ContactSalesSort{
		Field: req.SortBy,
		Order: req.SortOrder,
	}

	pagination := repository.Pagination{
		Page:     req.Page,
		PageSize: req.PageSize,
	}
	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.PageSize < 1 {
		pagination.PageSize = 10
	}

	res, err := s.repo.List(ctx, filter, sort, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to list contact sales requests: %w", err)
	}

	items := make([]contact.ContactSalesItem, 0, len(res.Items))
	for _, item := range res.Items {
		it := contact.ContactSalesItem{
			ID:           item.ID,
			FirstName:    item.FirstName,
			LastName:     item.LastName,
			Phone:        item.Phone,
			BusinessName: item.BusinessName,
			Description:  item.Description,
			Status:       string(item.Status),
			ReadAt:       item.ReadAt,
			CreatedAt:    item.CreatedAt,
		}
		if item.Email != nil {
			it.Email = *item.Email
		}
		items = append(items, it)
	}

	return &contact.ContactSalesListResponse{
		Items:      items,
		TotalCount: res.TotalCount,
		Page:       res.Page,
		PageSize:   res.PageSize,
		TotalPages: res.TotalPages,
	}, nil
}

func (s *ContactSalesService) MarkRead(ctx context.Context, id uint64) (*contact.ContactSalesActionResponse, error) {
	if id == 0 {
		return nil, errors.New("invalid id")
	}
	if err := s.repo.MarkRead(ctx, id); err != nil {
		return nil, fmt.Errorf("failed to mark as read: %w", err)
	}
	return &contact.ContactSalesActionResponse{Success: true, Message: "marked as read"}, nil
}

func (s *ContactSalesService) Delete(ctx context.Context, id uint64) (*contact.ContactSalesActionResponse, error) {
	if id == 0 {
		return nil, errors.New("invalid id")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return nil, fmt.Errorf("failed to delete: %w", err)
	}
	return &contact.ContactSalesActionResponse{Success: true, Message: "deleted"}, nil
}
