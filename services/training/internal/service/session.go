package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/muhdjabir/road-to-universe/services/training/internal/model"
	"github.com/muhdjabir/road-to-universe/services/training/internal/repository"
)

type SessionService interface {
	CreateSession(ctx context.Context, userID string, req *model.CreateSessionRequest) (*model.TrainingSession, error)
	GetSession(ctx context.Context, id string) (*model.TrainingSession, error)
	ListSessions(ctx context.Context, userID string) ([]*model.TrainingSession, error)
	DeleteSession(ctx context.Context, id string) error
}

type sessionService struct {
	repo repository.SessionRepository
}

func NewSessionService(repo repository.SessionRepository) SessionService {
	return &sessionService{repo: repo}
}

func (s *sessionService) CreateSession(ctx context.Context, userID string, req *model.CreateSessionRequest) (*model.TrainingSession, error) {
	now := time.Now().UTC()
	session := &model.TrainingSession{
		ID:          uuid.New().String(),
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Duration:    req.Duration,
		ThrowCount:  req.ThrowCount,
		SessionDate: req.SessionDate,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.Create(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *sessionService) GetSession(ctx context.Context, id string) (*model.TrainingSession, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *sessionService) ListSessions(ctx context.Context, userID string) ([]*model.TrainingSession, error) {
	return s.repo.ListByUserID(ctx, userID)
}

func (s *sessionService) DeleteSession(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
