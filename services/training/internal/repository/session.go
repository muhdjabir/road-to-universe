package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/muhdjabir/road-to-universe/services/training/internal/model"
)

var ErrNotFound = errors.New("session not found")

type SessionRepository interface {
	Create(ctx context.Context, session *model.TrainingSession) error
	GetByID(ctx context.Context, id string) (*model.TrainingSession, error)
	ListByUserID(ctx context.Context, userID string) ([]*model.TrainingSession, error)
	Delete(ctx context.Context, id string) error
}

type sessionRepository struct {
	db *sqlx.DB
}

func NewSessionRepository(db *sqlx.DB) SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(ctx context.Context, session *model.TrainingSession) error {
	query := `
		INSERT INTO training_sessions
			(id, user_id, title, description, duration_minutes, throw_count, session_date, created_at, updated_at)
		VALUES
			(:id, :user_id, :title, :description, :duration_minutes, :throw_count, :session_date, :created_at, :updated_at)`
	_, err := r.db.NamedExecContext(ctx, query, session)
	return err
}

func (r *sessionRepository) GetByID(ctx context.Context, id string) (*model.TrainingSession, error) {
	var session model.TrainingSession
	err := r.db.GetContext(ctx, &session, "SELECT * FROM training_sessions WHERE id = $1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *sessionRepository) ListByUserID(ctx context.Context, userID string) ([]*model.TrainingSession, error) {
	var sessions []*model.TrainingSession
	err := r.db.SelectContext(ctx, &sessions,
		"SELECT * FROM training_sessions WHERE user_id = $1 ORDER BY session_date DESC", userID)
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

func (r *sessionRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM training_sessions WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
