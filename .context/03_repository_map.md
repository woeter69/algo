# Repository Map

## Directory Structure & Module Descriptions

```
algo/
├── app/
│   ├── db/
│   │   ├── schema.sql              # Master PostgreSQL schema definition (users, channels, messages, verification)
│   │   └── seed_cic_channels.sql   # Seed script for initial channels and communities
│   ├── src/
│   │   ├── algo/                   # Core Python Flask application package
│   │   │   ├── __init__.py         # App factory (`create_app`), extension initialization & blueprint registry
│   │   │   ├── db.py               # Database connection helpers & pool management
│   │   │   ├── utils.py            # IST timezone converter, avatar generators, datetime formatting
│   │   │   ├── validators.py       # Form, email, password, and session validation helpers
│   │   │   └── blueprints/         # Flask blueprints
│   │   │       ├── auth/           # Login, registration, token verification, password recovery
│   │   │       ├── channels/       # Channel REST endpoints & member management
│   │   │       ├── chat/           # Direct messaging and chat room UI views
│   │   │       ├── communities/    # Community catalog and joining flows
│   │   │       ├── connections/    # Networking and mentorship connection requests
│   │   │       ├── core/           # Static routes (about, contact, error handling)
│   │   │       ├── dashboard/      # Role-based dashboards (student, alumni, admin)
│   │   │       ├── profile/        # User profiles, experiences, education, and photo updates
│   │   │       └── settings/       # Account and privacy settings management
│   │   ├── run.py                  # Flask entrypoint script
│   │   ├── main.go                 # Go WebSocket server entrypoint, message broker, connection tracker
│   │   ├── sockets.go              # WebSocket upgrader, client loop, channel hub management
│   │   ├── channels.go             # Channel membership and message persistence logic in Go
│   │   ├── oauth.go                # Go Google OAuth handlers and auth token validation
│   │   └── start-go-server.sh      # Launcher script for Go WebSocket service
│   ├── static/
│   │   ├── ts/                     # TypeScript sources for frontend interactivity
│   │   │   ├── go-websocket-client.ts # Typed WebSocket client wrapper with auto-reconnect
│   │   │   ├── chat-socket.ts      # Real-time chat socket integration
│   │   │   ├── chat-ui.ts          # Chat DOM manipulation and event handlers
│   │   │   └── channels.ts         # Channel UI controller
│   │   ├── js/                     # Compiled JavaScript assets generated from TypeScript
│   │   ├── styles/                 # Tailwind source (input.css) and compiled bundle (output.css)
│   │   └── data/                   # Seed and bot training datasets
│   └── templates/                  # Jinja2 HTML templates & reusable UI partials
├── configs/                        # TypeScript dev and prod compiler configs
├── deploy/                         # Production containerization (Dockerfile, Nginx, Supervisord, Render configs)
├── docs/                           # Documentation (SETUP.md, OAUTH_SETUP.md, SIH_PROJECT_DESCRIPTION.md)
├── scripts/
│   ├── start-all.sh                # Unified launcher for Python (5000) and Go (8080)
│   ├── cleanup.sh                  # Process cleanup script (kills lingering servers on 5000/8080)
│   └── test-servers.sh             # Health check script verifying both servers
├── tests/                          # Automated pytest suite
├── .context/                       # Living architectural state & persistent agent memory
├── package.json                    # Node.js scripts and frontend dependencies
├── requirements.txt                # Python package requirements
└── tailwind.config.js              # Tailwind CSS configuration
```
