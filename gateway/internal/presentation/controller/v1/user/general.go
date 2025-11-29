package user

import (
	"encoding/json"
	"net/http"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/auth"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/service"
)

type GeneralUserController struct {
	authUsecase *service.AuthUsecase
}

func NewAuthController(authUsecase *service.AuthUsecase) *GeneralUserController {
	return &GeneralUserController{authUsecase: authUsecase}
}

func (g *GeneralUserController) RequestOTPHandler(w http.ResponseWriter, r *http.Request) {
	var req auth.RequestOTPRequest

	// decoding request boy
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// calling usecase
	if err := g.authUsecase.RequestOTP(r.Context(), req); err != nil {
		http.Error(w, "failed to send OTP", http.StatusInternalServerError)
		return
	}

	// sending a simple success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "OTP sent successfully"})

}

func (g *GeneralUserController) VerifyOTPHandler(w http.ResponseWriter, r *http.Request) {
	var req auth.VerifyOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
	}

	response, err := g.authUsecase.VerifyOTP(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
