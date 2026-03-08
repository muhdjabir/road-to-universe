package model

import "time"

// ── Core session ───────────────────────────────────────────────────────────────

type TrainingSession struct {
	ID           string     `db:"id"               json:"id"`
	UserID       string     `db:"user_id"          json:"user_id"`
	SessionType  string     `db:"session_type"     json:"session_type"`
	Duration     int        `db:"duration_minutes" json:"duration_minutes"`
	Intensity    string     `db:"intensity"        json:"intensity"`
	Location     string     `db:"location"         json:"location,omitempty"`
	Weather      string     `db:"weather"          json:"weather,omitempty"`
	Notes        string     `db:"notes"            json:"notes,omitempty"`
	SessionDate  time.Time  `db:"session_date"     json:"session_date"`
	CreatedAt    time.Time  `db:"created_at"       json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"       json:"updated_at"`
	DeletedAt    *time.Time `db:"deleted_at"       json:"-"`

	// Populated on GET — not stored in training_sessions table
	Throwing     *ThrowingStats     `db:"-" json:"throwing,omitempty"`
	Conditioning *ConditioningStats `db:"-" json:"conditioning,omitempty"`
}

// ── Throwing stats ─────────────────────────────────────────────────────────────

type ThrowingStats struct {
	ID           string `db:"id"            json:"-"`
	SessionID    string `db:"session_id"    json:"-"`
	BackhandReps int    `db:"backhand_reps" json:"backhand_reps"`
	ForehandReps int    `db:"forehand_reps" json:"forehand_reps"`
	HammerReps   int    `db:"hammer_reps"   json:"hammer_reps"`
	ScooberReps  int    `db:"scoober_reps"  json:"scoober_reps"`
	BreakThrows  int    `db:"break_throws"  json:"break_throws"`
	Hucks        int    `db:"hucks"         json:"hucks"`
	Turnovers    int    `db:"turnovers"     json:"turnovers"`
}

// ── Conditioning stats ─────────────────────────────────────────────────────────

type ConditioningStats struct {
	ID            string  `db:"id"              json:"-"`
	SessionID     string  `db:"session_id"      json:"-"`
	Sprints       int     `db:"sprints"         json:"sprints"`
	DistanceKm    float64 `db:"distance_km"     json:"distance_km,omitempty"`
	MaxSpeedKmh   float64 `db:"max_speed_kmh"   json:"max_speed_kmh,omitempty"`
	HeartRateAvg  int     `db:"heart_rate_avg"  json:"heart_rate_avg,omitempty"`
	HeartRateMax  int     `db:"heart_rate_max"  json:"heart_rate_max,omitempty"`
}

// ── Request / response ─────────────────────────────────────────────────────────

type CreateSessionRequest struct {
	SessionType  string             `json:"session_type"     binding:"required,oneof=team_training throwing gym conditioning scrimmage other"`
	Duration     int                `json:"duration_minutes" binding:"required,min=1"`
	Intensity    string             `json:"intensity"        binding:"required,oneof=low medium high"`
	Location     string             `json:"location"`
	Weather      string             `json:"weather"`
	Notes        string             `json:"notes"`
	SessionDate  time.Time          `json:"session_date"     binding:"required"`
	Throwing     *ThrowingStats     `json:"throwing"`
	Conditioning *ConditioningStats `json:"conditioning"`
}

type UpdateSessionRequest struct {
	SessionType  string             `json:"session_type"     binding:"omitempty,oneof=team_training throwing gym conditioning scrimmage other"`
	Duration     int                `json:"duration_minutes" binding:"omitempty,min=1"`
	Intensity    string             `json:"intensity"        binding:"omitempty,oneof=low medium high"`
	Location     string             `json:"location"`
	Weather      string             `json:"weather"`
	Notes        string             `json:"notes"`
	SessionDate  *time.Time         `json:"session_date"`
	Throwing     *ThrowingStats     `json:"throwing"`
	Conditioning *ConditioningStats `json:"conditioning"`
}
