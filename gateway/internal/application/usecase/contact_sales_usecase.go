package usecase

import (
	"context"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/contact"
)

type ContactSalesUsecase interface {
	Submit(ctx context.Context, req contact.SubmitContactSalesRequest, remoteIP string) (*contact.SubmitContactSalesResponse, error)
	AdminList(ctx context.Context, req contact.AdminListContactSalesRequest) (*contact.ContactSalesListResponse, error)
	MarkRead(ctx context.Context, id uint64) (*contact.ContactSalesActionResponse, error)
	Delete(ctx context.Context, id uint64) (*contact.ContactSalesActionResponse, error)
}
