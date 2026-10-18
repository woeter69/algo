# Architectural Decision Records (ADR Log)

## Decision Log

### ADR-001: Polyglot Dual-Engine Architecture (Flask + Go)
- **Status**: Accepted
- **Context**: The application requires traditional MVC web views, form handling, and relational data modeling alongside high-concurrency real-time WebSocket communication for SIH 2025.
- **Decision**: Split the application into two communicating runtimes:
  - Python Flask (Port 5000) for standard page rendering, user authentication, profile forms, and REST endpoints.
  - Go (Port 8080) utilizing `gorilla/websocket` for low-latency, high-throughput chat rooms, presence monitoring, and instant message broadcast.
- **Consequences**:
  - *Pros*: Go drastically reduces memory footprint and easily manages 10,000+ concurrent WebSocket connections without GIL or event-loop starvation.
  - *Cons*: Dual-service orchestration required in development and deployment (`supervisord` / dual processes).

### ADR-002: Direct PostgreSQL Shared State
- **Status**: Accepted
- **Context**: Both Flask and Go services need access to user profiles, channel metadata, and message history.
- **Decision**: Connect both Flask and Go directly to the shared PostgreSQL database rather than routing chat storage requests through Flask HTTP endpoints.
- **Consequences**:
  - *Pros*: Minimizes latency during real-time message persistence; Go inserts directly into `messages` and `channel_messages`.
  - *Cons*: Both codebases must adhere strictly to the shared SQL schema and handle connection pooling gracefully.

### ADR-003: Python 3.12 Virtual Environment via `uv`
- **Status**: Accepted
- **Context**: System default Python on Arch Linux is 3.14, for which `psycopg2-binary` pre-compiled wheels are not yet available, causing C-extension compilation errors in the absence of `pg_config`.
- **Decision**: Pin the virtual environment to CPython 3.12 managed via `uv` in `.venv/`.
- **Consequences**:
  - *Pros*: Clean wheel installation for `psycopg2-binary==2.9.10`, fast builds, guaranteed stability.
  - *Cons*: Requires `uv` or Python 3.12 binary available on the host.

### ADR-004: Local User-Space PostgreSQL Deployment
- **Status**: Accepted
- **Context**: The remote Render and Neon PostgreSQL instances were unreachable, rendering the web platform unresponsive.
- **Decision**: Deployed a fully isolated local PostgreSQL 18.6 instance running in user space at `~/.local/pgsql` with data directory `~/.local/share/postgresql/alumnigo_data`, bound to `localhost:5432`.
- **Consequences**:
  - *Pros*: Zero external network dependencies, instantaneous response times, zero cloud costs during local dev and testing.
  - *Cons*: Requires running PostgreSQL locally when testing.
