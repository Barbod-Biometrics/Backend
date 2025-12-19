package communication

import "context"

type EmailService interface {
	Send(ctx context.Context, to string, subject string, body string) error
	SendWithTemplate(ctx context.Context, to string, subject string, templateName string, data map[string]interface{}) error
}
