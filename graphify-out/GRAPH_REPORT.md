# Graph Report - algo  (2026-09-30)

## Corpus Check
- 126 files · ~117,948 words
- Verdict: corpus is large enough that graph structure adds value.
- Unclassified: 32 file(s) not represented in the graph (top: .css 25, (none) 4, .conf 2)

## Summary
- 965 nodes · 1506 edges · 76 communities (56 shown, 20 thin omitted)
- Extraction: 92% EXTRACTED · 8% INFERRED · 0% AMBIGUOUS · INFERRED: 116 edges (avg confidence: 0.88)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- Server Utilities & Mail
- Database Schema & Tables
- User Profile Management
- Role Access & Verification
- Frontend Chat UI & Emojis
- Frontend Build & Dependencies
- Admin Dashboard Interface
- Profile Media & Uploads
- Role Access & Verification
- OAuth & Network Handlers
- OAuth & Network Handlers
- Common DOM Utilities
- User Settings & Privacy
- Frontend Chat UI & Emojis
- Client WebSocket Protocol
- Tsconfig Operations
- Networking & Alumni Directory
- Frontend Chat UI & Emojis
- UI Templates & Modals
- OAuth & Network Handlers
- Profile Media & Uploads
- Role Access & Verification
- Database Schema & Tables
- Networking & Alumni Directory
- UI Templates & Modals
- OAuth & Network Handlers
- Role Access & Verification
- Networking & Alumni Directory
- User Dashboard
- Networking & Alumni Directory
- OAuth & Network Handlers
- Frontend Chat UI & Emojis
- Community Channel Operations
- Blueprints Auth
- UI Templates & Modals
- Blueprints Core
- UI Templates & Modals
- Deployment & DevOps
- Testing & Test Harness
- Go WebSocket Hub & Channels
- Frontend Chat UI & Emojis
- Docs Setup
- Questions Operations
- UI Templates & Modals
- Home Operations
- Login Operations
- UI Templates & Modals
- Architecture State & Context
- Db Close Db
- UI Templates & Modals
- Configs Tsconfig Dev
- Configs Tsconfig Prod
- Architecture State & Context
- Frontend Build & Dependencies
- Go WebSocket Hub & Channels
- Register Operations
- Thanks Operations
- Token Expired
- Token Invalid
- Networking & Alumni Directory
- Init Create App
- Chat Socket
- Scripts Start All
- Start Go Server
- Go WebSocket Hub & Channels
- Scripts Cleanup
- Testing & Test Harness
- Architecture State & Context
- Architecture State & Context
- Pkg Websocket Server

## God Nodes (most connected - your core abstractions)
1. `get_db()` - 33 edges
2. `users` - 22 edges
3. `Hub` - 22 edges
4. `GoWebSocketClient` - 20 edges
5. `WSMessage` - 19 edges
6. `get_db_connection()` - 18 edges
7. `GoogleOAuth` - 18 edges
8. `scripts` - 17 edges
9. `compilerOptions` - 16 edges
10. `require_auth()` - 13 edges

## Surprising Connections (you probably didn't know these)
- `Flask Web Framework & Extensions` --conceptually_related_to--> `Python Flask Server (Port 5000)`  [INFERRED]
  requirements.txt → .context/01_architecture_contracts.md
- `AlumniGo Quick Start Setup Procedure` --semantically_similar_to--> `Unified Startup Orchestration (scripts/start-all.sh)`  [INFERRED] [semantically similar]
  docs/SETUP.md → GEMINI.md
- `Unified Startup Orchestration (scripts/start-all.sh)` --conceptually_related_to--> `Hybrid Micro-Service Architecture`  [INFERRED]
  GEMINI.md → .context/00_project_manifest.md
- `AlumniGo Flask Application Render Service` --conceptually_related_to--> `Python Flask Server (Port 5000)`  [INFERRED]
  deploy/render.yaml → .context/01_architecture_contracts.md
