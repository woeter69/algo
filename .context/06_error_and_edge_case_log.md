# Error & Edge Case Debugging Ledger

## Debugging Ledger

### ERR-001: Missing psycopg2 wheel on Python 3.14
- **Date**: 2026-09-30
- **Symptoms**: `pip install -r requirements.txt` failed with `Error: pg_config executable not found` when attempting to build `psycopg2-binary==2.9.10`.
- **Root Cause**: Host system Python is 3.14.7, which lacks pre-built binary wheels on PyPI for psycopg2-binary, and local `pg_config` was not installed.
- **Fix**: Re-created `.venv` targeting Python 3.12 via `uv venv .venv --python 3.12` and installed dependencies cleanly with pre-built binary wheels.
- **Regression Guardrail**: Always activate `.venv` with Python 3.12 or use `uv` when updating dependencies.

### ERR-002: Missing Go executable on system path
- **Date**: 2026-09-30
- **Symptoms**: `go: command not found` when attempting `go mod tidy` in `app/src/`.
- **Root Cause**: Go 1.22.5 was installed in user space at `~/.local/go_dist/go/bin/go` but not symlinked to `~/.local/bin`.
- **Fix**: Created symlinks `ln -sf ~/.local/go_dist/go/bin/go ~/.local/bin/go` and `gofmt`.
- **Regression Guardrail**: Verify `which go` points to `~/.local/bin/go`.

### ERR-003: Schema table dependency ordering error
- **Date**: 2026-09-30
- **Symptoms**: `psql` failed with `relation "communities" does not exist` when creating `users` table.
- **Root Cause**: `users` table declared a foreign key to `communities(community_id)`, but `communities` was defined later in `app/db/schema.sql`.
- **Fix**: Moved `CREATE TABLE communities` to the top before `CREATE TABLE users`.

### ERR-004: Premature connection close in user_roles.py
- **Date**: 2026-09-30
- **Symptoms**: `psycopg2.InterfaceError: connection already closed` on protected routes like `/connect`.
- **Root Cause**: Helper functions in `app/src/algo/auth/user_roles.py` called `mydb.close()`, which terminated the active pooled connection stored in Flask's `g.db`.
- **Fix**: Removed `mydb.close()` calls from `user_roles.py` so that connection lifecycle is cleanly managed by Flask's `teardown_appcontext`.

### ERR-005: Missing forgot_password endpoint and template path mismatch
- **Date**: 2026-09-30
- **Symptoms**: `werkzeug.routing.exceptions.BuildError: Could not build url for endpoint 'forgot_password'` and template not found for `auth/login.html`.
- **Root Cause**: `login.html` was referencing `url_for('forgot_password')` which lacked an application route, and `auth.py` was looking for templates under an nonexistent `auth/` subfolder.
- **Fix**: Registered `forgot_password` on Flask app and mapped template paths to `login.html` and `register.html`.
