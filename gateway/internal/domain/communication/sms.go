package communication

import "context"

type SMSService interface {
	Send(ctx context.Context, phoneNumber string, code string) error
}
