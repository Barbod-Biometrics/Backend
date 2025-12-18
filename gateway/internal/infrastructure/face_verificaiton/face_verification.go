package face_verification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	faceVerificationDTO "github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/face_verification"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/exception"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
)

const (
	defaultTimeout = 5 * time.Minute
)

type FaceVerificationClient struct {
	baseURL    string
	httpClient *http.Client
	logger     logger.Logger
}

func NewFaceVerificationClient(baseURL string, logger logger.Logger) *FaceVerificationClient {
	return &FaceVerificationClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		logger: logger,
	}
}

func (c *FaceVerificationClient) VerifyFace(ctx context.Context, photo []byte, video []byte) (*faceVerificationDTO.FaceVerificationResponse, error) {
	url := fmt.Sprintf("%s/verify", c.baseURL)
	c.logger.Info("Sending face verification request",
		logger.Field{Key: "url", Value: url},
	)

	if len(photo) == 0 {
		c.logger.Warn("Photo is empty")
		return nil, exception.ErrEmptyPhoto
	}

	if len(video) == 0 {
		c.logger.Warn("Video is empty")
		return nil, exception.ErrEmptyVideo
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	photoPart, err := writer.CreateFormFile("photo", "photo.jpg")
	if err != nil {
		c.logger.Error("Failed to create photo form field",
			logger.Field{Key: "error", Value: err},
		)
		return nil, exception.NewInternalError("ERR_FACE_CREATE_PHOTO_FIELD", "failed to create photo form field", err)
	}
	if _, err := io.Copy(photoPart, bytes.NewReader(photo)); err != nil {
		c.logger.Error("Failed to copy photo data",
			logger.Field{Key: "error", Value: err},
		)
		return nil, exception.NewInternalError("ERR_FACE_COPY_PHOTO", "failed to copy photo data", err)
	}

	videoPart, err := writer.CreateFormFile("video", "video.mp4")
	if err != nil {
		c.logger.Error("Failed to create video form field",
			logger.Field{Key: "error", Value: err},
		)
		return nil, exception.NewInternalError("ERR_FACE_CREATE_VIDEO_FIELD", "failed to create video form field", err)
	}
	if _, err := io.Copy(videoPart, bytes.NewReader(video)); err != nil {
		c.logger.Error("Failed to copy video data",
			logger.Field{Key: "error", Value: err},
		)
		return nil, exception.NewInternalError("ERR_FACE_COPY_VIDEO", "failed to copy video data", err)
	}

	if err := writer.Close(); err != nil {
		c.logger.Error("Failed to close multipart writer",
			logger.Field{Key: "error", Value: err},
		)
		return nil, exception.NewInternalError("ERR_FACE_CLOSE_WRITER", "failed to close multipart writer", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		c.logger.Error("Failed to create HTTP request",
			logger.Field{Key: "error", Value: err},
		)
		return nil, exception.NewInternalError("ERR_FACE_CREATE_REQUEST", "failed to create HTTP request", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to send face verification request",
			logger.Field{Key: "error", Value: err},
		)
		return nil, exception.NewInternalError("ERR_FACE_SEND_REQUEST", "failed to send face verification request", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("Failed to read response body",
			logger.Field{Key: "error", Value: err},
		)
		return nil, exception.NewInternalError("ERR_FACE_READ_RESPONSE", "failed to read response body", err)
	}

	// Parse response
	var result faceVerificationDTO.FaceVerificationResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		c.logger.Error("Failed to parse face verification response",
			logger.Field{Key: "error", Value: err},
		)
		return nil, exception.NewInternalError("ERR_FACE_PARSE_RESPONSE", "failed to parse face verification response", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
		c.logger.Warn("Face verification returned non-OK status",
			logger.Field{Key: "status", Value: resp.StatusCode},
		)
	}

	c.logger.Info("Face verification response",
		logger.Field{Key: "success", Value: result.Success},
		logger.Field{Key: "reason", Value: result.Reason},
	)
	return &result, nil
}

func (c *FaceVerificationClient) CropImage(ctx context.Context, image []byte) (*faceVerificationDTO.CropImageResponse, error) {
	url := fmt.Sprintf("%s/crop", c.baseURL)
	c.logger.Info("Sending crop image request",
		logger.Field{Key: "url", Value: url},
	)

	if len(image) == 0 {
		c.logger.Warn("Image is empty")
		return nil, exception.ErrEmptyImage
	}

	// Create multipart request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add image field
	imagePart, err := writer.CreateFormFile("image", "image.jpg")
	if err != nil {
		c.logger.Error("Failed to create image form field",
			logger.Field{Key: "error", Value: err},
		)
		return nil, exception.NewInternalError("ERR_CROP_CREATE_IMAGE_FIELD", "failed to create image form field", err)
	}
	if _, err := io.Copy(imagePart, bytes.NewReader(image)); err != nil {
		c.logger.Error("Failed to copy image data",
			logger.Field{Key: "error", Value: err},
		)
		return nil, exception.NewInternalError("ERR_CROP_COPY_IMAGE", "failed to copy image data", err)
	}

	if err := writer.Close(); err != nil {
		c.logger.Error("Failed to close multipart writer",
			logger.Field{Key: "error", Value: err},
		)
		return nil, exception.NewInternalError("ERR_CROP_CLOSE_WRITER", "failed to close multipart writer", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		c.logger.Error("Failed to create HTTP request",
			logger.Field{Key: "error", Value: err},
		)
		return nil, exception.NewInternalError("ERR_CROP_CREATE_REQUEST", "failed to create HTTP request", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to send crop image request",
			logger.Field{Key: "error", Value: err},
		)
		return nil, exception.NewInternalError("ERR_CROP_SEND_REQUEST", "failed to send crop image request", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("Failed to read response body",
			logger.Field{Key: "error", Value: err},
		)
		return nil, exception.NewInternalError("ERR_CROP_READ_RESPONSE", "failed to read response body", err)
	}

	// Handle binary image response or JSON error
	if resp.StatusCode == http.StatusOK && resp.Header.Get("Content-Type") == "image/jpeg" {
		c.logger.Info("Crop image successful",
			logger.Field{Key: "url", Value: url},
		)
		return &faceVerificationDTO.CropImageResponse{
			Success: true,
			Message: "Image cropped successfully",
			Image:   respBody,
		}, nil
	}

	// Try to parse as JSON error response
	var result faceVerificationDTO.CropImageResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		c.logger.Error("Failed to parse crop image response",
			logger.Field{Key: "error", Value: err},
		)
		return nil, exception.NewInternalError("ERR_CROP_PARSE_RESPONSE", "failed to parse crop image response", err)
	}

	c.logger.Warn("Crop image failed",
		logger.Field{Key: "error", Value: result.Error},
	)
	return &result, nil
}

func (c *FaceVerificationClient) HealthCheck(ctx context.Context) (*faceVerificationDTO.HealthCheckResponseDTO, error) {
	url := fmt.Sprintf("%s/health", c.baseURL)
	c.logger.Info("Sending health check request",
		logger.Field{Key: "url", Value: url},
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		c.logger.Error("Failed to create HTTP request",
			logger.Field{Key: "error", Value: err},
		)
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Failed to send health check request",
			logger.Field{Key: "error", Value: err},
		)
		return nil, fmt.Errorf("failed to send health check request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("Failed to read response body",
			logger.Field{Key: "error", Value: err},
		)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var result faceVerificationDTO.HealthCheckResponseDTO
	if err := json.Unmarshal(respBody, &result); err != nil {
		c.logger.Error("Failed to parse health check response",
			logger.Field{Key: "error", Value: err},
		)
		return nil, fmt.Errorf("failed to parse health check response: %w", err)
	}

	c.logger.Info("Health check status",
		logger.Field{Key: "status", Value: result.Status},
	)
	return &result, nil
}
