# SQLite to PostgreSQL Migration Playbook

## Overview

This playbook documents the migration strategy from SQLite (development) to PostgreSQL (production), including schema/data compatibility considerations, execution steps, validation checklist, and rollback procedures.

## Pre-Migration Considerations

### 1. Database Feature Differences

#### JSON Storage
- **SQLite**: Uses `TEXT` type with JSON validation functions
  - GORM tag: `gorm:"type:jsonb"` → stored as `TEXT`
  - Application handles JSON marshaling via `JSONMap.Value()` and `JSONMap.Scan()`
- **PostgreSQL**: Native `JSONB` type with indexing and query capabilities
  - GORM tag: `gorm:"type:jsonb"` → stored as `JSONB`
  - Same application-level JSON handling works transparently

**Compatibility**: ✅ The custom `JSONMap` type in `internal/model/content_document.go` and `internal/model/content_version.go` handles both databases transparently through GORM's driver abstraction.

#### Data Types Mapping
| Model Field | Go Type | SQLite Type | PostgreSQL Type |
|-------------|---------|-------------|-----------------|
| User.ID | uint | INTEGER | SERIAL (INT) |
| User.Username | string | TEXT | VARCHAR |
| User.PasswordHash | string | TEXT | VARCHAR |
| User.Role | string | TEXT | VARCHAR |
| User.CreatedAt | time.Time | DATETIME | TIMESTAMP |
| RefreshToken.Token | string | TEXT UNIQUE | VARCHAR UNIQUE |
| RefreshToken.ExpiresAt | time.Time | DATETIME | TIMESTAMP |
| ContentDocument.PageKey | PageKey (string) | TEXT PK | VARCHAR PK |
| ContentDocument.DraftConfig | JSONMap | TEXT | JSONB |
| ContentDocument.PublishedConfig | JSONMap | TEXT | JSONB |
| ContentVersion.Config | JSONMap | TEXT | JSONB |

**Compatibility**: ✅ GORM handles all type conversions automatically. No schema changes required.

#### Constraints and Indexes
- **Unique Constraints**: Both databases support unique constraints on columns
- **Foreign Keys**: Both support foreign key relationships
- **Composite Indexes**: PostgreSQL composite unique index `(PageKey, Version)` works identically to SQLite
- **Auto-increment IDs**: SQLite `INTEGER PRIMARY KEY AUTOINCREMENT` maps to PostgreSQL `SERIAL PRIMARY KEY`

**Compatibility**: ✅ All constraints defined in models are portable.

### 2. Connection DSN Format

#### SQLite DSN Examples
```bash
# File-based database
DB_DSN="data/inkless.db"
DB_DSN="file:./data/inkless.db?cache=shared&mode=rwc"

# In-memory database (testing)
DB_DSN=":memory:"
```

#### PostgreSQL DSN Examples
```bash
# Standard format (used in docker-compose)
DB_DSN="host=db user=inkless password=inkless_dev_password dbname=inkless port=5432 sslmode=disable TimeZone=Asia/Shanghai"

# Connection string format
DB_DSN="postgres://inkless:inkless_dev_password@db:5432/inkless?sslmode=disable&TimeZone=Asia/Shanghai"

# Production with SSL
DB_DSN="host=prod-db.example.com user=inkless password=SECURE_PASSWORD dbname=inkless port=5432 sslmode=require TimeZone=UTC"
```

**Implementation**: The `internal/db/db.go` Init function auto-detects database type from DSN prefix (`postgres` vs file path).

### 3. Transaction Isolation Differences

- **SQLite**: Default isolation level is `SERIALIZABLE` with single-writer model
- **PostgreSQL**: Default isolation level is `READ COMMITTED`

**Impact**: The application's optimistic locking pattern (draft version checks in `internal/service/content_publish.go`) works correctly in both databases. PostgreSQL's `READ COMMITTED` is sufficient for our workflow.

