package verification

import (
	"log/slog"

	"github.com/Barbod-Biometrics/Backened/common/internal/domain/verification"
)

// Service is the use case
type Service struct {
	repo   verification.VerificationRepository // this holds the interface
	logger *slog.Logger
	// mq MessageQueuePublisher -> Would also be an interface
}

// NewService creates a new verification service.
func NewService(r verification.VerificationRepository, l *slog.Logger) *Service {
	return &Service{
		repo:   r,
		logger: l,
	}
}

// RequestVerification is the core business logic
func (s *Service) RequestVerification(userID string) (*verification.Verification, error) {
	s.logger.Info("Use case: Requesting verification", "user_id", userID)

	// 1. Business Logic: Creates the entity
	v := verification.NewVerification(userID)

	// 2. Business logic
	// We are calling the INTERFACE, not a real database
	err := s.repo.Save(v)
	if err != nil {
		s.logger.Error("Use Case: Failed to save verification", "error", err)
		return nil, err
	}

	// 3. Business Logic: publish a message (if we had an MQ interface)
	// s.mq.PublishOCRTask(v.ID, ...)

	s.logger.Info("Use case: Verification saved", "verification_id", v.ID)
	return v, nil
}
