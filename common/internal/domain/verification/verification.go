package verification

import "time"

// Verification is the core business entity.
// It has no database tags or JSON tags. It's pure
type Verification struct {
	ID        string
	UserID    string
	Status    string
	CreatedAt time.Time
}

// NewVerification is a "factory" to create a valid verification
func NewVerification(userID string) *Verification {
	return &Verification{
		ID:        "Some-unique-id",
		UserID:    userID,
		Status:    "pendidng",
		CreatedAt: time.Now(),
	}
}

// VerificationRepository is the interface (the contract)
// The use case depends on this, not on a real database
type VerificationRepository interface {
	Save(v *Verification) error
	GetByID(id string) (*Verification, error)
}