### 4. Performance Characteristics

- **SQLite**: Single-writer, local file access, excellent for read-heavy development
- **PostgreSQL**: Multi-client concurrent writes, network-based, better for production scale

**Recommendation**: SQLite for local development, PostgreSQL for staging/production.

---

## Migration Execution Plan

### Phase 1: Pre-Migration Checks

#### 1.1 Verify Current SQLite Schema

```bash
# Check SQLite database schema
sqlite3 data/inkless.db ".schema"

# Expected tables:
# - users
# - refresh_tokens
# - content_documents
# - content_versions
# - migration_history
```

#### 1.2 Export SQLite Data

```bash
# Dump data to SQL file
sqlite3 data/inkless.db ".dump" > sqlite_dump.sql

# Count records per table
sqlite3 data/inkless.db <<EOF
SELECT 'users', COUNT(*) FROM users
UNION ALL
SELECT 'refresh_tokens', COUNT(*) FROM refresh_tokens
UNION ALL
SELECT 'content_documents', COUNT(*) FROM content_documents
UNION ALL
SELECT 'content_versions', COUNT(*) FROM content_versions;
EOF
```

#### 1.3 Backup Existing Data

```bash
# Create timestamped backup
BACKUP_DIR="backups/$(date +%Y%m%d_%H%M%S)"
mkdir -p "$BACKUP_DIR"
cp data/inkless.db "$BACKUP_DIR/inkless.db.backup"
sqlite3 data/inkless.db ".dump" > "$BACKUP_DIR/inkless_dump.sql"
echo "Backup saved to $BACKUP_DIR"
```

### Phase 2: PostgreSQL Setup

#### 2.1 Launch PostgreSQL Instance

**Option A: Docker Compose (Recommended)**
```bash
# Use existing docker-compose.yml
docker compose up -d db

# Verify database is ready
docker compose logs db
docker compose exec db pg_isready -U inkless -d inkless
```

**Option B: Standalone Docker**
```bash
docker run -d \
  --name inkless-postgres \
  -e POSTGRES_DB=inkless \
  -e POSTGRES_USER=inkless \
  -e POSTGRES_PASSWORD=inkless_dev_password \
  -p 5432:5432 \
  -v inkless_postgres_data:/var/lib/postgresql/data \
  postgres:15-alpine
```

**Option C: Managed Cloud Service**
```bash
# Example: AWS RDS, Google Cloud SQL, Azure Database for PostgreSQL
# Obtain connection details from cloud provider
# Update DB_DSN with production credentials
```

#### 2.2 Initialize PostgreSQL Schema

The backend application automatically creates schema on first connection via GORM AutoMigrate.

```bash
# Set PostgreSQL DSN
export DB_DSN="host=localhost user=inkless password=inkless_dev_password dbname=inkless port=5432 sslmode=disable TimeZone=Asia/Shanghai"
export JWT_SECRET="dev_jwt_secret_change_in_production"
export JWT_REFRESH_SECRET="dev_jwt_refresh_secret_change_in_production"
export ENV="development"

# Run backend to trigger auto-migration
go run cmd/server/main.go

# Alternative: Run migration directly
# (requires adding a migration-only command or script)
```

Verify schema creation:

```bash
# Connect to PostgreSQL
docker compose exec db psql -U inkless -d inkless

# List tables
\dt

# Expected output:
# - users
# - refresh_tokens
# - content_documents
# - content_versions
# - migration_history

# Check specific schema
\d content_documents

# Expected columns:
# - page_key (varchar, PK)
# - draft_config (jsonb)
# - draft_version (integer)
# - published_config (jsonb)
# - published_version (integer)
# - updated_at (timestamp)

\q
```

### Phase 3: Data Migration

#### 3.1 Strategy Selection

**Strategy A: Application-Level Export/Import (Recommended)**

Pros:
- Uses existing repository layer
- Validates data through application logic
- Handles JSON serialization correctly
- Type-safe and tested

