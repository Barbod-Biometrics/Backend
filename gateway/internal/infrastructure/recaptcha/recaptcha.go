package recaptcha

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Verifier interface {
	Verify(ctx context.Context, token string, remoteIP string) error
	IsEnabled() bool
}

type GoogleVerifier struct {
	secret    string
	verifyURL string
	enabled   bool
	client    *http.Client
}

type verifyResponse struct {
	Success     bool     `json:"success"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	Score       float64  `json:"score,omitempty"`
	Action      string   `json:"action,omitempty"`
	ErrorCodes  []string `json:"error-codes,omitempty"`
}

func NewGoogleVerifier(secret string, verifyURL string, enabled bool) *GoogleVerifier {
	return &GoogleVerifier{
		secret:    secret,
		verifyURL: verifyURL,
		enabled:   enabled,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (v *GoogleVerifier) IsEnabled() bool {
	return v != nil && v.enabled
}

func (v *GoogleVerifier) Verify(ctx context.Context, token string, remoteIP string) error {
	if !v.enabled {
		return nil
	}
	if v.secret == "" {
		return errors.New("recaptcha secret not configured")
	}
	if token == "" {
		return errors.New("captcha token is empty")
	}

	form := url.Values{}
	form.Set("secret", v.secret)
	form.Set("response", token)
	if remoteIP != "" {
		form.Set("remoteip", remoteIP)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.verifyURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("failed to build recaptcha request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := v.client.Do(req)
	if err != nil {
		return fmt.Errorf("recaptcha request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("recaptcha verification returned status %d", resp.StatusCode)
	}

	var vr verifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&vr); err != nil {
		return fmt.Errorf("failed to decode recaptcha response: %w", err)
	}

	if !vr.Success {
		return fmt.Errorf("recaptcha verification failed: %v", vr.ErrorCodes)
	}

	return nil
}
