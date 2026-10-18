# Unified Task Matrix

## Epics & Subtasks

### Epic 1: Environment & Project Initialization [COMPLETED]
- [x] **TASK-001**: Configure Python 3.12 environment in `.venv` and install `requirements.txt`.
- [x] **TASK-002**: Symlink Go 1.22 binary and tidy Go modules in `app/src/`.
- [x] **TASK-003**: Compile frontend assets (TypeScript and Tailwind CSS via `npm run build`).
- [x] **TASK-004**: Verify `.env` configuration file keys and paths.

### Epic 2: Autonomous Context Architecture [COMPLETED]
- [x] **TASK-005**: Initialize `.context/` living memory matrix using `init_context.sh`.
- [x] **TASK-006**: Populate project manifest, architecture contracts, tech stack specs, repository map, and initial ADRs.

### Epic 3: Graphify Knowledge Graph [COMPLETED]
- [x] **TASK-007**: Run graphify file detection and AST extraction across Python, Go, TypeScript, and SQL code.
- [x] **TASK-008**: Build graph, cluster modules, detect god nodes & bridges, and label communities.
- [x] **TASK-009**: Export interactive HTML visualizer (`graphify-out/graph.html`) and generate `GRAPH_REPORT.md`.

### Epic 4: Local Database & Testing Verification [COMPLETED]
- [x] **TASK-010**: Install user-space PostgreSQL 18.6 and initialize local cluster on port 5432.
- [x] **TASK-011**: Resolve schema circular foreign key ordering in `app/db/schema.sql` and apply schema.
- [x] **TASK-012**: Seed local database (`seed_local_db.py`) with admin, student, and alumni test users and channels.
- [x] **TASK-013**: Resolve blueprint routing, template errors, and pool closure issues (`mydb.close()`).
- [x] **TASK-014**: Execute automated test suite (`pytest tests/`) and end-to-end integration tests with 100% pass rate.

### Epic 5: Contact Inquiry Pipeline & Admin Resolution [COMPLETED]
- [x] **TASK-015**: Database schema migration adding resolution tracking columns (`status`, `resolved_by`, `resolved_at`, `resolution_notes`, index) to `contacts` table.
- [x] **TASK-016**: Contact submission query persistence and asynchronous SMTP forwarding to `alumnigo.sih@gmail.com`.
- [x] **TASK-017**: Admin dashboard metrics and management interface for viewing and resolving contact queries.
- [x] **TASK-018**: End-to-end automated test suite verifying submission, persistence, dashboard display, and admin resolution.
- [x] **TASK-019**: Implement querier email notification `send_inquiry_resolved_email` on inquiry resolution.
- [x] **TASK-020**: Admin dashboard header auth state visualization, logout trigger, and flash messaging banners.

### Epic 6: Branded HTML Email Template Architecture [COMPLETED]
- [x] **TASK-021**: Build responsive, modern HTML email templates matching ALGO theme with gradient banner, status badges, card containers, and mobile responsiveness (`app/src/algo/email_templates.py`).
- [x] **TASK-022**: Upgrade all system email dispatchers (contact inquiry, resolution notification, password reset/change) to `multipart/alternative` with HTML and plain-text fallback.
- [x] **TASK-023**: Implement automated unit test suite for email builders (`tests/test_email_templates.py`).

### Epic 7: Brand & Year Standardization [COMPLETED]
- [x] **TASK-024**: Update copyright year to 2026 across all 20 HTML templates in `app/templates/`.
- [x] **TASK-025**: Remove mentions of Smart India Hackathon and standardize year to 2026 in email templates and chatbot knowledge bases.

### Epic 8: Contact Auto-Responder System [COMPLETED]
- [x] **TASK-026**: Design branded auto-confirmation HTML email template `build_inquiry_confirmation_email` with Reference ID tracking.
- [x] **TASK-027**: Implement asynchronous `send_inquiry_confirmation_email` in `utils.py` and hook into `/contact` route in `core.py`.
- [x] **TASK-028**: Add automated unit test for auto-confirmation template (`tests/test_email_templates.py`).

### Epic 9: Inbound Reply Routing [COMPLETED]
- [x] **TASK-029**: Configure `Reply-To: alumnigo.sih@gmail.com` across all outgoing emails so recipient replies route directly to the AlumniGo team mailbox.
- [x] **TASK-030**: Test and verify delivery to all 4 team members (`pranjul.here@gmail.com`, `adityabhagora@gmail.com`, `chandragupt.jsr@gmail.com`, `himanshu809809@gmail.com`).

### Epic 10: Email Layout Refinements & Card Header Restructuring [COMPLETED]
- [x] **TASK-031**: Restore status badges on the header banner across all email templates.
- [x] **TASK-032**: Remove inline pill badges right above greetings and titles in email content bodies.
- [x] **TASK-033**: Redesign Admin Response card: move title ("Admin Response / Resolution Note") and subtitle ("Resolved on ... by ...") outside the green box matching its theme, keeping only the response text inside the box.
- [x] **TASK-034**: Verify email tests and live dispatch.

### Epic 11: Upstream Sync, Pull Request & Merge [COMPLETED]
- [x] **TASK-035**: Synced and merged latest upstream commits (`woeter69/algo:main`), resolving schema foreign key ordering.
- [x] **TASK-036**: Pushed all feature commits to fork (`Pranjul-00/algo:main`).
- [x] **TASK-037**: Created Pull Request #99 on `woeter69/algo`.
- [x] **TASK-038**: Merged PR #99 into `woeter69/algo:main` and synced local and fork remotes.

