### Auth Service — `/services/auth-service`

Responsibility: bridge between Auth0 and the platform's user profile data.
Auth0 owns credentials. This service owns profile fields.

**Endpoints:**
```
POST   /api/auth/callback  — exchange Auth0 code for token, set HttpOnly cookie,
                             provision profile row if this is the user's first login
GET    /api/auth/me        — return the authenticated user's profile
PUT    /api/auth/me        — update profile (display_name, position, team_name)
DELETE /api/auth/me        — delete account (removes Auth0 user + local profile row)
POST   /api/auth/logout    — clear the access_token cookie
```

**Database: `users` table**
```sql
users (
    id           UUID PRIMARY KEY,         -- same value as Auth0 sub
    auth0_id     VARCHAR(128) UNIQUE NOT NULL,
    email        VARCHAR(255) UNIQUE NOT NULL,
    display_name VARCHAR(100),
    position     VARCHAR(20),              -- 'handler' | 'cutter' | 'hybrid'
    team_name    VARCHAR(100),
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    updated_at   TIMESTAMPTZ DEFAULT NOW()
)
```

**No event publishing from this service.**
