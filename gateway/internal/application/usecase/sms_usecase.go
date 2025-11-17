package usecase

import "context"

type SMSService interface {
	Send(ctx context.Context, phoneNumber string, message string) error
}
