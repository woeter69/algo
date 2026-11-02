# Changelog

All notable changes to AlumniGo will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

---

## [Unreleased] — Go Backend Migration (`feat/go-backend-migration`)

### Added

- **`app/src/render.go`** — Pongo2 preprocessing loader and custom template filters; handles Jinja2 → Pongo2 compatibility (newlines, `url_for`, Python dict literals, inline ternaries); template globals (current user, session role, etc.)
- **`app/src/middleware.go`** — `authMiddleware`, `adminMiddleware`, session helpers (`getSessionUser`, `getSessionRole`, `setSession`, `clearSession`)
- **`app/src/utils.go`** — `HashPassword`, `CheckPasswordHash` (bcrypt + Werkzeug PBKDF2/scrypt/plaintext fallback), `GenerateDefaultAvatar`, `FormatISTTime`, misc helpers
- **`app/src/email.go`** — Full email system: `sendEmailAsync` (goroutine-based SMTP), `renderEmailWrapper` (HTML email layout), and high-level send functions:
  - `SendPasswordResetEmail`
  - `SendInquiryConfirmationEmail`
  - `SendInquiryResolvedEmail`
  - `BuildPasswordResetEmail`, `BuildInquiryConfirmationEmail`, `BuildInquiryResolvedEmail`, `BuildInquiryReceivedEmail` — testable builder functions that return `(subject, plainText, htmlBody string)`
- **`app/src/handlers_core.go`** — Routes: `/`, `/about`, `/contact` (GET+POST), `/thanks`, `/recommendations`
- **`app/src/handlers_auth.go`** — Routes: `/register`, `/login`, `/logout`, `/check_email`, `/verify/{token}`, `/confirmation`, `/forgot_password`; Google OAuth flow
- **`app/src/handlers_dashboard.go`** — Routes: `/dashboard`, `/user_dashboard`, `/admin_dashboard`, `/admin/contact/{id}/resolve`, `/limited_dashboard`, `/verification_request`, `/api/handle_verification_request`
- **`app/src/handlers_profile.go`** — Routes: `/profile`, `/profile/{username}`, `/api/profile/update`, `/api/profile/experience` (POST+DELETE), `/api/profile/education` (POST+DELETE), `/complete_profile`, `/interests`
- **`app/src/handlers_settings.go`** — Full settings routes + `/api/update_preferences`
- **`app/src/handlers_connections.go`** — Routes: `/connect`, `/requests`, `/api/send_connection_request`, `/api/respond_connection_request`, `/api/get_connection_requests`, `/api/cancel_connection_request`, `/api/search_alumni`
- **`app/src/handlers_chat.go`** — Routes: `/chat`, `/chat/{username}`, `/api/online_status`, `/api/upload_chat_file`, `/api/search_users`
- **`app/src/handlers_channels.go`** — Routes: `/api/channels`, `/api/channels` (POST), `/api/channels/{id}/messages`, `/api/channels/{id}/messages` (POST), `/api/channels/{id}/members`, `/api/channels/{id}/reactions`
- **`app/src/handlers_communities.go`** — Routes: `/communities`, `/api/communities` (POST), `/api/communities` (GET), `/api/communities/discover`, `/api/communities/{id}/join`, `/api/communities/{id}/leave`, `/api/communities/{id}/join_request`, `/api/communities/{id}/invite`, `/api/communities/{id}/members`
- **`app/src/seed.go`** — Go port of `seed_local_db.py`; `SeedDatabase(db)` seeds users, alumni, connections, communities, channels with bcrypt-hashed passwords; invoked via `-seed` CLI flag
- **`app/src/render_test.go`** — Template parse/render tests for all 27 Pongo2 templates
- **`app/src/email_test.go`** — Unit tests for all 4 `Build*Email` functions
- **`go.mod` / `go.sum`** — Added `github.com/google/uuid` dependency

### Changed

- **`app/src/main.go`** — Unified Go server now handles web pages, REST API, and WebSockets on port 8080; added `-seed` flag; all routes wired from Go handlers; removed Python proxy logic
- **`scripts/start-all.sh`** — Rewritten: Go-only server startup; builds binary, starts `websocket-server`, Ctrl+C cleanup; removed Python/Flask/port-5000 references
- **`scripts/cleanup.sh`** — Rewritten: kills port 8080 and `websocket-server` processes only; removed Flask/port-5000 kill logic
- **`scripts/test-servers.sh`** — Rewritten: runs `go build`, `go test -v ./...`, `go mod verify`, `go vet`; removed Python syntax checks and `.venv` references
- **`Dockerfile`** — Rewritten as pure Go multi-stage build: `golang:1.22-alpine` builder → `alpine:3.19` runtime; no Python, pip, supervisor, or nginx
- **`deploy/Dockerfile`** — Same pure Go multi-stage build as root Dockerfile
- **`deploy/supervisord.conf`** — Simplified to single `[program:go-backend]` block; removed `[program:python-flask]` and `[program:nginx]`
- **`deploy/Procfile`** — Changed from `gunicorn "app.src.run:app"` to Go build+run command
- **Template fixes** (Pongo2 compatibility): `verification_request.html`, `connect.html`, `requests.html`, `chat.html`, `channels.html`, `chat_list.html`, `forgot_password.html`, `admin_dashboard.html`, `partials/edit_profile_modal.html`

### Removed

- **`app/src/algo/`** — Entire Python Flask application (blueprints, auth, email templates, utils)
- **`app/src/__init__.py`** — Python package init
- **`app/db/seed_local_db.py`** — Python database seeder (replaced by `seed.go`)
- **`tests/`** — Python test suite (`test_basic.py`, `test_endpoints.py`, `test_email_templates.py`)
- **`requirements.txt`** — Python dependencies
- **`.venv/`** — Python virtual environment
- All `__pycache__/` directories

### Fixed

- Pongo2 template engine incompatibilities with Jinja2 syntax (newlines in tags, Python keyword args, dict literals, inline ternary expressions)
- `CheckPasswordHash` now supports legacy Werkzeug PBKDF2/scrypt password hashes from the Python era for zero-downtime migration
- `getSessionRole` return value (1 value, not 2) used correctly throughout handlers
