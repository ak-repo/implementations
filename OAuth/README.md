# OAuth 2.0 Demo Application

A production-style OAuth 2.0 authentication demo using **Go** (backend) and **React + TypeScript** (frontend).

## Features

- ✅ Google OAuth 2.0 authentication flow
- ✅ JWT-based session management with secure httpOnly cookies
- ✅ Clean architecture (handlers, services, repositories)
- ✅ SQLite database for user storage
- ✅ CSRF protection via state parameter
- ✅ Protected routes (backend & frontend)
- ✅ TypeScript frontend with React Router
- ✅ Modern TailwindCSS UI

## Project Structure

```
OAuth/
├── backend/
│   ├── cmd/
│   │   └── main.go              # Application entry point
│   ├── internal/
│   │   ├── handlers/            # HTTP request handlers
│   │   │   └── auth.go          # OAuth & user handlers
│   │   ├── services/            # Business logic
│   │   │   ├── oauth.go         # Google OAuth service
│   │   │   └── jwt.go           # JWT token service
│   │   ├── repository/          # Data access layer
│   │   │   └── user.go          # User repository (SQLite)
│   │   ├── middleware/          # HTTP middleware
│   │   │   └── auth.go          # JWT validation middleware
│   │   └── models/              # Data models
│   │       └── user.go          # User & GoogleUserInfo structs
│   ├── pkg/
│   │   └── database/
│   │       └── sqlite.go        # Database connection & migrations
│   ├── migrations/
│   │   └── 001_create_users.sql # SQL schema
│   ├── go.mod                   # Go module dependencies
│   └── .env.example             # Environment variables template
│
└── frontend/
    ├── src/
    │   ├── pages/
    │   │   ├── Login.tsx        # Login page with Google button
    │   │   └── Home.tsx         # User dashboard
    │   ├── components/
    │   │   └── ProtectedRoute.tsx # Route guard component
    │   ├── context/
    │   │   └── AuthContext.tsx  # Global auth state management
    │   ├── api/
    │   │   └── auth.ts          # API client functions
    │   ├── types/
    │   │   └── auth.ts          # TypeScript type definitions
    │   ├── App.tsx              # Main app component with router
    │   └── main.tsx             # React entry point
    ├── package.json
    ├── .env.example
    └── README.md
```

## Architecture Overview

### OAuth 2.0 Flow

```
┌─────────────┐                                    ┌─────────────┐
│   Frontend  │                                    │   Backend   │
└──────┬──────┘                                    └──────┬──────┘
       │                                                  │
       │  1. Click "Continue with Google"                  │
       │ ───────────────────────────────────────────────> │
       │                                                  │
       │  2. Redirect to Google consent screen             │
       │ <─────────────────────────────────────────────── │
       │                                                  │
       │  3. User authenticates with Google                │
       │  4. Google redirects to /auth/google/callback     │
       │ ───────────────────────────────────────────────> │
       │                                                  │
       │  5. Exchange code for access token                │
       │  6. Fetch user info from Google                   │
       │  7. Create/find user in database                  │
       │  8. Generate JWT token                            │
       │  9. Set JWT in secure httpOnly cookie             │
       │                                                  │
       │ 10. Redirect to /home                            │
       │ <─────────────────────────────────────────────── │
       │                                                  │
       │ 11. Request /api/me (with cookie)               │
       │ ───────────────────────────────────────────────> │
       │ 12. Validate JWT, return user info               │
       │ <─────────────────────────────────────────────── │
       │                                                  │
```

### Security Features

- **httpOnly Cookies**: JWT tokens are stored in httpOnly cookies (inaccessible to JavaScript, XSS protection)
- **CSRF Protection**: OAuth state parameter prevents CSRF attacks
- **Secure Flag**: Cookies marked Secure (requires HTTPS in production)
- **SameSite=Lax**: Protects against CSRF for cross-site requests
- **Stateless JWT**: No server-side session storage required
- **CORS**: Configured for cross-origin requests with credentials

## Step-by-Step Setup Instructions

### Prerequisites

- Go 1.23+ installed
- Node.js 20+ installed
- Google Cloud Console account (for OAuth credentials)

---

### 1. Set Up Google OAuth Credentials

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing one
3. Navigate to **APIs & Services** > **Credentials**
4. Click **Create Credentials** > **OAuth client ID**
5. Configure the consent screen (External for testing)
6. Add these scopes:
   - `openid`
   - `https://www.googleapis.com/auth/userinfo.email`
   - `https://www.googleapis.com/auth/userinfo.profile`
