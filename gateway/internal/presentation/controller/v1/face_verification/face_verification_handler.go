package face_verification

import (
	"io"
	"net/http"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
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
// @Success 200 {object} models.FaceVerificationResponse "Verification successful"
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

	result, err := fv.faceVerificationUsecase.VerifyFace(c.Request.Context(), pid, photoBytes, videoBytes)
	if err != nil {
		fv.logger.Error("Face verification failed",
			logger.Field{Key: "profile_id", Value: profileID},
			logger.Field{Key: "error", Value: err},
		)
		msg := "Face verification failed"
		if translator != nil {
			if translated, err := translator.Translate("errors.verification_error"); err == nil {
				msg = translated
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "verification_error", "message": msg})
		return
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

	result, err := fv.faceVerificationUsecase.CropImage(c.Request.Context(), imageBytes)
	if err != nil {
		fv.logger.Error("Image crop failed",
			logger.Field{Key: "profile_id", Value: profileID},
			logger.Field{Key: "error", Value: err},
		)
		msg := "Image crop failed"
		if translator != nil {
			if translated, err := translator.Translate("errors.crop_error"); err == nil {
				msg = translated
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "crop_error", "message": msg})
		return
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
// @Success 200 {object} models.HealthCheckResponseDTO "Service is healthy"
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
