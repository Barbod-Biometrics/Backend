package http

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Barbod-Biometrics/Backened/common/internal/domain/verification"
	// The handler is NOT allowed to import the domain, only the use case
)

// We need an interface for our use case to keep the handler "clean"
// This allows us to test the handler by mocking the service.
type VerificationService interface {
	RequestVerification(userID string) (*verification.Verification, error)
}

type VerificationHandler struct {
	service VerificationService
	logger  *slog.Logger
}

func NewVerificationHandler(s VerificationService, l *slog.Logger) *VerificationHandler {
	return &VerificationHandler{
		service: s,
		logger:  l,
	}
}

// DTO = Data Transfer Object. Just for web requests.
type requestDTO struct {
	UserID string `json:"user_id"`
}

func (h *VerificationHandler) HandleRequestVerification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	var dto requestDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	v, err := h.service.RequestVerification(dto.UserID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(v)
}
