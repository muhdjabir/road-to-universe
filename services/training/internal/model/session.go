package model

import "time"

type TrainingSession struct {
	ID          string    `db:"id"               json:"id"`
	UserID      string    `db:"user_id"          json:"user_id"`
	Title       string    `db:"title"            json:"title"`
	Description string    `db:"description"      json:"description"`
	Duration    int       `db:"duration_minutes" json:"duration_minutes"`
	ThrowCount  int       `db:"throw_count"      json:"throw_count"`
	SessionDate time.Time `db:"session_date"     json:"session_date"`
	CreatedAt   time.Time `db:"created_at"       json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"       json:"updated_at"`
}

type CreateSessionRequest struct {
	Title       string    `json:"title"            binding:"required"`
	Description string    `json:"description"`
	Duration    int       `json:"duration_minutes" binding:"required,min=1"`
	ThrowCount  int       `json:"throw_count"      binding:"min=0"`
	SessionDate time.Time `json:"session_date"     binding:"required"`
}
