package v1

import (
	"net/http"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/user"
	"github.com/gin-gonic/gin"
)

type Route struct {
	authController *user.GeneralUserController
}

func NewRouter(
	authController *user.GeneralUserController,
) *Route {
	return &Route{
		authController: authController,
	}
}

func (r *Route) RegisterRoutes() http.Handler {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	auth := router.Group("/auth")
	{
		auth.POST("/request-otp", r.authController.RequestOTPHandler)
		auth.POST("/verify-otp", r.authController.VerifyOTPHandler)
	}

	return router
}