Cons:
- Requires custom migration script
- Slower for large datasets

**Strategy B: SQL Dump Conversion**

Pros:
- Faster for large datasets
- Direct database-to-database transfer

Cons:
- Requires manual SQL dialect translation
- Risk of JSON format incompatibility
- Bypasses application validation

**Recommendation**: Use Strategy A for initial migration. Datasets are small (< 10 pages, < 100 versions expected).

#### 3.2 Application-Level Migration Script

Create `scripts/migrate-sqlite-to-postgres.go`:

```go
package main

import (
  "context"
  "fmt"
  "log"
  "os"
  "time"

  "inkless-cms/internal/db"
  "inkless-cms/internal/model"
  "inkless-cms/internal/repository"
)

func main() {
  // Source: SQLite
  srcDSN := os.Getenv("SOURCE_DB_DSN")
  if srcDSN == "" {
    srcDSN = "data/inkless.db"
  }

  // Target: PostgreSQL
  targetDSN := os.Getenv("TARGET_DB_DSN")
  if targetDSN == "" {
    log.Fatal("TARGET_DB_DSN environment variable required")
  }

  // Connect to source SQLite
  srcDB, err := db.Init(db.InitOptions{DSN: srcDSN})
  if err != nil {
    log.Fatalf("Failed to connect to source database: %v", err)
  }
  defer srcDB.Close()

  // Connect to target PostgreSQL
  targetDB, err := db.Init(db.InitOptions{DSN: targetDSN})
  if err != nil {
    log.Fatalf("Failed to connect to target database: %v", err)
  }
  defer targetDB.Close()

  // Initialize repositories
  srcUserRepo := repository.NewUserRepository(srcDB)
  targetUserRepo := repository.NewUserRepository(targetDB)

  srcTokenRepo := repository.NewRefreshTokenRepository(srcDB)
  targetTokenRepo := repository.NewRefreshTokenRepository(targetDB)

  srcDocRepo := repository.NewContentDocumentRepository(srcDB)
  targetDocRepo := repository.NewContentDocumentRepository(targetDB)

  srcVersionRepo := repository.NewContentVersionRepository(srcDB)
  targetVersionRepo := repository.NewContentVersionRepository(targetDB)

  ctx := context.Background()

  // Migrate Users
  fmt.Println("Migrating users...")
  users, err := srcUserRepo.ListAll(ctx)
  if err != nil {
    log.Fatalf("Failed to fetch users: %v", err)
  }
  for _, user := range users {
    if err := targetUserRepo.Create(ctx, user); err != nil {
      log.Printf("Warning: Failed to migrate user %s: %v", user.Username, err)
    } else {
      fmt.Printf("  ✓ Migrated user: %s (ID: %d)\n", user.Username, user.ID)
    }
  }

  // Migrate Content Documents
  fmt.Println("\nMigrating content documents...")
  for _, pageKey := range model.ValidPageKeys {
    doc, err := srcDocRepo.GetByPageKey(ctx, pageKey)
    if err != nil {
      continue // Skip if document doesn't exist
    }
    if err := targetDocRepo.Upsert(ctx, doc); err != nil {
      log.Printf("Warning: Failed to migrate document %s: %v", pageKey, err)
    } else {
      fmt.Printf("  ✓ Migrated document: %s (draft v%d, published v%d)\n",
        pageKey, doc.DraftVersion, doc.PublishedVersion)
    }
  }

  // Migrate Content Versions
  fmt.Println("\nMigrating content versions...")
  for _, pageKey := range model.ValidPageKeys {
    versions, err := srcVersionRepo.ListByPageKey(ctx, pageKey, 1000, 0)
    if err != nil {
      continue
    }
    for _, version := range versions {
      if err := targetVersionRepo.Create(ctx, version); err != nil {
        log.Printf("Warning: Failed to migrate version %s v%d: %v", pageKey, version.Version, err)
      } else {
        fmt.Printf("  ✓ Migrated version: %s v%d\n", pageKey, version.Version)
      }
    }
  }

  // Note: Skip refresh_tokens migration (they expire quickly)
  fmt.Println("\nMigration complete!")
  fmt.Println("Note: Refresh tokens were not migrated. Users will need to re-login.")
}
```

