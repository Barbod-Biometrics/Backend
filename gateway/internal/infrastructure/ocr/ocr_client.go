package ocr

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	ocrDTO "github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/ocr"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/exception"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
)

const (
	defaultTimeout = 2 * time.Minute
)

type OCRClient struct {
	baseURL    string
	httpClient *http.Client
	logger     logger.Logger
}

func NewOCRClient(baseURL string, logger logger.Logger) *OCRClient {
	return &OCRClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		logger: logger,
	}
}

func (c *OCRClient) ExtractText(ctx context.Context, image []byte) (*ocrDTO.OCRResponse, error) {
	url := fmt.Sprintf("%s/chat/completions", c.baseURL)
	c.logger.Info("Sending OCR extract text request",
		logger.Field{Key: "url", Value: url},
	)

	if len(image) == 0 {
		c.logger.Warn("Image is empty")
		return nil, exception.ErrEmptyImage
	}

	b64 := base64.StdEncoding.EncodeToString(image)
	dataURL := "data:image/png;base64," + b64

	systemInstruction := "You are a helpful AI that outputs raw JSON only."
	userPrompt := "Extract the following information from this Iranian National ID card image. Return the result as a strictly valid JSON object using exactly these Persian keys: شماره_ملی, نام, نام_خانوادگی, نام_پدر, تاریخ_تولد, پایان_اعتبار. Do not include any explanation or markdown formatting. Double check that field names and what you extract is correct."

	reqBody := map[string]interface{}{
		"model": "mini",
		"messages": []interface{}{
			map[string]interface{}{"role": "system", "content": systemInstruction},
			map[string]interface{}{
				"role": "user",
				"content": []interface{}{
					map[string]interface{}{"type": "text", "text": userPrompt},
					map[string]interface{}{"type": "image_url", "image_url": map[string]interface{}{"url": dataURL}},
				},
			},
		},
		"temperature": 0.1,
		"max_tokens":  1024,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		c.logger.Error("Failed to marshal OCR request",
			logger.Field{Key: "error", Value: err},
		)
		return nil, exception.NewInternalError("ERR_OCR_MARSHAL", "failed to marshal ocr request", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		c.logger.Error("Failed to create HTTP request",
			logger.Field{Key: "error", Value: err},
		)
		return nil, exception.NewInternalError("ERR_OCR_CREATE_REQ", "failed to create http request", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("OCR request failed",
			logger.Field{Key: "error", Value: err},
		)
		return nil, exception.NewInternalError("ERR_OCR_SEND_REQ", "failed to send ocr request", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("Failed to read response body",
			logger.Field{Key: "error", Value: err},
		)
		return nil, exception.NewInternalError("ERR_OCR_READ_RESP", "failed to read response body", err)
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("OCR service returned non-OK status",
			logger.Field{Key: "status", Value: resp.StatusCode},
			logger.Field{Key: "body", Value: string(respBody)},
		)
		return nil, exception.NewInternalError("ERR_OCR_SERVICE_ERROR", fmt.Sprintf("OCR service error: %d", resp.StatusCode), nil)
	}

	var chatResp struct {
		Choices []struct {
			Message struct {
				Content json.RawMessage `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		c.logger.Error("Failed to parse OCR response",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "body", Value: string(respBody)},
		)
		return nil, exception.NewInternalError("ERR_OCR_PARSE_RESP", "failed to parse ocr response", err)
	}

	if len(chatResp.Choices) == 0 {
		c.logger.Error("OCR response has no choices", logger.Field{Key: "body", Value: string(respBody)})
		return nil, exception.NewInternalError("ERR_OCR_NO_CHOICES", "no choices in ocr response", nil)
	}

	contentRaw := chatResp.Choices[0].Message.Content

	// The content may be a JSON object or a JSON string containing JSON.
	// First try to unmarshal directly into OCRResponse.
	var result ocrDTO.OCRResponse
	if err := json.Unmarshal(contentRaw, &result); err == nil {
		result.Success = true
		c.logger.Info("OCR extraction completed", logger.Field{Key: "success", Value: result.Success})
		return &result, nil
	}

	// If direct unmarshal failed, try to unwrap string and parse it.
	var asString string
	if err := json.Unmarshal(contentRaw, &asString); err == nil {
		// asString should contain JSON text; try to parse it
		if err := json.Unmarshal([]byte(asString), &result); err == nil {
			result.Success = true
			c.logger.Info("OCR extraction completed (unwrapped string)", logger.Field{Key: "success", Value: result.Success})
			return &result, nil
		}
	}

	// Fallback: attempt to extract a plain text body and return as message
	c.logger.Error("Failed to decode OCR content into expected DTO",
		logger.Field{Key: "body", Value: string(contentRaw)},
	)
	return nil, exception.NewInternalError("ERR_OCR_DECODE_CONTENT", "failed to decode ocr content", nil)
}

func (c *OCRClient) HealthCheck(ctx context.Context) (*ocrDTO.HealthCheckResponseDTO, error) {
	url := fmt.Sprintf("%s/health", c.baseURL)
	c.logger.Info("Sending OCR health check request",
		logger.Field{Key: "url", Value: url},
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		c.logger.Error("Failed to create health check request",
			logger.Field{Key: "error", Value: err},
		)
		return nil, fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Health check request failed",
			logger.Field{Key: "error", Value: err},
		)
		return nil, fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("Failed to read health check response",
			logger.Field{Key: "error", Value: err},
		)
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result ocrDTO.HealthCheckResponseDTO
	if err := json.Unmarshal(respBody, &result); err != nil {
		c.logger.Error("Failed to parse health check response",
			logger.Field{Key: "error", Value: err},
		)
		return nil, fmt.Errorf("failed to parse health check response: %w", err)
	}

	c.logger.Info("OCR health check status", logger.Field{Key: "status", Value: result.Status})
	return &result, nil
}
