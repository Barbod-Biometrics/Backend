package email

import (
	"bytes"
	"context"
	"html/template"
	"path/filepath"

	"github.com/Barbod-Biometrics/Backend/gateway/bootstrap"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/communication"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	"gopkg.in/gomail.v2"
)

type GmailService struct {
	gmailConfig bootstrap.Gmail
	appLogger   logger.Logger
}

func NewGmailService(gmailConfig bootstrap.Gmail, appLogger logger.Logger) *GmailService {
	return &GmailService{
		gmailConfig: gmailConfig,
		appLogger:   appLogger,
	}
}

var _ communication.EmailService = (*GmailService)(nil)

func (g *GmailService) Send(ctx context.Context, to string, subject string, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", g.gmailConfig.SenderEmail)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	s := gomail.NewDialer(g.gmailConfig.SmtpHost, g.gmailConfig.SmtpPort, g.gmailConfig.SenderEmail, g.gmailConfig.GmailAppPassword)

	if err := s.DialAndSend(m); err != nil {
		g.appLogger.ErrorContext(ctx, "Failed to send email", logger.Field{Key: "error", Value: err})

		return err
	}

	return nil
}

func (g *GmailService) SendWithTemplate(ctx context.Context, to string, subject string, templateName string, data map[string]interface{}) error {
	templatePath := filepath.Join(g.gmailConfig.TemplateDir, templateName)

	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		g.appLogger.ErrorContext(ctx, "Failed to parse email template", logger.Field{Key: "error", Value: err})
		return err
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		g.appLogger.ErrorContext(ctx, "Failed to execute email template", logger.Field{Key: "error", Value: err})
		return err
	}

	m := gomail.NewMessage()
	m.SetHeader("From", g.gmailConfig.SenderEmail)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body.String())

	s := gomail.NewDialer(g.gmailConfig.SmtpHost, g.gmailConfig.SmtpPort, g.gmailConfig.SenderEmail, g.gmailConfig.GmailAppPassword)

	if err := s.DialAndSend(m); err != nil {
		g.appLogger.ErrorContext(ctx, "Failed to send email", logger.Field{Key: "error", Value: err})
		return err
	}

	return nil
}