- `Algo WebSocket Server Render Service` --conceptually_related_to--> `Go WebSocket Server (Port 8080)`  [INFERRED]
  deploy/render-go.yaml → .context/01_architecture_contracts.md

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Dual-Server Polyglot Architecture (Flask HTTP + Go WebSocket)** — context_01_architecture_contracts_flask_server, context_01_architecture_contracts_go_websocket_server, context_04_decision_log_adr_adr_001 [INFERRED 0.85]
- **Direct PostgreSQL Shared Database Contract** — context_01_architecture_contracts_postgresql_database, context_04_decision_log_adr_adr_002, context_00_project_manifest_database_integrity [INFERRED 0.85]
- **Role-Based Access Control and Institutional Verification Framework** — context_00_project_manifest_role_based_hierarchy, context_00_project_manifest_institutional_verification, docs_sih_project_description_role_based_access_control [INFERRED 0.85]
- **User Registration, Authentication and Onboarding Flow** — app_templates_register, app_templates_login, app_templates_check_email, app_templates_complete_profile, app_templates_interests, app_templates_confirmation [INFERRED 0.95]
- **Real-time Messaging and Community Channels Suite** — app_templates_channels, app_templates_chat, app_templates_chat_list [INFERRED 0.95]
- **Alumni Directory, Profile Inspection and Admin Verification Suite** — app_templates_connect, app_templates_profile, app_templates_limited_dashboard, app_templates_admin_dashboard [INFERRED 0.85]

## Communities (76 total, 20 thin omitted)

### Community 0 - "Server Utilities & Mail"
Cohesion: 0.05
Nodes (17): create_password_changed_email_content(), create_password_reset_email_content(), format_ist_time(), format_utc_timestamp(), generate_default_avatar(), get_room_id(), is_student_email(), load_cities() (+9 more)

### Community 1 - "Database Schema & Tables"
Cohesion: 0.09
Nodes (45): admin_permissions, channel_members, channel_messages, channels, communities, community_invitations, community_join_requests, community_members (+37 more)

### Community 2 - "User Profile Management"
Cohesion: 0.06
Nodes (25): complete_profile(), interests(), profile_redirect(), user_profile(), is_safe_url(), clearFieldError(), form, hamburger (+17 more)

### Community 3 - "Role Access & Verification"
Cohesion: 0.10
Nodes (18): get_communities(), is_admin(), add_reaction(), broadcast_to_go_websocket(), create_channel(), get_channel_members(), get_channel_messages(), get_channels() (+10 more)

### Community 4 - "Frontend Chat UI & Emojis"
Cohesion: 0.12
Nodes (25): blockUser(), clearChat(), emojiData, ensureConversationItem(), fetchOnlineStatus(), getConversationItem(), handleMoreOptionsAction(), initializeEmojiPicker() (+17 more)

### Community 5 - "Frontend Build & Dependencies"
Cohesion: 0.06
Nodes (31): description, devDependencies, tailwindcss, @tailwindcss/forms, @tailwindcss/typography, @types/node, typescript, name (+23 more)

### Community 6 - "Admin Dashboard Interface"
Cohesion: 0.08
Nodes (19): adminStyle, closeReviewModal(), initializeAdminDashboard(), initializeAdminNavigation(), initializeFilters(), initializeModal(), featureCards, hamburgerLimited (+11 more)

### Community 7 - "Profile Media & Uploads"
Cohesion: 0.08
Nodes (20): initProfilePictureUpload(), uploadProfilePicture(), showNotification(), validateField(), validateForm(), validatePasswordChange(), Edit Profile Education Tab, Edit Profile Personal Info Tab (+12 more)

### Community 8 - "Role Access & Verification"
Cohesion: 0.07
Nodes (14): login_required(), decorated_function(), approve_verification_request(), can_access_community_features(), community_access_required(), decorated_function(), get_pending_verification_requests(), get_user_role_info() (+6 more)

### Community 9 - "OAuth & Network Handlers"
Cohesion: 0.11
Nodes (6): connectDB(), handleBroadcastAPI(), main(), NewHub(), NewGoogleOAuth(), MessageType

### Community 10 - "OAuth & Network Handlers"
Cohesion: 0.18
Nodes (6): enableCORS(), healthCheck(), setCORSHeaders(), GoogleOAuth, GoogleUserInfo, UserSession