### Epic 12: Docker CI Workflow & Containerization Fixes [COMPLETED]
- [x] **TASK-039**: Restored root `Dockerfile` and configured root build context with `.dockerignore` for GitHub Actions CI.
- [x] **TASK-040**: Corrected deployment paths in `deploy/Dockerfile` and fixed entrypoint command in `deploy/supervisord.conf` (`run.py`).
- [x] **TASK-041**: Verified `Docker Image CI` workflow passes with success on GitHub Actions (`woeter69/algo` run #36720643486).

### Epic 13: DM Engine & Real-Time Chat [COMPLETED]
- [x] **TASK-042**: Upgraded `go-websocket-client.ts` with direct chat message routing (`sendDirectMessage`, `sendChannelMessage`, direct typing indicators).
- [x] **TASK-043**: Implemented `/api/online_status`, `/api/search_users`, and chat attachments in `app/src/algo/blueprints/chat.py`.
- [x] **TASK-044**: Connected "New Message" modal, live search, and optimistic attachment delivery in `app/static/ts/chat.ts`.
- [x] **TASK-045**: Verified chat endpoints and WebSocket communications with automated tests.

### Epic 14: Channels Architecture & Real-Time Persistence [COMPLETED]
- [x] **TASK-046**: Rewrote `app/static/ts/channels.ts`: dynamic `loadChannelMessages(channelId)` from `GET /channels/<id>/messages`, `loadChannelMembers(channelId)` with active user preview modal, typing indicator, and real-time message broadcasting via Go WebSockets.
- [x] **TASK-047**: Implemented public channel fallback in `get_channel_members` (`app/src/algo/blueprints/channels.py`) to query `community_members` so members list is never stuck on loading.
- [x] **TASK-048**: Added "+ Create Channel" button and modal dialog in `app/templates/channels.html`, posting to `POST /communities/<id>/channels`.

### Epic 15: Verification Workflow & Limited Dashboard [COMPLETED]
- [x] **TASK-049**: Fixed broken `url_for('profile_redirect')` and `url_for('verification_request')` in `app/templates/limited_dashboard.html`.
- [x] **TASK-050**: Connected `verification_request.html` to dynamically load colleges from `communities` table.
- [x] **TASK-051**: Implemented full submission in `dashboard.verification_request`: inserts into `verification_requests` table and updates `users.verification_status = 'pending'`.
- [x] **TASK-052**: Updated `admin_dashboard()` to join `verification_requests` and `communities` to display real student ID, department, grad year, college, and message.
- [x] **TASK-053**: Updated `/api/handle_verification_request` to approve/reject both `users` and `verification_requests`, and auto-enrolled approved users into `community_members`.
- [x] **TASK-054**: Added root `/logout` route forwarder in `core.py`.

### Epic 16: User Dashboard Metrics & Dynamic Feed [COMPLETED]
- [x] **TASK-055**: Replaced hardcoded stats (1,247 alumni, 89 jobs, etc.) in `app/src/algo/blueprints/dashboard.py` (`user_dashboard`) and `app/templates/user_dashboard.html` with real DB queries (`total_alumni`, `total_connections`, `total_channels`, `pending_requests`).
- [x] **TASK-056**: Replaced broken `href="#"` buttons with working endpoints: Browse Alumni -> `connections.connect`, Requests -> `connections.requests_page`, Channels -> `communities.channels`, Open Chat -> `chat.chat_list`.
- [x] **TASK-057**: Connected dropdown menu items to `/profile`, `/settings`, and `/logout`.
- [x] **TASK-058**: Rendered dynamic recent activity feed of incoming connection requests.

### Epic 17: User Profile Editing & Experience/Education CRUD [COMPLETED]
- [x] **TASK-059**: Populated `app/templates/partials/edit_profile_modal.html` with real user attributes instead of dummy data.
- [x] **TASK-060**: Implemented `normalize_degree_type` and added `/api/profile/update`, `/api/profile/experience` (POST & DELETE), and `/api/profile/education` (POST & DELETE) in `app/src/algo/blueprints/profile.py`.
- [x] **TASK-061**: Updated `app/static/ts/profile.ts` to submit `#profileForm` via `fetch('/api/profile/update')`.
- [x] **TASK-062**: Added avatar file upload with local disk storage in `app/static/uploads/avatars/`.

### Epic 18: Community Discovery & Settings Persistence [COMPLETED]
- [x] **TASK-063**: Implemented `/communities/discover`, `/communities/<id>/join` (direct join for verified members), and `/communities/<id>/leave` in `channels.py`.
- [x] **TASK-064**: Built "Explore Communities" modal dialog in `app/templates/channels.html` and connected in `app/static/ts/channels.ts`.
- [x] **TASK-065**: Wired Settings page (`settings.py` and `settings.html`) with real DB persistence for `profile_visibility`, `email_notifications`, and `job_alerts`.
- [x] **TASK-066**: Fixed tuple index out of range crash in `/api/export_data` JSON export endpoint.

### Epic 19: Navigation Link & Footer Audit [COMPLETED]
- [x] **TASK-067**: Audited and replaced dead `href="#"` navbar and footer links across `register.html`, `thanks.html`, `token_expired.html`, `token_invalid.html`, `verification_request.html`, `user_dashboard.html`, and `profile.html`.
- [x] **TASK-068**: Ensured all automated unit, integration, and endpoint tests pass 100% (16 of 16 tests passing).








