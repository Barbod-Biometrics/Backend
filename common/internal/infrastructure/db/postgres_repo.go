package db

import (
	"database/sql"

	"github.com/Barbod-Biometrics/Backened/common/internal/domain/verification"
)

// PostgresRepo is the "porter" who knows how to talk to Postgres
type PostgresRepo struct {
	db *sql.DB
}

// NewPostgresRepo creates a new repository
func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

// Save implements the VerificationRepository interafce.
// The name and signature MUST match the interface
func (r *PostgresRepo) Save(v *verification.Verification) error {
	query := `INSERT INTO verifications (id, user_id, status, created_at)
	          VALUES ($1, $2, $3, $4)`

	_, err := r.db.Exec(query, v.ID, v.UserID, v.Status, v.CreatedAt)
	return err
}

// GetByID implements the other part of the interface
func (r *PostgresRepo) GetByID(id string) (*verification.Verification, error) {
	// ... code to SELECT from database ...
	return nil, nil
}