### Community 11 - "Common DOM Utilities"
Cohesion: 0.13
Nodes (7): DOMUtils, FormValidator, initializeProfilePictures(), Navigation, NotificationSystem, observer, showInitials()

### Community 12 - "User Settings & Privacy"
Cohesion: 0.24
Nodes (19): applyTheme(), deactivateAccount(), deleteAccount(), exportData(), handleAccountSubmit(), handleNotificationsSubmit(), handlePreferencesSubmit(), handlePrivacySubmit() (+11 more)

### Community 13 - "Frontend Chat UI & Emojis"
Cohesion: 0.26
Nodes (3): getChannelID(), Hub, WSMessage

### Community 16 - "Tsconfig Operations"
Cohesion: 0.11
Nodes (18): compilerOptions, declaration, esModuleInterop, forceConsistentCasingInFileNames, lib, module, moduleResolution, noImplicitAny (+10 more)

### Community 17 - "Networking & Alumni Directory"
Cohesion: 0.15
Nodes (13): applyFilters(), ConnectData, createPersonCard(), escapeHtml(), handleLogout(), initializeUserDropdown(), openConnectionModal(), openProfileModal() (+5 more)

### Community 18 - "Frontend Chat UI & Emojis"
Cohesion: 0.12
Nodes (11): channels, chatMessages, currentChannelIcon, currentChannelName, hamburger, messageInput, navLinks, sendBtn (+3 more)

### Community 19 - "UI Templates & Modals"
Cohesion: 0.15
Nodes (9): cancelBtn, closeModal, modalError, newChatBtn, newChatModal, OnlineStatusManager, OnlineStatusResponse, startChatBtn (+1 more)

### Community 20 - "OAuth & Network Handlers"
Cohesion: 0.18
Nodes (6): Python Flask Server (Port 5000), Go WebSocket Server (Port 8080), PostgreSQL Shared Database, AlumniGo Flask Application Render Service, Algo WebSocket Server Render Service, Google OAuth 2.0 Authentication Workflow

### Community 21 - "Profile Media & Uploads"
Cohesion: 0.35
Nodes (12): change_password(), deactivate_account(), delete_account(), export_data(), logout_all_sessions(), settings_page(), update_account(), update_notifications() (+4 more)

### Community 22 - "Role Access & Verification"
Cohesion: 0.16
Nodes (7): is_verified_user(), verified_user_required(), decorated_function(), channels(), check_community_access(), create_community(), user_role_info()

### Community 23 - "Database Schema & Tables"
Cohesion: 0.14
Nodes (11): Environment Variables Schema (.env), Go 1.22 Runtime Environment, Node.js & npm Frontend Tooling, PostgreSQL 14+ Database Runtime, Python 3.12 Runtime Environment, ERR-001: Missing psycopg2 Wheel on Python 3.14, ERR-002: Missing Go Executable on System Path, Flask-Bcrypt & OAuth Security Suite (+3 more)

### Community 24 - "Networking & Alumni Directory"
Cohesion: 0.20
Nodes (6): cancel_connection_request(), connect(), get_connection_requests(), requests_page(), respond_connection_request(), send_connection_request()

### Community 25 - "UI Templates & Modals"
Cohesion: 0.25
Nodes (11): clearFieldError(), escapeHtml(), handleFormSubmission(), initializeForm(), initializeValidation(), showFieldError(), showLoading(), showNotification() (+3 more)

### Community 26 - "OAuth & Network Handlers"
Cohesion: 0.20
Nodes (9): Base HTML Layout Template, Global Navigation and Footer, Homepage Hero Section & Value Propositions, Public Landing Page View, Google OAuth Login Button, User Authentication Login View, Redesigned Homepage View, Role Selection Component (+1 more)

### Community 27 - "Role Access & Verification"
Cohesion: 0.22
Nodes (6): admin_required(), admin_dashboard(), handle_verification_request(), limited_dashboard(), user_dashboard(), verification_request()

