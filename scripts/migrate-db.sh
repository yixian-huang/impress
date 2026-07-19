#!/bin/bash
# SQLite to PostgreSQL Migration Script Wrapper

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "=== SQLite to PostgreSQL Migration Wrapper ==="
echo ""

# Check required tools
if ! command -v go &> /dev/null; then
    echo "ERROR: Go is not installed or not in PATH"
    exit 1
fi

# Set default source DSN if not provided
if [ -z "$SOURCE_DB_DSN" ]; then
    SOURCE_DB_DSN="$PROJECT_ROOT/data/inkless.db"
    if [ ! -f "$SOURCE_DB_DSN" ]; then
        echo "WARNING: Source database not found at $SOURCE_DB_DSN"
        echo "         Set SOURCE_DB_DSN environment variable to specify custom path"
    fi
fi

# Validate target DSN is provided
if [ -z "$TARGET_DB_DSN" ]; then
    echo "ERROR: TARGET_DB_DSN environment variable is required"
    echo ""
    echo "Examples:"
    echo "  # Docker Compose PostgreSQL"
    echo "  export TARGET_DB_DSN=\"host=localhost user=inkless password=inkless_dev_password dbname=inkless port=5432 sslmode=disable\""
    echo ""
    echo "  # Production PostgreSQL"
    echo "  export TARGET_DB_DSN=\"host=prod-db.example.com user=inkless password=SECURE_PASSWORD dbname=inkless port=5432 sslmode=require\""
    echo ""
    exit 1
fi

# Create backup before migration
BACKUP_DIR="$PROJECT_ROOT/backups/$(date +%Y%m%d_%H%M%S)"
mkdir -p "$BACKUP_DIR"

if [ -f "$SOURCE_DB_DSN" ]; then
    echo "Creating backup of source database..."
    cp "$SOURCE_DB_DSN" "$BACKUP_DIR/inkless.db.backup"
    sqlite3 "$SOURCE_DB_DSN" ".dump" > "$BACKUP_DIR/inkless_dump.sql"
    echo "✓ Backup saved to $BACKUP_DIR"
    echo ""
fi

# Export environment variables for Go script
export SOURCE_DB_DSN
export TARGET_DB_DSN

# Run migration script (from backend module)
cd "$PROJECT_ROOT/backend"
go run ./scripts/migrate-sqlite-to-postgres.go

# Exit with same status as migration script
exit $?
