# Alumnigo (algo)

## Project Overview
AlumniGo is a modern, real-time alumni networking platform built for the Smart India Hackathon (SIH) 2025. It connects students, alumni, and institutions via community channels and blazing-fast messaging. 

**Tech Stack & Architecture:**
- **Backend (Go):** Unified high-performance server handling web pages (Pongo2 templates), REST API, and WebSockets. (Runs on port 8080)
- **Database:** PostgreSQL.
- **Frontend:** HTML templates (rendered via Pongo2), Vanilla JavaScript / TypeScript, and TailwindCSS for styling.

## Building and Running

### Prerequisites
- Go 1.21+
- Node.js & npm (for Tailwind/TypeScript)
- PostgreSQL

### Setup & Installation
1. **Go Dependencies:**
   ```bash
   cd app/src/ && go mod tidy && cd ../../
   ```
2. **Frontend Assets:**
   ```bash
   npm install
   npm run build # Builds Tailwind CSS and TypeScript
   ```
3. **Environment Variables:**
   ```bash
   cp .env.example .env
   # Ensure you configure your database and OAuth settings
   ```

### Running the Application
The easiest way to start the server is by using the provided startup script:
```bash
./scripts/start-all.sh
```
*This starts the Go server on port 8080.*

**To run the frontend asset watcher during development:**
```bash
npm run watch
```

**Stopping the application:**
```bash
./scripts/cleanup.sh
```

## Development Conventions
- **Go:** Use `go fmt` for code formatting.
- **Commits:** Write descriptive commit messages.
- **Structure:**
  - `app/src/`: Contains the Go source code.
  - `app/static/`: Contains static assets (CSS, JS, TS, Images).
  - `app/templates/`: Contains Jinja-compatible templates (Pongo2).

