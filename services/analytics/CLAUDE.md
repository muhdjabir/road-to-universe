### Analytics Service — `/services/analytics-service`

Responsibility: consume training events and compute aggregate stats.
This service has its own database — never query it from other services.

**Endpoints:**
```
GET /api/analytics/weekly  — throwing volume for the past 7 days
GET /api/analytics/stats   — totals, streaks, throw type breakdown
```

**Consumes:**
- `training.session.created` → add to aggregates
- `training.session.deleted` → reverse aggregates

**Must be idempotent.** Use `event_id` as a deduplication key — processing the same
event twice must not corrupt aggregate data.