#### 3.3 Execute Migration

```bash
# Set environment variables
export SOURCE_DB_DSN="data/inkless.db"
export TARGET_DB_DSN="host=localhost user=inkless password=inkless_dev_password dbname=inkless port=5432 sslmode=disable"

# Run migration script
go run scripts/migrate-sqlite-to-postgres.go

# Expected output:
# Migrating users...
#   ✓ Migrated user: admin (ID: 1)
#   ✓ Migrated user: editor (ID: 2)
#
# Migrating content documents...
#   ✓ Migrated document: home (draft v5, published v4)
#   ✓ Migrated document: about (draft v3, published v3)
#   ...
#
# Migrating content versions...
#   ✓ Migrated version: home v1
#   ✓ Migrated version: home v2
#   ...
#
# Migration complete!
```

### Phase 4: Validation

#### 4.1 Record Count Verification

```bash
# PostgreSQL record counts
docker compose exec db psql -U inkless -d inkless -c "
SELECT 'users' AS table_name, COUNT(*) AS count FROM users
UNION ALL
SELECT 'content_documents', COUNT(*) FROM content_documents
UNION ALL
SELECT 'content_versions', COUNT(*) FROM content_versions;
"

# Compare with SQLite counts from Phase 1.2
```

#### 4.2 JSON Data Integrity Check

```bash
# Verify JSONB data is queryable
docker compose exec db psql -U inkless -d inkless -c "
SELECT
  page_key,
  draft_version,
  published_version,
  jsonb_typeof(draft_config) AS draft_type,
  jsonb_typeof(published_config) AS published_type
FROM content_documents;
"

# Expected: All draft_type and published_type should be 'object'
```

#### 4.3 API Behavior Parity Validation

Run the following validation checklist:

**Auth Endpoints**
```bash
# Test login
curl -X POST http://localhost:8088/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# Expected: 200 OK with access_token and refresh_token

# Test /auth/me
curl http://localhost:8088/auth/me \
  -H "Authorization: Bearer <access_token>"

# Expected: 200 OK with user details
```

**Public Content Endpoints**
```bash
# Test published content retrieval
curl http://localhost:8088/public/content/home

# Expected: 200 OK with published config matching SQLite version
```

**Admin Content Endpoints** (requires authentication)
```bash
# Get draft content
curl http://localhost:8088/admin/content/home/draft \
  -H "Authorization: Bearer <access_token>"

# Expected: 200 OK with draft config and version matching SQLite

# Update draft
curl -X PUT http://localhost:8088/admin/content/home/draft \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -H "If-Match: <current_version>" \
  -d '{"zh":{"title":"测试"},"en":{"title":"Test"}}'

# Expected: 200 OK with incremented version

# Validate draft
curl -X POST http://localhost:8088/admin/content/home/validate \
  -H "Authorization: Bearer <access_token>"

# Expected: 200 OK with validation results

# Publish content
curl -X POST http://localhost:8088/admin/content/home/publish \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{"expectedDraftVersion":<current_version>}'

# Expected: 200 OK with new published version

# List versions
curl http://localhost:8088/admin/content/home/versions \
  -H "Authorization: Bearer <access_token>"

# Expected: 200 OK with version history matching SQLite
```

#### 4.4 Performance Baseline

```bash
# Measure public endpoint latency
for i in {1..100}; do
  curl -o /dev/null -s -w "%{time_total}\n" http://localhost:8088/public/content/home
done | awk '{sum+=$1; count++} END {print "Average:", sum/count, "seconds"}'

# Expected: < 50ms average (local Docker network)
```

