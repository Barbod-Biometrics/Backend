package face_verification

import (
	"errors"
	"io"
	"net/http"

	faceVerificationDto "github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/face_verification"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/exception"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/logger"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/middleware"
	"github.com/gin-gonic/gin"
)

// FaceVerificationHandler handles requests related to face verification service operations
type FaceVerificationHandler struct {
	faceVerificationUsecase usecase.FaceVerificationUsecase
	logger                  logger.Logger
}

// NewFaceVerificationHandler creates a new face verification handler
func NewFaceVerificationHandler(faceVerificationUsecase usecase.FaceVerificationUsecase, logger logger.Logger) *FaceVerificationHandler {
	return &FaceVerificationHandler{
		faceVerificationUsecase: faceVerificationUsecase,
		logger:                  logger,
	}
}

// VerifyFace godoc
// @Summary Verify face in video matches photo
// @Description Verify if the face detected in a video matches the provided reference photo
// @Tags Face-Verification
// @Security ApiKeyAuth
// @Accept multipart/form-data
// @Produce json
// @Param photo formData file true "Reference photo"
// @Param video formData file true "Video file for verification"
// @Success 200 {object} face_verification.FaceVerificationResponse "Verification successful"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /face-verification/verify [post]
func (fv *FaceVerificationHandler) VerifyFace(c *gin.Context) {
	translator := middleware.GetTranslator(c)
	profileID, exists := c.Get("profile_id")
	if !exists {
		fv.logger.Warn("Unauthorized face verification request",
			logger.Field{Key: "ip", Value: c.ClientIP()},
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	fv.logger.Info("Received face verification request",
		logger.Field{Key: "profile_id", Value: profileID},
		logger.Field{Key: "ip", Value: c.ClientIP()},
	)

	photoFile, err := c.FormFile("photo")
	if err != nil {
		fv.logger.Warn("Failed to get photo file",
			logger.Field{Key: "error", Value: err},
		)
		msg := "Please provide a 'photo' file"
		if translator != nil {
			if translated, err := translator.Translate("validation.missing_photo"); err == nil {
				msg = translated
			}
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_photo", "message": msg})
		return
	}

	videoFile, err := c.FormFile("video")
	if err != nil {
		fv.logger.Warn("Failed to get video file",
			logger.Field{Key: "error", Value: err},
		)
		msg := "Please provide a 'video' file"
		if translator != nil {
			if translated, err := translator.Translate("validation.missing_video"); err == nil {
				msg = translated
			}
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_video", "message": msg})
		return
	}

	photoSrc, err := photoFile.Open()
	if err != nil {
		fv.logger.Error("Failed to open photo file",
			logger.Field{Key: "error", Value: err},
		)
		msg := "Failed to process photo file"
		if translator != nil {
			if translated, err := translator.Translate("errors.file_error"); err == nil {
				msg = translated
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "file_error", "message": msg})
		return
	}
	defer photoSrc.Close()

	photoBytes, err := io.ReadAll(photoSrc)
	if err != nil {
		fv.logger.Error("Failed to read photo bytes",
			logger.Field{Key: "error", Value: err},
		)
		msg := "Failed to read photo file"
		if translator != nil {
			if translated, err := translator.Translate("errors.file_read_error"); err == nil {
				msg = translated
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "file_error", "message": msg})
		return
	}

	videoSrc, err := videoFile.Open()
	if err != nil {
		fv.logger.Error("Failed to open video file",
			logger.Field{Key: "error", Value: err},
		)
		msg := "Failed to process video file"
		if translator != nil {
			if translated, err := translator.Translate("errors.file_error"); err == nil {
				msg = translated
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "file_error", "message": msg})
		return
	}
	defer videoSrc.Close()

	videoBytes, err := io.ReadAll(videoSrc)
	if err != nil {
		fv.logger.Error("Failed to read video bytes",
			logger.Field{Key: "error", Value: err},
		)
		msg := "Failed to read video file"
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

	var result *faceVerificationDto.FaceVerificationResponse
	result, err = fv.faceVerificationUsecase.VerifyFaceWithIP(c.Request.Context(), pid, photoBytes, videoBytes, clientIP)

	if err != nil {
		if errors.Is(err, exception.ErrTrialExceeded) && result != nil {
			c.JSON(http.StatusTooManyRequests, result)
			return
		}

		fv.logger.Error("Face verification failed",
			logger.Field{Key: "profile_id", Value: profileID},
			logger.Field{Key: "error", Value: err},
		)
		panic(err)
	}

	statusCode := http.StatusOK
	if !result.Success {
		fv.logger.Warn("Face verification returned failure",
			logger.Field{Key: "profile_id", Value: profileID},
			logger.Field{Key: "reason", Value: result.Reason},
		)
		statusCode = http.StatusBadRequest
	} else {
		fv.logger.Info("Face verification completed successfully",
			logger.Field{Key: "profile_id", Value: profileID},
		)
	}

	c.JSON(statusCode, result)
}

// DemoVerify godoc
// @Summary Demo face verification (limited trials per IP)
// @Description Public demo endpoint for face verification without API key. Limited to a small number of trials per IP address.
// @Tags Demo
// @Accept multipart/form-data
// @Produce json
// @Param photo formData file true "Reference photo"
// @Param video formData file true "Video file for verification"
// @Success 200 {object} face_verification.FaceVerificationResponse "Verification successful"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 429 {object} map[string]interface{} "Too many requests"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /demo/face-verification/verify [post]
func (fv *FaceVerificationHandler) DemoVerify(c *gin.Context) {
	translator := middleware.GetTranslator(c)

	fv.logger.Info("Received face verification demo request",
		logger.Field{Key: "ip", Value: c.ClientIP()},
	)

	photoFile, err := c.FormFile("photo")
	if err != nil {
		fv.logger.Warn("Failed to get photo file",
			logger.Field{Key: "error", Value: err},
		)
		msg := "Please provide a 'photo' file"
		if translator != nil {
			if translated, err := translator.Translate("validation.missing_photo"); err == nil {
				msg = translated
			}
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_photo", "message": msg})
		return
	}

	videoFile, err := c.FormFile("video")
	if err != nil {
		fv.logger.Warn("Failed to get video file",
			logger.Field{Key: "error", Value: err},
		)
		msg := "Please provide a 'video' file"
		if translator != nil {
			if translated, err := translator.Translate("validation.missing_video"); err == nil {
				msg = translated
			}
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_video", "message": msg})
		return
	}

	photoSrc, err := photoFile.Open()
	if err != nil {
		fv.logger.Error("Failed to open photo file",
			logger.Field{Key: "error", Value: err},
		)
		msg := "Failed to process photo file"
		if translator != nil {
			if translated, err := translator.Translate("errors.file_error"); err == nil {
				msg = translated
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "file_error", "message": msg})
		return
	}
	defer photoSrc.Close()

	photoBytes, err := io.ReadAll(photoSrc)
	if err != nil {
		fv.logger.Error("Failed to read photo bytes",
			logger.Field{Key: "error", Value: err},
		)
		msg := "Failed to read photo file"
		if translator != nil {
			if translated, err := translator.Translate("errors.file_read_error"); err == nil {
				msg = translated
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "file_error", "message": msg})
		return
	}

	videoSrc, err := videoFile.Open()
	if err != nil {
		fv.logger.Error("Failed to open video file",
			logger.Field{Key: "error", Value: err},
		)
		msg := "Failed to process video file"
		if translator != nil {
			if translated, err := translator.Translate("errors.file_error"); err == nil {
				msg = translated
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "file_error", "message": msg})
		return
	}
	defer videoSrc.Close()

	videoBytes, err := io.ReadAll(videoSrc)
	if err != nil {
		fv.logger.Error("Failed to read video bytes",
			logger.Field{Key: "error", Value: err},
		)
		msg := "Failed to read video file"
		if translator != nil {
			if translated, err := translator.Translate("errors.file_read_error"); err == nil {
				msg = translated
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "file_error", "message": msg})
		return
	}

	var pid uint64 = 0
	clientIP := c.ClientIP()

	var result *faceVerificationDto.FaceVerificationResponse
	result, err = fv.faceVerificationUsecase.VerifyFaceWithIP(c.Request.Context(), pid, photoBytes, videoBytes, clientIP)

	if err != nil {
		if errors.Is(err, exception.ErrTrialExceeded) && result != nil {
			c.JSON(http.StatusTooManyRequests, result)
			return
		}

		fv.logger.Error("Face verification demo failed",
			logger.Field{Key: "error", Value: err},
			logger.Field{Key: "ip", Value: clientIP},
		)
		panic(err)
	}

	statusCode := http.StatusOK
	if !result.Success {
		fv.logger.Warn("Face verification demo returned failure",
			logger.Field{Key: "reason", Value: result.Reason},
			logger.Field{Key: "ip", Value: clientIP},
		)
		statusCode = http.StatusBadRequest
	}

	c.JSON(statusCode, result)
}

// CropImage godoc
// @Summary Crop image to passport-style photo
// @Description Crop the provided image to a passport-style photo format
// @Tags Face-Verification
// @Security ApiKeyAuth
// @Accept multipart/form-data
// @Produce json,image/jpeg
// @Param image formData file true "Image to crop"
// @Success 200 {file} file "Cropped image"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /face-verification/crop [post]
func (fv *FaceVerificationHandler) CropImage(c *gin.Context) {
	translator := middleware.GetTranslator(c)
	profileID, exists := c.Get("profile_id")
	if !exists {
		fv.logger.Warn("Unauthorized crop image request",
			logger.Field{Key: "ip", Value: c.ClientIP()},
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	fv.logger.Info("Received crop image request",
		logger.Field{Key: "profile_id", Value: profileID},
		logger.Field{Key: "ip", Value: c.ClientIP()},
	)

	imageFile, err := c.FormFile("image")
	if err != nil {
		fv.logger.Warn("Failed to get image file",
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
		fv.logger.Error("Failed to open image file",
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
		fv.logger.Error("Failed to read image bytes",
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

	profID, _ := profileID.(uint64)
	result, err := fv.faceVerificationUsecase.CropImage(c.Request.Context(), profID, imageBytes)
	if err != nil {
		fv.logger.Error("Image crop failed",
			logger.Field{Key: "profile_id", Value: profileID},
			logger.Field{Key: "error", Value: err},
		)
		panic(err)
	}

	if !result.Success {
		fv.logger.Warn("Image crop returned failure",
			logger.Field{Key: "profile_id", Value: profileID},
			logger.Field{Key: "error", Value: result.Error},
		)
		msg := result.Error
		if translator != nil {
			if translated, err := translator.Translate("errors.crop_error"); err == nil {
				msg = translated
			}
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "crop_error", "message": msg})
		return
	}

	fv.logger.Info("Image crop completed successfully",
		logger.Field{Key: "profile_id", Value: profileID},
	)
	c.Data(http.StatusOK, "image/jpeg", result.Image)
}

// HealthCheck godoc
// @Summary Check models service health
// @Description Check if the models service is healthy and available
// @Tags Face-Verification
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} face_verification.HealthCheckResponseDTO "Service is healthy"
// @Failure 500 {object} map[string]interface{} "Service error"
// @Router /face-verification/health [get]
func (fv *FaceVerificationHandler) HealthCheck(c *gin.Context) {
	fv.logger.Info("Received health check request",
		logger.Field{Key: "ip", Value: c.ClientIP()},
	)

	result, err := fv.faceVerificationUsecase.HealthCheck(c.Request.Context())
	if err != nil {
		fv.logger.Error("Health check failed",
			logger.Field{Key: "error", Value: err},
		)
		msg := "Health check failed"
		translator := middleware.GetTranslator(c)
		if translator != nil {
			if translated, err := translator.Translate("errors.health_check_error"); err == nil {
				msg = translated
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "health_check_error", "message": msg})
		return
	}

	fv.logger.Info("Health check completed",
		logger.Field{Key: "status", Value: result.Status},
	)
	c.JSON(http.StatusOK, result)
}

// GetReport returns the data for the frontend chart/table
// @Summary Get Face Verification Report
// @Tags Model-Reports
// @Accept json
// @Produce json
// @Param profile_id query int true "Profile ID"
// @Param page query int false "Page" default(1)
// @Param limit query int false "Limit" default(20)
// @Param status query string false "Status (success, failed)"
// @Param from_date query string false "From Date (YYYY-MM-DD)"
// @Param to_date query string false "To Date (YYYY-MM-DD)"
// @Param sort_by query string false "Sort By (date, rate)"
// @Param sort_order query string false "Order (asc, desc)"
// @Success 200 {object} face_verification.FaceReportResponse
// @Failure 400 {object} map[string]string "Invalid Parameters"
// @Failure 404 {object} map[string]string "Profile Not Found"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /api/v1/report/face-verification [get]
func (h *FaceVerificationHandler) GetReport(c *gin.Context) {
	var req faceVerificationDto.GetFaceReportRequest

	// Bind Query Params
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parameters"})
		return
	}

	// Call UseCase
	resp, err := h.faceVerificationUsecase.GetReports(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
