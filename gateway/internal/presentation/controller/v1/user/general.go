package user

import (
	"net/http"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/auth"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/service"
	exception "github.com/Barbod-Biometrics/Backend/gateway/internal/domain/exception"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/middleware"
	"github.com/gin-gonic/gin"
)

type GeneralUserController struct {
	authUsecase *service.AuthUsecase
}

func NewAuthController(authUsecase *service.AuthUsecase) *GeneralUserController {
	return &GeneralUserController{authUsecase: authUsecase}
}

// @Summary Request OTP
// @Description Request a one-time password to be sent to a phone number
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body auth.RequestOTPRequest true "Request OTP Request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /auth/request-otp [post]
func (g *GeneralUserController) RequestOTPHandler(c *gin.Context) {
	var req auth.RequestOTPRequest

	// bind and validate
	if err := c.ShouldBindJSON(&req); err != nil {
		msg := "Invalid request body"
		if tr := middleware.GetTranslator(c); tr != nil {
			if tmsg, terr := tr.Translate("errors.invalid_request_body"); terr == nil && tmsg != "" {
				msg = tmsg
			}
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   msg,
			"details": err.Error(),
		})
		return
	}

	// call usecase
	if err := g.authUsecase.RequestOTP(c.Request.Context(), req); err != nil {
		msg := err.Error()
		if tr := middleware.GetTranslator(c); tr != nil {
			if tmsg, terr := tr.Translate("errors.generic"); terr == nil && tmsg != "" {
				msg = tmsg
			}
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": msg,
		})
		return
	}

	// send response
	message := "OTP sent successfully"
	if tr := middleware.GetTranslator(c); tr != nil {
		if tmsg, terr := tr.Translate("successMessage.phoneVerification"); terr == nil && tmsg != "" {
			message = tmsg
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"message": message,
	})

}

// @Summary Verify OTP
// @Description Verify OTP and return auth tokens
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body auth.VerifyOTPRequest true "Verify OTP Request"
// @Success 200 {object} auth.UserInfoResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/verify-otp [post]
func (g *GeneralUserController) VerifyOTPHandler(c *gin.Context) {
	var req auth.VerifyOTPRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		msg := "Invalid request body"
		if tr := middleware.GetTranslator(c); tr != nil {
			if tmsg, terr := tr.Translate("errors.invalid_request_body"); terr == nil && tmsg != "" {
				msg = tmsg
			}
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   msg,
			"details": err.Error(),
		})
		return
	}

	response, err := g.authUsecase.VerifyOTP(c.Request.Context(), req)
	if err != nil {
		// if this is an auth-related error, wrap as domain auth error when appropriate
		authErr := exception.NewUnauthorizedError("", err)
		msg := err.Error()
		if tr := middleware.GetTranslator(c); tr != nil {
			if tmsg, terr := tr.Translate("errors." + authErr.Type); terr == nil && tmsg != "" {
				msg = tmsg
			}
		}
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": msg,
			"type":  authErr.Type,
		})
		return
	}

	c.JSON(http.StatusOK, response)

}
