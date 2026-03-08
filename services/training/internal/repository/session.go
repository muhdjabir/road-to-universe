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
	Update(ctx context.Context, session *model.TrainingSession) error
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

	if err = upsertThrowingStats(ctx, tx, session.ID, session.Throwing); err != nil {
		return err
	}
	if err = upsertConditioningStats(ctx, tx, session.ID, session.Conditioning); err != nil {
		return err
	}

	return tx.Commit()
}

// GetByID fetches an active session and loads its stat records.
func (r *sessionRepository) GetByID(ctx context.Context, id string) (*model.TrainingSession, error) {
	var session model.TrainingSession
	err := r.db.GetContext(ctx, &session,
		"SELECT * FROM training_sessions WHERE id = $1 AND deleted_at IS NULL", id)
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

// ListByUserID returns all active sessions for a user (without nested stats for list performance).
func (r *sessionRepository) ListByUserID(ctx context.Context, userID string) ([]*model.TrainingSession, error) {
	var sessions []*model.TrainingSession
	err := r.db.SelectContext(ctx, &sessions,
		"SELECT * FROM training_sessions WHERE user_id = $1 AND deleted_at IS NULL ORDER BY session_date DESC", userID)
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

// Update patches the session core fields and upserts any provided stat records.
func (r *sessionRepository) Update(ctx context.Context, session *model.TrainingSession) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	const updateQuery = `
		UPDATE training_sessions SET
			session_type     = :session_type,
			duration_minutes = :duration_minutes,
			intensity        = :intensity,
			location         = :location,
			weather          = :weather,
			notes            = :notes,
			session_date     = :session_date,
			updated_at       = :updated_at
		WHERE id = :id AND deleted_at IS NULL`

	result, err := tx.NamedExecContext(ctx, updateQuery, session)
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

	if err = upsertThrowingStats(ctx, tx, session.ID, session.Throwing); err != nil {
		return err
	}
	if err = upsertConditioningStats(ctx, tx, session.ID, session.Conditioning); err != nil {
		return err
	}

	return tx.Commit()
}

// Delete soft-deletes the session by setting deleted_at.
func (r *sessionRepository) Delete(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE training_sessions SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL", id)
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

// ── Helpers ────────────────────────────────────────────────────────────────────

type txExecer interface {
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
}

func upsertThrowingStats(ctx context.Context, tx txExecer, sessionID string, stats *model.ThrowingStats) error {
	if stats == nil {
		return nil
	}
	if stats.ID == "" {
		stats.ID = uuid.New().String()
	}
	stats.SessionID = sessionID

	const q = `
		INSERT INTO throwing_stats
			(id, session_id, backhand_reps, forehand_reps, hammer_reps, scoober_reps, break_throws, hucks, turnovers)
		VALUES
			(:id, :session_id, :backhand_reps, :forehand_reps, :hammer_reps, :scoober_reps, :break_throws, :hucks, :turnovers)
		ON CONFLICT (session_id) DO UPDATE SET
			backhand_reps = EXCLUDED.backhand_reps,
			forehand_reps = EXCLUDED.forehand_reps,
			hammer_reps   = EXCLUDED.hammer_reps,
			scoober_reps  = EXCLUDED.scoober_reps,
			break_throws  = EXCLUDED.break_throws,
			hucks         = EXCLUDED.hucks,
			turnovers     = EXCLUDED.turnovers`

	_, err := tx.NamedExecContext(ctx, q, stats)
	return err
}

func upsertConditioningStats(ctx context.Context, tx txExecer, sessionID string, stats *model.ConditioningStats) error {
	if stats == nil {
		return nil
	}
	if stats.ID == "" {
		stats.ID = uuid.New().String()
	}
	stats.SessionID = sessionID

	const q = `
		INSERT INTO conditioning_stats
			(id, session_id, sprints, distance_km, max_speed_kmh, heart_rate_avg, heart_rate_max)
		VALUES
			(:id, :session_id, :sprints, :distance_km, :max_speed_kmh, :heart_rate_avg, :heart_rate_max)
		ON CONFLICT (session_id) DO UPDATE SET
			sprints        = EXCLUDED.sprints,
			distance_km    = EXCLUDED.distance_km,
			max_speed_kmh  = EXCLUDED.max_speed_kmh,
			heart_rate_avg = EXCLUDED.heart_rate_avg,
			heart_rate_max = EXCLUDED.heart_rate_max`

	_, err := tx.NamedExecContext(ctx, q, stats)
	return err
}
