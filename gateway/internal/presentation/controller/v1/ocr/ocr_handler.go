package ocr

import (
	"errors"
	"io"
	"net/http"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/ocr"
	ocrDto "github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/ocr"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/exception"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/middleware"
	"github.com/gin-gonic/gin"
)

type OCRHandler struct {
	ocrUsecase usecase.OCRUsecase
	logger     logger.Logger
}

func NewOCRHandler(ocrUsecase usecase.OCRUsecase, logger logger.Logger) *OCRHandler {
	return &OCRHandler{ocrUsecase: ocrUsecase, logger: logger}
}

// getUserID extracts the authenticated user's ID from the request context.
// The user ID is set by the JWT middleware after validating the access token.
func (h *OCRHandler) getUserID(c *gin.Context) uint64 {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	}
	return userID
}

// ExtractText godoc
// @Router /ocr/extract [post]
// @Summary Extract text from image using OCR
// @Description Extract text content from an uploaded image using OCR technology
// @Tags OCR - Machine
// @Security ApiKeyAuth
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "Image file for text extraction"
// @Success 200 {object} ocr.OCRResponse "Extraction successful"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
func (h *OCRHandler) ExtractText(c *gin.Context) {
	translator := middleware.GetTranslator(c)
	profileID, exists := c.Get("profile_id")
	if !exists {
		h.logger.Warn("Unauthorized OCR request",
			logger.Field{Key: "ip", Value: c.ClientIP()},
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	h.logger.Info("Received OCR extract text request",
		logger.Field{Key: "ip", Value: c.ClientIP()},
		logger.Field{Key: "profile_id", Value: profileID},
	)

	imageFile, err := c.FormFile("image")
	if err != nil {
		h.logger.Warn("Failed to get image file",
			logger.Field{Key: "error", Value: err},
		)
		msg := "Please provide an 'image' file"
		if translator != nil {
			if translated, err := translator.Translate("validation.missing_image"); err == nil {
				msg = translated
			}
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_image", "message": msg})
		return
	}

	imageSrc, err := imageFile.Open()
	if err != nil {
		h.logger.Error("Failed to open image file",
			logger.Field{Key: "error", Value: err},
		)
		msg := "Failed to process image file"
		if translator != nil {
			if translated, err := translator.Translate("errors.file_error"); err == nil {
				msg = translated
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "file_error", "message": msg})
		return
	}
	defer imageSrc.Close()

	imageBytes, err := io.ReadAll(imageSrc)
	if err != nil {
		h.logger.Error("Failed to read image bytes",
			logger.Field{Key: "error", Value: err},
		)
		msg := "Failed to read image file"
		if translator != nil {
			if translated, err := translator.Translate("errors.file_read_error"); err == nil {
				msg = translated
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "file_error", "message": msg})
		return
	}

	var pid uint64
	if v, ok := profileID.(uint64); ok {
		pid = v
	}

	// Get client IP
	clientIP := c.ClientIP()

	var result *ocrDto.OCRResponse
	result, err = h.ocrUsecase.ExtractTextWithIP(c.Request.Context(), pid, imageBytes, clientIP)

	if err != nil {
		if errors.Is(err, exception.ErrTrialExceeded) && result != nil {
			c.JSON(http.StatusTooManyRequests, result)
			return
		}

		h.logger.Error("OCR extraction failed",
			logger.Field{Key: "profile_id", Value: profileID},
			logger.Field{Key: "error", Value: err},
		)
		panic(err)
	}

	statusCode := http.StatusOK
	if !result.Success {
		h.logger.Warn("OCR extraction returned failure",
			logger.Field{Key: "profile_id", Value: profileID},
			logger.Field{Key: "message", Value: result.Message},
		)
		statusCode = http.StatusBadRequest
	} else {
		h.logger.Info("OCR extraction completed successfully",
			logger.Field{Key: "profile_id", Value: profileID},
		)
	}

	c.JSON(statusCode, result)
}

// DemoExtract godoc
// @Summary Demo OCR extraction (limited trials per IP)
// @Description Public demo endpoint for OCR without API key. Limited to a small number of trials per IP address.
// @Tags Demo
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "Image file for text extraction"
// @Success 200 {object} ocr.OCRResponse "Extraction successful"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 429 {object} map[string]interface{} "Too many requests"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /demo/ocr/extract [post]
func (h *OCRHandler) DemoExtract(c *gin.Context) {
	translator := middleware.GetTranslator(c)

	h.logger.Info("Received OCR demo extract request",
		logger.Field{Key: "ip", Value: c.ClientIP()},
	)

	imageFile, err := c.FormFile("image")
	if err != nil {
		h.logger.Warn("Failed to get image file",
			logger.Field{Key: "error", Value: err},
		)
		msg := "Please provide an 'image' file"
		if translator != nil {
			if translated, err := translator.Translate("validation.missing_image"); err == nil {
				msg = translated
			}
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_image", "message": msg})
		return
	}

	imageSrc, err := imageFile.Open()
	if err != nil {
		h.logger.Error("Failed to open image file",
			logger.Field{Key: "error", Value: err},
		)
		msg := "Failed to process image file"
		if translator != nil {
			if translated, err := translator.Translate("errors.file_error"); err == nil {
				msg = translated
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "file_error", "message": msg})
		return
	}
	defer imageSrc.Close()

	imageBytes, err := io.ReadAll(imageSrc)
	if err != nil {
		h.logger.Error("Failed to read image bytes",
			logger.Field{Key: "error", Value: err},
		)
		msg := "Failed to read image file"
		if translator != nil {
			if translated, err := translator.Translate("errors.file_read_error"); err == nil {
				msg = translated
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "file_error", "message": msg})
		return
	}

	// Call service with IP for trial tracking (profileID = 0)
	var pid uint64 = 0
	clientIP := c.ClientIP()

	var result *ocrDto.OCRResponse
	result, err = h.ocrUsecase.ExtractTextWithIP(c.Request.Context(), pid, imageBytes, clientIP)

	if err != nil {
		if errors.Is(err, exception.ErrTrialExceeded) && result != nil {
			c.JSON(http.StatusTooManyRequests, result)
			return
		}

		h.logger.Error("OCR demo extraction failed",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "ip", Value: clientIP},
		)
		panic(err)
	}

	statusCode := http.StatusOK
	if !result.Success {
		h.logger.Warn("OCR demo extraction returned failure",
			logger.Field{Key: "message", Value: result.Message},
			logger.Field{Key: "ip", Value: clientIP},
		)
		statusCode = http.StatusBadRequest
	}

	c.JSON(statusCode, result)
}

// HealthCheck godoc
// @Router /ocr/health [get]
// @Summary Check OCR service health
// @Description Check if the OCR service is running and healthy
// @Tags OCR - Machine
// @Produce json
// @Success 200 {object} ocr.HealthCheckResponseDTO "Service is healthy"
// @Failure 500 {object} map[string]interface{} "Service is unhealthy"
func (h *OCRHandler) HealthCheck(c *gin.Context) {
	h.logger.Info("Received OCR health check request",
		logger.Field{Key: "ip", Value: c.ClientIP()},
	)

	result, err := h.ocrUsecase.HealthCheck(c.Request.Context())
	if err != nil {
		h.logger.Error("OCR health check failed",
			logger.Field{Key: "error", Value: err},
		)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "OCR service is not responding", "status": "unhealthy"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// ApproveResult approves a specific OCR record
// @Summary Approve OCR Result
// @Description Marks an OCR result as verified/approved by the user
// @Tags OCR - User
// @Accept json
// @Produce json
// @Param request body ocr.ApproveOCRRequest true "Approval Request"
// @Success 200 {object} map[string]string "Success"
// @Failure 400 {object} map[string]string "Invalid Request"
// @Failure 403 {object} map[string]string "Forbidden - Profile does not belong to user"
// @Failure 404 {object} map[string]string "Record not found"
// @Router /api/v1/ocr/approve [post]
func (h *OCRHandler) ApproveResult(c *gin.Context) {

	var req ocr.ApproveOCRRequest
	if err := c.ShouldBind(&req); err != nil {
		h.logger.Error("Invalid request body/structure for approving the OCR record", logger.Field{Key: "error", Value: err})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	userID := h.getUserID(c)
	if userID == 0 {
		h.logger.Warn("Unauthorized attempt to access OCR approval API")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	err := h.ocrUsecase.ApproveResult(c.Request.Context(), userID, req.ID, req.ProfileID)
	if err != nil {
		switch err.Error() {
		case "record not found or access denied":
			c.JSON(http.StatusNotFound, gin.H{"error": "OCR record not found"})
		case "profile access denied":
			c.JSON(http.StatusForbidden, gin.H{"error": "You do not own this profile"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "OCR result approved successfully",
	})
}
