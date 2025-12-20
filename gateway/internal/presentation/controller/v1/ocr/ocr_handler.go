package ocr

import (
    "io"
    "net/http"

    "github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
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

// ExtractText godoc
// @Router /ocr/extract [post]
// @Summary Extract text from image using OCR
// @Description Extract text content from an uploaded image using OCR technology
// @Tags OCR
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

    result, err := h.ocrUsecase.ExtractText(c.Request.Context(), pid, imageBytes)
    if err != nil {
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

// HealthCheck godoc
// @Router /ocr/health [get]
// @Summary Check OCR service health
// @Description Check if the OCR service is running and healthy
// @Tags OCR
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