7. Application type: **Web application**
8. Authorized JavaScript origins: `http://localhost:3050`
9. Authorized redirect URIs: `http://localhost:8090/auth/google/callback`
10. Save the **Client ID** and **Client Secret**

---

### 2. Backend Setup

```bash
# Navigate to backend directory
cd backend

# Download dependencies
go mod tidy

# Create environment file
cp .env.example .env

# Edit .env with your credentials
# GOOGLE_CLIENT_ID=your_client_id
# GOOGLE_CLIENT_SECRET=your_client_secret
# JWT_SECRET=your_random_secret (use: openssl rand -base64 32)

# Run the server
go run cmd/main.go

# Server will start on port 8090
```

---

### 3. Frontend Setup

```bash
# Navigate to frontend directory
cd frontend

# Install dependencies
npm install

# Create environment file
cp .env.example .env

# Edit .env (if needed)
# VITE_API_URL=http://localhost:8090

# Start development server
npm run dev

# Frontend will start on port 3050
```

---

### 4. Access the Application

1. Open browser: `http://localhost:3050`
2. Click "Continue with Google"
3. Authenticate with your Google account
4. You'll be redirected to the Home page with your profile info

---

## API Endpoints

| Endpoint | Method | Auth Required | Description |
|----------|--------|---------------|-------------|
| `/auth/google/login` | GET | No | Initiate OAuth flow, redirect to Google |
| `/auth/google/callback` | GET | No | Handle OAuth callback, issue JWT |
| `/auth/logout` | POST | No | Clear JWT cookie |
| `/api/me` | GET | Yes | Get current user profile |
| `/health` | GET | No | Health check endpoint |

---

## Environment Variables

### Backend (.env)

```bash
# Google OAuth
GOOGLE_CLIENT_ID=your_client_id
GOOGLE_CLIENT_SECRET=your_client_secret
GOOGLE_REDIRECT_URL=http://localhost:8090/auth/google/callback

# JWT
JWT_SECRET=your_super_secret_key_min_32_chars

# App Config
PORT=8090
FRONTEND_URL=http://localhost:3050
DB_PATH=./users.db
```

### Frontend (.env)

```bash
VITE_API_URL=http://localhost:8090
```

---

## Database Schema

### Users Table

```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    picture TEXT,
    created_at DATETIME NOT NULL
);
```

- Auto-created on first run
- Users are automatically inserted on first OAuth login
- Email is unique to prevent duplicates

---

## Security Considerations

### For Development

- Backend runs on HTTP (localhost)
- Secure cookie flag should be set to `false` for local development
- CORS allows `localhost:3050`

### For Production

1. **Use HTTPS**: Set `Secure: true` for cookies
2. **Strong JWT Secret**: Generate with `openssl rand -base64 32`
3. **Environment Variables**: Never commit `.env` files
4. **Rate Limiting**: Add rate limiting for OAuth endpoints
5. **Token Rotation**: Implement refresh tokens for enhanced security
6. **Content Security Policy**: Add CSP headers
7. **Audit Logging**: Log authentication events

---

## Common Issues

### CORS Errors

Make sure `FRONTEND_URL` in backend `.env` matches your frontend URL exactly (including protocol and port).

### Cookie Not Being Set

For local development, cookies work with HTTP. In production, you need HTTPS with `Secure: true`.

### "Unauthorized" on /api/me

- Check that the JWT cookie is being sent (check browser DevTools > Application > Cookies)
- Verify `credentials: 'include'` is set in fetch requests
- Check backend logs for token validation errors

### OAuth Redirect Mismatch

- Ensure redirect URI in Google Console matches exactly: `http://localhost:8090/auth/google/callback`
- No trailing slash, correct protocol

---

## Extending the Application

### Add Refresh Tokens

1. Uncomment the `refresh_tokens` table in `migrations/001_create_users.sql`
2. Create a `/auth/refresh` endpoint
3. Issue short-lived access tokens (15 min) and long-lived refresh tokens (7 days)
4. Store refresh token hash in database, not the token itself

### Add User Roles

1. Add `role` column to users table
2. Create RBAC middleware
3. Add role checks to protected routes

### Add Session Management

1. Track active sessions in database
2. Allow users to view/revoke sessions
3. Implement session expiration notifications

---

## License

MIT License - Feel free to use for your projects!

## Resources

- [OAuth 2.0 Spec](https://oauth.net/2/)
- [Google OAuth Documentation](https://developers.google.com/identity/protocols/oauth2)
- [JWT Best Practices](https://tools.ietf.org/html/bcp195)
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
