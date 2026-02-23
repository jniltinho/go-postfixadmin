#!/bin/sh
set -e

# Wait for database if DB_HOST and DB_PORT are set
if [ -n "$DB_HOST" ] && [ -n "$DB_PORT" ]; then
    echo "Waiting for database at $DB_HOST:$DB_PORT..."
    while ! nc -z "$DB_HOST" "$DB_PORT"; do
      sleep 1
    done
    echo "Database is up!"
fi

# Run database migrations
echo "Running database migrations..."
./postfixadmin migrate

# Create initial admin if requested via environment variables
if [ -n "$ADMIN_EMAIL" ] && [ -n "$ADMIN_PASSWORD" ]; then
    echo "Checking/Creating initial admin: $ADMIN_EMAIL"
    # Note: admin command might fail if email already exists, which is fine for initial setup
    ./postfixadmin admin --add-superadmin "$ADMIN_EMAIL:$ADMIN_PASSWORD" || echo "Admin user check/creation finished."
fi

# Execute the main application
exec "$@"
