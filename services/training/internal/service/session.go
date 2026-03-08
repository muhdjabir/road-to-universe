package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/muhdjabir/road-to-universe/services/training/internal/model"
	"github.com/muhdjabir/road-to-universe/services/training/internal/repository"
	"go.uber.org/zap"
)

type SessionService interface {
	CreateSession(ctx context.Context, userID string, req *model.CreateSessionRequest) (*model.TrainingSession, error)
	GetSession(ctx context.Context, id string) (*model.TrainingSession, error)
	ListSessions(ctx context.Context, userID string) ([]*model.TrainingSession, error)
	UpdateSession(ctx context.Context, id string, req *model.UpdateSessionRequest) (*model.TrainingSession, error)
	DeleteSession(ctx context.Context, id string) error
}

type sessionService struct {
	repo repository.SessionRepository
	log  *zap.Logger
}

func NewSessionService(repo repository.SessionRepository, log *zap.Logger) SessionService {
	return &sessionService{repo: repo, log: log}
}

func (s *sessionService) CreateSession(ctx context.Context, userID string, req *model.CreateSessionRequest) (*model.TrainingSession, error) {
	now := time.Now().UTC()
	session := &model.TrainingSession{
		ID:           uuid.New().String(),
		UserID:       userID,
		SessionType:  req.SessionType,
		Duration:     req.Duration,
		Intensity:    req.Intensity,
		Location:     req.Location,
		Weather:      req.Weather,
		Notes:        req.Notes,
		SessionDate:  req.SessionDate,
		CreatedAt:    now,
		UpdatedAt:    now,
		Throwing:     req.Throwing,
		Conditioning: req.Conditioning,
	}

	if err := s.repo.Create(ctx, session); err != nil {
		return nil, err
	}

	s.log.Info("session created",
		zap.String("session_id", session.ID),
		zap.String("user_id", userID),
		zap.String("session_type", session.SessionType),
		zap.Int("duration_minutes", session.Duration),
		zap.String("intensity", session.Intensity),
	)

	return session, nil
}

func (s *sessionService) GetSession(ctx context.Context, id string) (*model.TrainingSession, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *sessionService) ListSessions(ctx context.Context, userID string) ([]*model.TrainingSession, error) {
	return s.repo.ListByUserID(ctx, userID)
}

func (s *sessionService) UpdateSession(ctx context.Context, id string, req *model.UpdateSessionRequest) (*model.TrainingSession, error) {
	session, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.SessionType != "" {
		session.SessionType = req.SessionType
	}
	if req.Duration != 0 {
		session.Duration = req.Duration
	}
	if req.Intensity != "" {
		session.Intensity = req.Intensity
	}
	if req.SessionDate != nil {
		session.SessionDate = *req.SessionDate
	}
	session.Location = req.Location
	session.Weather = req.Weather
	session.Notes = req.Notes
	session.UpdatedAt = time.Now().UTC()

	if req.Throwing != nil {
		session.Throwing = req.Throwing
	}
	if req.Conditioning != nil {
		session.Conditioning = req.Conditioning
	}

	if err := s.repo.Update(ctx, session); err != nil {
		return nil, err
	}

	s.log.Info("session updated",
		zap.String("session_id", session.ID),
		zap.String("user_id", session.UserID),
		zap.String("session_type", session.SessionType),
	)

	return session, nil
}

func (s *sessionService) DeleteSession(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	s.log.Info("session deleted", zap.String("session_id", id))

	return nil
}
