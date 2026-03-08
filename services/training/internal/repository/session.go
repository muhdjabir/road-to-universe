package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
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

// Create inserts the session and any optional stat records in a single transaction.
func (r *sessionRepository) Create(ctx context.Context, session *model.TrainingSession) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	const sessionQuery = `
		INSERT INTO training_sessions
			(id, user_id, session_type, duration_minutes, intensity, location, weather, notes, session_date, created_at, updated_at)
		VALUES
			(:id, :user_id, :session_type, :duration_minutes, :intensity, :location, :weather, :notes, :session_date, :created_at, :updated_at)`

	if _, err = tx.NamedExecContext(ctx, sessionQuery, session); err != nil {
		return err
	}

	if session.Throwing != nil {
		session.Throwing.ID = uuid.New().String()
		session.Throwing.SessionID = session.ID
		const throwingQuery = `
			INSERT INTO throwing_stats
				(id, session_id, backhand_reps, forehand_reps, hammer_reps, scoober_reps, break_throws, hucks, turnovers)
			VALUES
				(:id, :session_id, :backhand_reps, :forehand_reps, :hammer_reps, :scoober_reps, :break_throws, :hucks, :turnovers)`
		if _, err = tx.NamedExecContext(ctx, throwingQuery, session.Throwing); err != nil {
			return err
		}
	}

	if session.Conditioning != nil {
		session.Conditioning.ID = uuid.New().String()
		session.Conditioning.SessionID = session.ID
		const condQuery = `
			INSERT INTO conditioning_stats
				(id, session_id, sprints, distance_km, max_speed_kmh, heart_rate_avg, heart_rate_max)
			VALUES
				(:id, :session_id, :sprints, :distance_km, :max_speed_kmh, :heart_rate_avg, :heart_rate_max)`
		if _, err = tx.NamedExecContext(ctx, condQuery, session.Conditioning); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetByID fetches the session then loads throwing/conditioning stats separately.
func (r *sessionRepository) GetByID(ctx context.Context, id string) (*model.TrainingSession, error) {
	var session model.TrainingSession
	err := r.db.GetContext(ctx, &session, "SELECT * FROM training_sessions WHERE id = $1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	var throwing model.ThrowingStats
	if err = r.db.GetContext(ctx, &throwing, "SELECT * FROM throwing_stats WHERE session_id = $1", id); err == nil {
		session.Throwing = &throwing
	}

	var conditioning model.ConditioningStats
	if err = r.db.GetContext(ctx, &conditioning, "SELECT * FROM conditioning_stats WHERE session_id = $1", id); err == nil {
		session.Conditioning = &conditioning
	}

	return &session, nil
}

// ListByUserID returns all sessions for a user (without nested stats for list performance).
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
