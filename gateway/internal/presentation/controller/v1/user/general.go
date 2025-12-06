package user

import (
	"net/http"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/auth"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/user"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/service"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/middleware"
	"github.com/gin-gonic/gin"
)

type GeneralUserController struct {
	authUsecase *service.AuthUsecase
	userUsecase usecase.UserService
}

func NewUserController(authUsecase *service.AuthUsecase, userUsecase service.UserService) *GeneralUserController {
	return &GeneralUserController{
		authUsecase: authUsecase,
		userUsecase: &userUsecase,
	}
}

// @Summary Get User profile
// @Description Get the profile info of the currently logged-in user
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} user.UserInfoResponse
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /user/info [get]
func (g *GeneralUserController) GetUserProfileHandler(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "details": err.Error()})
		return
	}

	profile, err := g.userUsecase.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		if err == service.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user profile"})
		return
	}
	c.JSON(http.StatusOK, profile)
}

// @Summary Update user profile
// @Description Update phoone or email. ID is taken from token.
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body user.UpdateProfileRequest true "Update profile Request"
// @Success 200 {object} map[string]interfaceP{}
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /user/update-info [post]
func (g *GeneralUserController) UpdateProfileHandler(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "details": err.Error()})
		return
	}

	var req user.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	req.UserID = userID

	profile, err := g.userUsecase.UpdateProfile(c.Request.Context(), req)
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		case service.ErrEmailAlreadyExist:
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		case service.ErrPhoneAlreadyExist:
			c.JSON(http.StatusConflict, gin.H{"error": "Phone number already exists"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile", "details": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"profile": profile,
	})

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
