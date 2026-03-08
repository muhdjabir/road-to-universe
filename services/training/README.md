# Training Service

Handles logging and retrieval of Ultimate Frisbee training sessions. This is the core service of the platform.

---

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check |
| `POST` | `/api/v1/training` | Log a new training session |
| `GET` | `/api/v1/training` | List all sessions for a user |
| `GET` | `/api/v1/training/:id` | Get a single session with full stats |
| `DELETE` | `/api/v1/training/:id` | Delete a session |

> **Note:** `X-User-ID` header is used as a temporary user identifier until the Auth Service is wired up in Phase 2.

---

## Session Data

Every session starts with a core record, with optional throwing and conditioning stats attached.

### Core Session

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `session_type` | string | yes | `team_training`, `throwing`, `gym`, `conditioning`, `scrimmage`, `other` |
| `duration_minutes` | int | yes | Length of the session |
| `intensity` | string | yes | `low`, `medium`, `high` |
| `session_date` | date | yes | Date the session took place |
| `location` | string | no | e.g. turf field, grass, gym |
| `weather` | string | no | e.g. windy, sunny, raining |
| `notes` | string | no | Free-text notes |

### Throwing Stats

Tracks throwing volume and quality. Attached when the session involved disc work.

| Field | Type | Description |
|-------|------|-------------|
| `backhand_reps` | int | Backhand throw repetitions |
| `forehand_reps` | int | Forehand throw repetitions |
| `hammer_reps` | int | Hammer throw repetitions |
| `scoober_reps` | int | Scoober throw repetitions |
| `break_throws` | int | Break-mark throws attempted |
| `hucks` | int | Long throws attempted |
| `turnovers` | int | Throws that resulted in a turnover |

### Conditioning Stats

Tracks athletic load. Attach when the session involved physical conditioning.

| Field | Type | Description |
|-------|------|-------------|
| `sprints` | int | Number of sprints completed |
| `distance_km` | float | Total distance covered |
| `max_speed_kmh` | float | Peak speed recorded |
| `heart_rate_avg` | int | Average heart rate (bpm) |
| `heart_rate_max` | int | Max heart rate (bpm) |

---

## Example Requests

### Throwing session

```json
POST /api/v1/training

{
  "session_type": "throwing",
  "duration_minutes": 60,
  "intensity": "medium",
  "session_date": "2026-03-08T00:00:00Z",
  "weather": "windy",
  "location": "grass field",
  "throwing": {
    "backhand_reps": 120,
    "forehand_reps": 90,
    "break_throws": 30,
    "hucks": 12,
    "turnovers": 5
  }
}
```

### Team training session

```json
POST /api/v1/training

{
  "session_type": "team_training",
  "duration_minutes": 90,
  "intensity": "high",
  "session_date": "2026-03-08T00:00:00Z",
  "location": "turf",
  "notes": "focused on defensive positioning and break throws",
  "throwing": {
    "backhand_reps": 80,
    "forehand_reps": 60,
    "break_throws": 25,
    "turnovers": 8
  },
  "conditioning": {
    "sprints": 20,
    "distance_km": 5.2,
    "heart_rate_avg": 155,
    "heart_rate_max": 182
  }
}
```

### Conditioning-only session

```json
POST /api/v1/training

{
  "session_type": "conditioning",
  "duration_minutes": 45,
  "intensity": "high",
  "session_date": "2026-03-08T00:00:00Z",
  "conditioning": {
    "sprints": 15,
    "distance_km": 4.2,
    "max_speed_kmh": 27.5,
    "heart_rate_avg": 162,
    "heart_rate_max": 188
  }
}
```

---

## Database Schema

Three tables — the stat tables cascade delete with the session.

```
training_sessions
    └── throwing_stats      (1:1, optional)
    └── conditioning_stats  (1:1, optional)
```

Planned future tables (separate migrations):
- `offensive_stats` — cuts, assists, drops
- `defensive_stats` — blocks, layout blocks, forced turnovers
- `gym_exercises` — exercise, sets, reps, weight
- `pull_stats` — pulls, average pull distance

---

## Running Locally

```bash
# From repo root
make up              # start postgres + training service
make migrate-training  # apply schema
```

Service runs on `http://localhost:8081`.