### Phase 5: Cutover

#### 5.1 Application Configuration Update

Update `.env` or environment configuration:

```bash
# Old (SQLite)
DB_DSN="data/inkless.db"

# New (PostgreSQL - Docker Compose)
DB_DSN="host=db user=inkless password=inkless_dev_password dbname=inkless port=5432 sslmode=disable TimeZone=Asia/Shanghai"

# New (PostgreSQL - Production)
DB_DSN="host=prod-db.example.com user=inkless password=SECURE_PASSWORD dbname=inkless port=5432 sslmode=require TimeZone=UTC"
```

#### 5.2 Restart Backend Service

**Docker Compose**
```bash
docker compose restart backend

# Verify health
curl http://localhost:8088/health
```

**Standalone**
```bash
# Stop old process
pkill -f inkless-api

# Start with new DSN
export DB_DSN="host=localhost user=inkless password=inkless_dev_password dbname=inkless port=5432 sslmode=disable"
./inkless-api-linux-amd64
```

#### 5.3 Monitor Logs

```bash
# Docker Compose
docker compose logs -f backend

# Check for PostgreSQL connection success
# Expected log: "Database connection established" or similar
```

---

## Rollback Procedures

### Emergency Rollback to SQLite

If PostgreSQL migration encounters critical issues:

#### 1. Immediate Service Restoration

```bash
# Update environment to use SQLite DSN
export DB_DSN="data/inkless.db"

# Restart backend
docker compose restart backend
# OR
pkill -f inkless-api && ./inkless-api-linux-amd64
```

#### 2. Restore SQLite Data (if corrupted)

```bash
# Restore from backup
BACKUP_DIR="backups/YYYYMMDD_HHMMSS"  # Use appropriate timestamp
cp "$BACKUP_DIR/inkless.db.backup" data/inkless.db

# Verify restore
sqlite3 data/inkless.db "SELECT COUNT(*) FROM users;"
```

#### 3. Verify Service Recovery

```bash
# Health check
curl http://localhost:8088/health

# Login test
curl -X POST http://localhost:8088/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

### Partial Rollback (PostgreSQL to Fixed PostgreSQL)

If data in PostgreSQL is corrupted but PostgreSQL itself is viable:

#### 1. Drop and Recreate PostgreSQL Schema

```bash
# Connect to PostgreSQL
docker compose exec db psql -U inkless -d inkless

# Drop all tables
DROP TABLE IF EXISTS content_versions CASCADE;
DROP TABLE IF EXISTS content_documents CASCADE;
DROP TABLE IF EXISTS refresh_tokens CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS migration_history CASCADE;

\q
```

#### 2. Re-run Migration from Backup

```bash
# Ensure SQLite backup is intact
ls -lh backups/YYYYMMDD_HHMMSS/inkless.db.backup

# Use backup as source
export SOURCE_DB_DSN="file:backups/YYYYMMDD_HHMMSS/inkless.db.backup?mode=ro"
export TARGET_DB_DSN="host=localhost user=inkless password=inkless_dev_password dbname=inkless port=5432 sslmode=disable"

# Re-run migration
go run scripts/migrate-sqlite-to-postgres.go