### Community 28 - "Networking & Alumni Directory"
Cohesion: 0.24
Nodes (9): Connection, ConnectionRequest, escapeHtml(), handleCancelRequest(), handleConnectionResponse(), RequestsData, showLoading(), showNotification() (+1 more)

### Community 29 - "User Dashboard"
Cohesion: 0.18
Nodes (12): actionCards, changeSlide(), currentSlide(), dropdownArrow, hamburger, links, logoutBtn, navLinks (+4 more)

### Community 30 - "Networking & Alumni Directory"
Cohesion: 0.21
Nodes (9): Connections Diagram Image, People Illustration Image, Alumni Recommendation and Filter Grid, Alumni Directory and Networking View, Channel Profile Preview Modal, Alumni Profile Connect Preview Modal, Send Connection Request Form Dialog, Alumni Connection Management (+1 more)

### Community 32 - "Frontend Chat UI & Emojis"
Cohesion: 0.20
Nodes (7): Channel Real-time Messaging Window, Community Channels View, 1-on-1 Real-time Message Thread, Direct Message Chat View, Start New Chat Modal Dialog, New Message User Search List Modal, Real-time Chat & User Messaging

### Community 33 - "Community Channel Operations"
Cohesion: 0.18
Nodes (9): Channel & Community Infrastructure, Institutional Verification Workflow, Sub-millisecond Real-Time Messaging, Real-Time WebSocket Engine Contract, WebSocket MessageType Event Contracts, AlumniGo Platform Overview, Community Channels Feature, Real-Time Chat Feature (+1 more)

### Community 34 - "Blueprints Auth"
Cohesion: 0.31
Nodes (6): check_email(), confirmation(), login(), logout(), register(), verify()

### Community 35 - "UI Templates & Modals"
Cohesion: 0.27
Nodes (5): chat(), chat_list(), get_online_status(), upload_image(), Conversations Directory View

### Community 36 - "Blueprints Core"
Cohesion: 0.29
Nodes (5): about(), contact(), home(), recommendations(), thanks()

### Community 37 - "UI Templates & Modals"
Cohesion: 0.20
Nodes (6): hamburger, links, navLinks, About Us Page View, AlumniGo Team Showcase Section, Contact Support and Feedback View

### Community 38 - "Deployment & DevOps"
Cohesion: 0.36
Nodes (7): initializeAboutNavigation(), initializeAboutPage(), initializeAnimations(), initializeTeamCards(), renderTeamMembers(), teamMembers, updateTeamMember()

### Community 39 - "Testing & Test Harness"
Cohesion: 0.25
Nodes (3): app(), client(), test_home_page()

### Community 40 - "Go WebSocket Hub & Channels"
Cohesion: 0.29
Nodes (3): Hub, ChannelMessage, UserInfo

### Community 41 - "Frontend Chat UI & Emojis"
Cohesion: 0.25
Nodes (7): ChatData, ClientToServerEvents, JoinRoomData, MessageData, ServerToClientEvents, UserStatusData, Window

### Community 42 - "Docs Setup"
Cohesion: 0.25
Nodes (6): PostgreSQL Database Initialization, Installation Prerequisites & Dependencies, AlumniGo Quick Start Setup Procedure, SIH Problem Statement & Digital Ecosystem, Smart India Hackathon 2025 Context, Unified Startup Orchestration (scripts/start-all.sh)

### Community 44 - "UI Templates & Modals"
Cohesion: 0.29
Nodes (4): hamburger, links, navLinks, Email Verification Prompt View

### Community 45 - "Home Operations"
Cohesion: 0.29
Nodes (5): allSections, hamburger, links, navItems, navLinks

### Community 46 - "Login Operations"
Cohesion: 0.29
Nodes (4): hamburger, links, loginForm, navLinks

### Community 47 - "UI Templates & Modals"
Cohesion: 0.52
Nodes (6): addUserDropdownHTML(), generateDefaultAvatar(), handleLogout(), initializeDefaultAvatars(), initializeSharedComponents(), initializeUserDropdown()

