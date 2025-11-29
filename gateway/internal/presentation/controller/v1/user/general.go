package user

import (
	"net/http"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/auth"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/service"
	"github.com/gin-gonic/gin"
)

type GeneralUserController struct {
	authUsecase *service.AuthUsecase
}

func NewAuthController(authUsecase *service.AuthUsecase) *GeneralUserController {
	return &GeneralUserController{authUsecase: authUsecase}
}

func (g *GeneralUserController) RequestOTPHandler(c *gin.Context) {
	var req auth.RequestOTPRequest

	// bind and validate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// call usecase
	if err := g.authUsecase.RequestOTP(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// send response
	c.JSON(http.StatusOK, gin.H{
		"message": "OTP sent successfully",
	})

}

func (g *GeneralUserController) VerifyOTPHandler(c *gin.Context) {
	var req auth.VerifyOTPRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	response, err := g.authUsecase.VerifyOTP(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)

}
