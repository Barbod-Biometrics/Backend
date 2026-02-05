package sms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/communication"
)

const (
	serviceName  = "sam"
	messagerType = "bale"
)

type SMSService struct {
	apiKey       string
	baseURL      string
	backdoorCode string // backdoor code for testing
	client       *http.Client
}

type ferzzRequest struct {
	Token       string            `json:"token"`
	Destination string            `json:"destination"`
	Action      string            `json:"action"`
	Payload     map[string]string `json:"payload"`
}

type ferzzResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	TrackID string `json:"track_id"`
}

func NewSMSService(apiKey string, baseURL string, backdoorCode string) *SMSService {
	return &SMSService{
		apiKey:       apiKey,
		baseURL:      baseURL,
		backdoorCode: backdoorCode,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

}

var _ communication.SMSService = (*SMSService)(nil)

func (s *SMSService) Send(ctx context.Context, phoneNumber string, code string) error {

	// Backdoor
	if s.backdoorCode != "" {
		log.Printf("SMS Backdoor Active - You can use code %s for any phone number!!!", s.backdoorCode)
	}

	if s.apiKey == "" {
		log.Println("[Warning] SMS_GATEWAY_API_KEY is not configured, but Backdoor is OFF. SMS will fail.")
		return fmt.Errorf("sms provider not configured")
	}

	reqBody := ferzzRequest{
		Token:       s.apiKey,
		Destination: phoneNumber,
		Action:      serviceName,
		Payload: map[string]string{
			"messenger_type": messagerType,
			"code":           code,
		},
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal sms request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call sms provider: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sms provider returned non-200 status: %d", resp.StatusCode)
	}

	var result ferzzResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode sms response: %w", err)
	}

	log.Printf("SMS sent successfully to %s. TrackID: %s", phoneNumber, result.TrackID)

	return nil
}
