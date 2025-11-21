package sms

import (
	"context"
	"log"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/communication"
)

type smsService struct {
	apiKey       string
	backdoorCode string // backdoor code for testing
	// provider SMSProvider
}

func NewSMSService(apiKey string, backdoorCode string) communication.SMSService {
	return &smsService{
		apiKey:       apiKey,
		backdoorCode: backdoorCode,
	}

}

func (s *smsService) Send(ctx context.Context, phoneNumber string, message string) error {

	// Backdoor
	if s.backdoorCode != "" {
		log.Printf("SMS Backdoor Active - You can use code %s for any phone number!!!", s.backdoorCode)
	}

	// we'll send the OTP from here and save it in the redis
	log.Printf("SMS to %s: %s\n", phoneNumber, message)

	if s.apiKey == "" {
		log.Println("Warning: SMS_GATEWAY_API_KEY not configured")
	}

	return nil
}
