#!/bin/sh
# Run DB migrations then start the API server.
# Used by Render's startCommand so the DB is always up-to-date on deploy.
set -e

echo "[entrypoint] Running database migrations..."
migrate -path ./migrations -database "$DATABASE_URL" up
echo "[entrypoint] Migrations done."

exec ./app
