package v1

import (
	"net/http"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/presentation/controller/v1/user"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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
	router := chi.NewRouter()

	router.Use(middleware.Logger) // I don't know the status and how the implemented logger works and even if this logger would effect that or not but this should be analysed.
	router.Use(middleware.Recoverer)
	router.Use(middleware.AllowContentType("application/json"))

	router.Route("/auth", func(router chi.Router) {
		router.Post("/request-otp", r.authController.RequestOTPHandler)
		router.Post("/verify-otp", r.authController.VerifyOTPHandler)
	})

	return router
}
