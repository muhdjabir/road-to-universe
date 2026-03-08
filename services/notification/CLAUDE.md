### Notification Worker — `/services/notification-worker`

Responsibility: send timely reminders and summaries. No HTTP endpoints.

**Consumes:**
- `training.session.created` → resets the inactivity timer for that user

**Logic:**
- User has not trained in 3+ days → send training reminder
- Weekly trigger → send training summary