# Validate (Phase 4 checklist)
```

---

## Production Migration Checklist

Before migrating production data:

- [ ] **Backup**: Create full backup of production SQLite database
- [ ] **Dry Run**: Test migration on staging environment with production data clone
- [ ] **Maintenance Window**: Schedule downtime or read-only period
- [ ] **Credentials**: Secure production PostgreSQL credentials (use secrets manager)
- [ ] **SSL/TLS**: Enable `sslmode=require` for production PostgreSQL DSN
- [ ] **Connection Pool**: Configure `MaxOpenConn`, `MaxIdleConn`, `MaxLifetime` in `internal/db/db.go` Init call
- [ ] **Monitoring**: Set up PostgreSQL monitoring (connections, query performance, disk usage)
- [ ] **Rollback Plan**: Document rollback trigger conditions and test rollback procedure
- [ ] **Data Validation**: Run full API behavior parity validation (Phase 4.3)
- [ ] **Performance Testing**: Load test public endpoints under production traffic patterns
- [ ] **User Communication**: Notify users of planned downtime and expected re-login
- [ ] **Post-Migration Audit**: Compare record counts, version histories, and audit logs

---

## Troubleshooting

### Issue: "connection refused" to PostgreSQL

**Symptoms**: Backend fails to start with `dial tcp [::1]:5432: connect: connection refused`

**Solutions**:
1. Verify PostgreSQL is running: `docker compose ps db`
2. Check PostgreSQL logs: `docker compose logs db`
3. Test direct connection: `psql -h localhost -U inkless -d inkless`
4. Ensure DSN uses correct host (`localhost` for host machine, `db` for Docker Compose internal network)

### Issue: JSON data appears as string in PostgreSQL

**Symptoms**: `draft_config` and `published_config` stored as TEXT instead of JSONB

**Solutions**:
1. Check GORM tags in models: `gorm:"type:jsonb"`
2. Verify PostgreSQL detected: DSN must start with `postgres`
3. Drop tables and re-run AutoMigrate
4. Check PostgreSQL version: JSONB requires PostgreSQL 9.4+

### Issue: Migration script fails with "duplicate key value"

**Symptoms**: `ERROR: duplicate key value violates unique constraint "users_pkey"`

**Solutions**:
1. PostgreSQL database was not empty before migration
2. Drop all tables (see Partial Rollback section)
3. Re-run migration script

### Issue: Version numbers don't match after migration

**Symptoms**: Published version in PostgreSQL differs from SQLite

**Solutions**:
1. Check migration script completed without errors
2. Verify transaction boundaries in migration script
3. Re-run validation queries from Phase 4.2
4. If data is inconsistent, perform partial rollback and re-migrate

### Issue: Performance degradation after migration

**Symptoms**: Slow query response times in PostgreSQL vs SQLite

**Solutions**:
1. Run `ANALYZE` on all tables: `docker compose exec db psql -U inkless -d inkless -c "ANALYZE;"`
2. Check query plans: Add indexes if full table scans detected
3. Review connection pool settings in `internal/db/db.go`
4. Consider adding JSONB GIN indexes for complex queries (if needed in future)

---

## Appendix: Docker Compose Commands

```bash
# Start all services
docker compose up -d

# Stop all services
docker compose down

# Reset PostgreSQL data (WARNING: destructive)
docker compose down -v
docker volume rm inkless_postgres_data

# View logs
docker compose logs -f backend
docker compose logs -f db

# Execute commands in containers
docker compose exec backend /bin/sh
docker compose exec db psql -U inkless -d inkless

# Backup PostgreSQL (production)
docker compose exec db pg_dump -U inkless -d inkless > postgres_backup.sql

# Restore PostgreSQL from backup
cat postgres_backup.sql | docker compose exec -T db psql -U inkless -d inkless
```

---

## Summary

This playbook provides a complete migration path from SQLite to PostgreSQL with:

✅ **Pre-checks**: Schema/data compatibility verified through GORM abstraction layer
✅ **Execution steps**: Detailed phase-by-phase migration with application-level data transfer
✅ **Rollback procedure**: Emergency rollback to SQLite and partial rollback scenarios documented
✅ **Validation checklist**: API behavior parity verification covering all endpoints
✅ **Production readiness**: Security, monitoring, and operational considerations included

The custom `JSONMap` type ensures transparent JSON/JSONB handling across both databases. GORM's AutoMigrate feature guarantees schema consistency. Application-level migration script leverages existing repository layer for type-safe data transfer.

**Recommendation**: Use SQLite for local development (fast, simple), PostgreSQL for staging/production (concurrent writes, scalability, JSONB query capabilities).
