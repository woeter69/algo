# Tech Stack & Runtime Environment

## Runtimes & Tooling
- **Python**: 3.12 (managed via `uv`, virtualenv in `.venv`)
  - Flask 3.1.2, Flask-Bcrypt 1.0.1, Flask-Dance 7.1.0, psycopg2-binary 2.9.10, Gunicorn 23.0.0
- **Go**: 1.22.5 (binary in `~/.local/go_dist/go/bin/go`, symlinked to `~/.local/bin/go`)
  - Modules: `github.com/gorilla/websocket`, `github.com/lib/pq`, `golang.org/x/oauth2`
- **Node.js**: v22+ / npm 12+
  - Dependencies: Tailwind CSS CLI 3.4.10, TypeScript 5.5.4
- **Database**: PostgreSQL 14+
- **Process Management**:
  - Development: `./scripts/start-all.sh`, `./scripts/cleanup.sh`
  - Production: Docker (`deploy/Dockerfile`), Nginx proxy (`deploy/nginx.conf`), Supervisord (`deploy/supervisord.conf`)

## Environment Variables (.env Schema)

| Variable | Type | Description | Default / Example |
| :--- | :--- | :--- | :--- |
| `APP_URL` | String | Public URL of web app | `http://localhost:5000` |
| `PORT` | Integer | Flask web server port | `5000` (or `10000` on PaaS) |
| `FLASK_APP` | String | Flask entrypoint file | `app/src/run.py` |
| `FLASK_ENV` | String | Environment mode | `development` / `production` |
| `FLASK_DEBUG` | Boolean | Flask debug flag | `True` |
| `SECRET_KEY` | String | Flask session signing secret | `your-super-secret-key` |
| `DB_HOST` | String | PostgreSQL host address | `localhost` |
| `DB_PORT` | Integer | PostgreSQL port | `5432` |
| `DB_NAME` | String | PostgreSQL database name | `alumni_platform` |
| `DB_USER` | String | PostgreSQL username | `postgres` |
| `DB_PASSWORD` | String | PostgreSQL password | `postgres` |
| `DATABASE_URL` | String | Unified connection string | `postgresql://user:pass@host:5432/db` |
| `EMAIL_USER` | String | SMTP email user for verification | `user@example.com` |
| `EMAIL_PASS` | String | SMTP application password | `app-password` |
| `PFP_API` | String | Profile picture upload API key | `imgbb / cloud api key` |
| `TEST_EMAILS` | String | Comma-separated test recipient emails | `test@example.com` |

## Build Commands
```bash
# Build frontend assets (CSS + TypeScript)
npm run build

# Watch mode for frontend development
npm run watch

# Python dependencies
uv pip install --python .venv/bin/python -r requirements.txt

# Go dependencies
cd app/src && go mod tidy
```
