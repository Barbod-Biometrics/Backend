package session

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
	"github.com/google/uuid"
)

type SessionStore interface {
	Create(s *entity.Session) error
	Get(id string) (*entity.Session, bool)
	Update(s *entity.Session) error
	Delete(id string) error
	Close() error
}

type SessionManager struct {
	store SessionStore
	ttl   time.Duration
	repo  repository.SessionRepository
}

func NewSessionManager(store SessionStore, ttl time.Duration, repo repository.SessionRepository) *SessionManager {
	return &SessionManager{store: store, ttl: ttl, repo: repo}
}

func (m *SessionManager) StartSession(profileID uint64, clientIP string) (*entity.Session, error) {
	if m.store == nil {
		return nil, errors.New("no session store configured")
	}

	now := time.Now()
	s := &entity.Session{
		ID:        uuid.NewString(),
		ProfileID: profileID,
		ClientIP:  clientIP,
		State:     entity.StateCreated,
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(m.ttl),
		Metadata:  map[string]interface{}{},
	}

	if err := m.store.Create(s); err != nil {
		return nil, err
	}
	if m.repo != nil {
		if err := m.persistSession(context.Background(), s); err != nil {
			// log? for now ignore persistence error
		}
	}
	return s, nil
}

func (m *SessionManager) GetSession(id string) (*entity.Session, error) {
	if m.store == nil {
		return nil, errors.New("no session store configured")
	}
	s, ok := m.store.Get(id)
	if !ok {
		return nil, errors.New("session not found")
	}

	if time.Now().After(s.ExpiresAt) {
		s.State = entity.StateExpired
		_ = m.store.Update(s)
		return s, errors.New("session expired")
	}
	return s, nil
}

func (m *SessionManager) UpdateSession(s *entity.Session) error {
	if m.store == nil {
		return errors.New("no session store configured")
	}
	s.UpdatedAt = time.Now()

	s.ExpiresAt = time.Now().Add(m.ttl)
	return m.store.Update(s)
}

func (m *SessionManager) CompleteSession(s *entity.Session) error {
	s.Lock()
	s.State = entity.StateCompleted
	s.UpdatedAt = time.Now()
	s.ExpiresAt = time.Now().Add(time.Minute) // short grace
	s.Unlock()
	if err := m.store.Update(s); err != nil {
		return err
	}
	if m.repo != nil {
		return m.persistSession(context.Background(), s)
	}
	return nil
}

func (m *SessionManager) CancelSession(s *entity.Session) error {
	s.Lock()
	s.State = entity.StateCancelled
	s.UpdatedAt = time.Now()
	s.ExpiresAt = time.Now().Add(time.Minute)
	s.Unlock()
	if err := m.store.Update(s); err != nil {
		return err
	}
	if m.repo != nil {
		return m.persistSession(context.Background(), s)
	}
	return nil
}

func (m *SessionManager) persistSession(ctx context.Context, s *entity.Session) error {
	s.Lock()
	defer s.Unlock()

	dbEnt := &entity.Session{
		ID:               s.ID,
		ProfileID:        s.ProfileID,
		WorkflowConfigID: s.WorkflowConfigID,
		ClientIP:         s.ClientIP,
		State:            s.State,
		CreatedAt:        s.CreatedAt,
		UpdatedAt:        s.UpdatedAt,
		ExpiresAt:        s.ExpiresAt,
	}

	if s.OCRResult != nil {
		if b, err := json.Marshal(s.OCRResult); err == nil {
			dbEnt.OCRResultDB = b
		}
	}
	if s.FaceResult != nil {
		if b, err := json.Marshal(s.FaceResult); err == nil {
			dbEnt.FaceResultDB = b
		}
	}
	if s.Errors != nil {
		if b, err := json.Marshal(s.Errors); err == nil {
			dbEnt.ErrorsDB = b
		}
	}
	if s.Metadata != nil {
		if b, err := json.Marshal(s.Metadata); err == nil {
			dbEnt.MetadataDB = b
		}
	}

	if _, err := m.repo.GetByID(ctx, s.ID); err != nil {
		return m.repo.Save(ctx, dbEnt)
	}
	return m.repo.Update(ctx, dbEnt)
}