### Community 48 - "Architecture State & Context"
Cohesion: 0.29
Nodes (5): Epic 1: Environment & Project Initialization, Epic 2: Autonomous Context Architecture, Epic 3: Graphify Knowledge Graph, Epic 4: Verification & Runtime Testing, Active Session Snapshot

### Community 49 - "Db Close Db"
Cohesion: 0.33
Nodes (3): close_db(), init_app(), init_db_pool()

### Community 50 - "UI Templates & Modals"
Cohesion: 0.33
Nodes (5): User Mini Profile Summary Card, Dashboard Quick Action Cards, Recent Activity Timeline, User Dashboard Metrics Grid, User Dashboard & Network Telemetry

### Community 51 - "Configs Tsconfig Dev"
Cohesion: 0.33
Nodes (5): compilerOptions, removeComments, sourceMap, extends, ../tsconfig.json

### Community 52 - "Configs Tsconfig Prod"
Cohesion: 0.33
Nodes (5): compilerOptions, removeComments, sourceMap, extends, ../tsconfig.json

### Community 53 - "Architecture State & Context"
Cohesion: 0.33
Nodes (5): Role-Based Hierarchy Access Control, SIH Role-Tailored Dashboard System, SIH Role-Based Access Control Matrix, SIH Multi-Parameter Search & Discovery System, SIH Solution Architecture & Value Proposition

### Community 54 - "Frontend Build & Dependencies"
Cohesion: 0.33
Nodes (5): Flask Blueprints Architecture, AlumniGo Application Package (app/src/algo), Database Schema & Seed Scripts (app/db), Deployment Configurations (deploy/), Frontend TypeScript Architecture (app/static/ts)

### Community 56 - "Register Operations"
Cohesion: 0.40
Nodes (3): hamburger, links, navLinks

### Community 57 - "Thanks Operations"
Cohesion: 0.40
Nodes (3): hamburger, links, navLinks

### Community 58 - "Token Expired"
Cohesion: 0.40
Nodes (3): hamburger, links, navLinks

### Community 59 - "Token Invalid"
Cohesion: 0.40
Nodes (3): hamburger, links, navLinks

### Community 60 - "Networking & Alumni Directory"
Cohesion: 0.40
Nodes (4): Active Connections Directory Tab, Incoming Connection Requests Tab, Sent Connection Requests Tab, Connection Requests Stats Grid

## Knowledge Gaps
- **215 isolated node(s):** `verification_tokens`, `contacts`, `websocket-server`, `start-go-server.sh script`, `teamMembers` (+210 more)
  These have ≤1 connection - possible missing edges or undocumented components. (Counts symbols only; 400 node(s) total have ≤1 connection when file, concept and rationale nodes are included.)
- **20 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `get_db()` connect `Profile Media & Uploads` to `Blueprints Auth`, `UI Templates & Modals`, `Role Access & Verification`, `User Profile Management`, `Role Access & Verification`, `Db Close Db`, `Role Access & Verification`, `Networking & Alumni Directory`, `Role Access & Verification`?**
  _High betweenness centrality (0.137) - this node is a cross-community bridge._
- **Why does `chat()` connect `UI Templates & Modals` to `Frontend Chat UI & Emojis`, `Profile Media & Uploads`, `Role Access & Verification`?**
  _High betweenness centrality (0.049) - this node is a cross-community bridge._
- **Why does `login()` connect `Blueprints Auth` to `OAuth & Network Handlers`, `Profile Media & Uploads`?**
  _High betweenness centrality (0.031) - this node is a cross-community bridge._
- **Are the 30 inferred relationships involving `get_db()` (e.g. with `login()` and `register()`) actually correct?**
  _`get_db()` has 30 INFERRED edges - model-reasoned connections that need verification._
- **What connects `verification_tokens`, `contacts`, `websocket-server` to the rest of the system?**
  _215 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Server Utilities & Mail` be split into smaller, more focused modules?**
  _Cohesion score 0.04609929078014184 - nodes in this community are weakly interconnected._
- **Should `Database Schema & Tables` be split into smaller, more focused modules?**
  _Cohesion score 0.0927536231884058 - nodes in this community are weakly interconnected._