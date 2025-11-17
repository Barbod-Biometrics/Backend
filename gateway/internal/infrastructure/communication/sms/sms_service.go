package sms

import (
	"context"
	"log"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/communication"
)

type smsService struct {
	// client
	// fromnumber string
}

func NewSMSService( /* config */ ) communication.SMSService {
	return &smsService{
		// clident: client,
		// fromnumber: fromnumber,
	}

}

func (s *smsService) Send(ctx context.Context, phoneNumber string, message string) error {
	// we'll send the OTP from here and save it in the redis
	log.Printf("SMS to %s: %s\n", phoneNumber, message)

	return nil
}
