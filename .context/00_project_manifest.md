# Project Manifest & Mission Objectives

## Core Mission
AlumniGo (algo) is a high-performance alumni networking platform developed for the Smart India Hackathon (SIH) 2025. It bridges students, alumni, and educational institutions through real-time communication, structured communities, professional mentorship, and career growth pathways.

## Strategic Objectives & Acceptance Criteria
- [x] **Hybrid Micro-Service Architecture**: Split CPU-bound web serving/auth (Python Flask) from high-throughput real-time stateful connections (Go WebSocket).
- [x] **Sub-millisecond Real-Time Messaging**: Handle 10,000+ concurrent connections with Go-based WebSocket channels and direct peer messaging.
- [x] **Role-Based Hierarchy**: Enforce access control for Students, Alumni, Staff/Faculty, and Institutional Administrators.
- [x] **Channel & Community Infrastructure**: Discord/Slack-style structured channels (text, announcement, voice) under college and interest communities.
- [x] **Institutional Verification**: Multi-stage verification workflows (student ID, alumni proof, admin verification requests).
- [ ] **End-to-End Chat Resilience**: Distributed pub/sub or Redis clustering for multi-instance scaling of Go sockets.
- [ ] **AI-Powered Recommendation & Mentorship**: Algorithmic matching based on career trajectory, skill set, and shared interests.

## Non-Negotiable Domain Rules
1. **Separation of Concerns**: Flask handles HTTP routes, templates, standard REST APIs, and auth; Go server handles WebSocket connections and high-speed chat broadcasting on port 8080.
2. **Database Integrity**: PostgreSQL is the single source of truth for both Python and Go services. Foreign keys, cascade behaviors, and transaction consistency must be respected.
3. **Password Security**: Passwords must always be hashed with `bcrypt` (Flask-Bcrypt). Plaintext passwords must never hit logs or persistent storage.
4. **Verification Gate**: Unverified users must have restricted platform actions (limited dashboard) until verified by an institution admin or verified email/student token.
5. **Real-time Event Schema**: WebSocket payloads exchanged between frontend clients and Go server must conform to established `MessageType` definitions (`join_channel`, `send_channel_message`, `chat_message`, `typing_start`, etc.).
