# Architecture Contracts & System Diagrams

## Component Boundaries

```
                             +------------------------+
                             |   Client / Browser     |
                             | (TS/JS + Tailwind CSS) |
                             +-----------+------------+
                                         |
                       +-----------------+-----------------+
         HTTP Requests |                                   | WebSockets
         (Port 5000)   v                                   v (Port 8080)
     +-----------------------+                    +--------------------+
     |  Python Flask Server  |                    | Go WebSocket Server|
     | - Jinja Templates     |                    | - Channel Manager  |
     | - REST API Endpoints  |                    | - Direct Chat Hub  |
     | - Auth & Session Mgmt |                    | - Presence State   |
     | - Verification Flow   |                    | - OAuth Relay      |
     +-----------+-----------+                    +---------+----------+
                 |                                          |
                 |        +-------------------------+       |
                 +------->|   PostgreSQL Database   |<------+
                          | (Users, Channels, Msgs) |
                          +-------------------------+
```

## System Interfaces & Communication Contracts

### 1. HTTP Server (Flask - Port 5000)
- **Blueprints**:
  - `core.bp`: Public landing, about, contact, global error handlers.
  - `auth.bp`: Registration, login, logout, password resets, email verification.
  - `profile.bp`: User profile details, experience, education, avatar uploads.
  - `dashboard.bp`: Student / Alumni / Admin dashboards, verification review queues.
  - `connections.bp`: Friend/mentor connection requests, status toggles.
  - `settings.bp`: Notification, privacy, and account configuration.
  - `chat.bp`: Chat UI views (`/chat`, `/chat/<username>`).
  - `communities.bp`: Community discovery, join requests, creation.
  - `channels.channels_bp` (`/api` prefix): Channel CRUD, category management, membership list.

### 2. Real-Time WebSocket Engine (Go - Port 8080)
- **Connection Endpoint**: `ws://localhost:8080/ws` or `/ws/channel/{id}`
- **Event Contracts (`MessageType`)**:
  - `join_channel` / `leave_channel`: Client subscription to a specific channel room.
  - `send_channel_message`: Client sends message `{ channel_id, content, reply_to, ... }`.
  - `new_message`: Go broadcasts payload to all connected clients in channel.
  - `get_channel_messages` / `messages_history`: Retrieve paginated channel history.
  - `join_chat_room` / `leave_chat_room`: Join direct 1:1 conversation between two users.
  - `chat_message` / `new_chat_message`: Direct message sent / delivered.
  - `typing_start` / `typing_stop` / `user_typing`: Ephemeral typing indicators.
  - `user_joined` / `user_left` / `user_online`: Real-time presence indicators.

### 3. Database Contracts (PostgreSQL)
- Schema definitions in `app/db/schema.sql`:
  - `users`: Core identity, roles (`student`, `alumni`, `staff`, `admin`), profile fields, verification metadata.
  - `communities`: Educational institution / group entities.
  - `community_members`: M:N relationship with roles (`admin`, `moderator`, `member`).
  - `channels`: Discord-like channels per community (`text`, `voice`, `announcement`).
  - `channel_messages`: Channel chat records with reply threading, soft-delete, and edit flags.
  - `messages`: 1-on-1 direct messaging history.
  - `connections`: Alumni-student mentorship and networking links (`pending`, `accepted`, `denied`).
  - `verification_requests`: Institutional admin review queue.
