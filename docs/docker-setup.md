# Docker Compose Development Environment

This document provides instructions for running the Inkless CMS (Inkless CMS) application stack using Docker Compose.

## Overview

The default Docker Compose stack (`docker-compose.yml`) includes three services:

- **db**: PostgreSQL 15 database
- **backend**: Go/Gin/GORM REST API server
- **frontend**: React/Vite SPA development server

All services are connected via a Docker bridge network with health checks and automatic restart policies.

For lightweight local validation or single-host deployment, use `docker-compose.sqlite.yml` (no PostgreSQL container, backend uses SQLite file storage).

## Prerequisites

- Docker Desktop 20.10+ (includes Docker Compose V2)
- At least 4GB RAM allocated to Docker
- Ports 3000, 5432, and 8088 available on your host machine

## First-Time Setup

### 1. Copy Environment Configuration

```bash
cp .env.example .env
```

Review and customize `.env` if needed. The default values are suitable for local development.

### 2. Build and Start Services

```bash
docker-compose up --build
```

### Lightweight Mode (SQLite, no PostgreSQL)

```bash
docker compose -f docker-compose.sqlite.yml up --build
```

This mode runs only:
- `backend` (SQLite database at `/app/data/inkless.db`)
- `frontend`

It is suitable for local demos, low-traffic deployments, or environments where running PostgreSQL is unnecessary.

This command will:
- Build backend and frontend Docker images
- Pull the PostgreSQL 15 Alpine image
- Create a Docker network and persistent volume for database data
- Start all services with dependency ordering (db → backend → frontend)

### 3. Verify Services

Once all services are running, verify health:

```bash
# Check service status
docker-compose ps

# Check backend health endpoint
curl http://localhost:8088/health

# Access frontend
open http://localhost:3000
```

Expected outputs:
- `docker-compose ps` shows all services as "running" with healthy status
- Backend health endpoint returns `{"status":"healthy","timestamp":"..."}`
- Frontend loads in browser at http://localhost:3000

### 4. View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f backend
docker-compose logs -f frontend
docker-compose logs -f db
```

## Daily Development Workflow

### Start Stack

```bash
docker-compose up
```

Or run in detached mode:

```bash
docker-compose up -d
```

### Stop Stack

```bash
docker-compose down
```

### Rebuild After Code Changes

Backend changes require rebuild:

```bash
docker-compose up --build backend
```

Frontend changes are hot-reloaded automatically via volume mounts (no rebuild needed).

### Reset Database

To wipe database and start fresh:

```bash
docker-compose down -v
docker-compose up --build
```

The `-v` flag removes named volumes, including `postgres_data`.

## Service Details

### Database (PostgreSQL)

- **Port**: 5432 (exposed to host)
- **Credentials**: See `.env` file
- **Persistent Volume**: `postgres_data` (survives `docker-compose down`)
- **Connection String**: `host=db user=inkless password=inkless_dev_password dbname=inkless port=5432 sslmode=disable TimeZone=Asia/Shanghai`

To connect with `psql` from host:

```bash
psql -h localhost -U inkless -d inkless
# Password: inkless_dev_password
```

### Database (SQLite mode)

When using `docker-compose.sqlite.yml`:

- No separate DB service is required
- Backend DSN defaults to `file:/app/data/inkless.db?cache=shared&mode=rwc`
- Persistent volume: `sqlite_data`
- Reset database:

```bash
docker compose -f docker-compose.sqlite.yml down -v
docker compose -f docker-compose.sqlite.yml up --build
```

### Backend (Go API)

- **Port**: 8088 (exposed to host)
- **Health Check**: `GET /health`
- **Metrics**: `GET /metrics`
- **Source Mounts**: `cmd/`, `internal/`, `pkg/`, `go.mod`, `go.sum` mounted as volumes for live reload (requires manual restart)

To manually restart backend after Go code changes:

```bash
docker-compose restart backend
```

### Frontend (React/Vite)

- **Port**: 3000 (exposed to host)
- **Hot Module Reload**: Enabled via volume mounts
- **Source Mounts**: `src/`, `public/`, `index.html`, config files mounted as volumes
- **API Proxy**: Configured to proxy `/api` requests to backend (if needed)

Vite HMR will automatically reflect changes to `src/` files.

## Troubleshooting

### Port Already in Use

If you see "port is already allocated" errors:

```bash
# Check what's using the port (example: 8088)
lsof -i :8088

# Kill the process or stop conflicting containers
docker ps
docker stop <container_id>
```

### Database Connection Failures

If backend fails to connect to database:

1. Check database health: `docker-compose ps`
2. Verify database logs: `docker-compose logs db`
3. Ensure `depends_on` health check is working
4. Try restarting stack: `docker-compose down && docker-compose up`

### Frontend Not Loading

If frontend shows blank page:

1. Check frontend logs: `docker-compose logs frontend`
2. Verify Vite dev server started: look for "Local: http://0.0.0.0:3000"
3. Check browser console for errors
4. Ensure `node_modules` volume is not corrupted: `docker-compose down -v && docker-compose up --build`

### Backend Build Failures

If Go build fails:

1. Ensure `go.mod` and `go.sum` are up to date
2. Check Go version in `backend/Dockerfile` matches `backend/go.mod`
3. Rebuild with no cache: `docker-compose build --no-cache backend`

### Volume Permission Issues

On Linux, if you encounter permission errors:

```bash
# Fix volume permissions
sudo chown -R $USER:$USER .
```

### Service Won't Start (Dependency)

If services fail dependency health checks:

1. Increase health check timeout in `docker-compose.yml`
2. Check service logs for startup errors
3. Verify network connectivity: `docker network ls` and `docker network inspect inkless-network`

## Production Build (Optional)

To build production-ready images:

```bash
# Build frontend production image (from repo root, if Dockerfile.frontend exists)
# docker build -f Dockerfile.frontend --target production -t inkless-frontend:prod frontend

# Build backend production image
# docker build -t inkless-backend:prod backend

# Test production frontend
docker run -p 80:80 inkless-frontend:prod
```

## Cleanup

### Stop and Remove Containers

```bash
docker-compose down
```

### Remove Volumes (Database Data)

```bash
docker-compose down -v
```

### Remove Images

```bash
docker-compose down --rmi all
```

### Full Cleanup

```bash
docker-compose down -v --rmi all --remove-orphans
```

## Environment Variables

All environment variables are defined in `docker-compose.yml` and can be overridden via `.env` file.

Key variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 8088 | Backend API port |
| `DB_DSN` | (see .env.example) | Database DSN (PostgreSQL or SQLite) |
| `JWT_SECRET` | dev_jwt_secret_change_in_production | JWT signing secret |
| `JWT_REFRESH_SECRET` | dev_jwt_refresh_secret_change_in_production | Refresh token signing secret |
| `ENV` | development | Environment mode (development/production) |
| `VITE_API_BASE_URL` | http://localhost:8088 | Frontend API base URL |

⚠️ **Security Warning**: Never use the default JWT secrets in production. Generate strong secrets before deploying.

## Next Steps

- Review `docs/api-spec.md` for API endpoints
- Review `docs/architecture.md` for system design
- Use admin credentials from seed data (see `internal/seed/seeder.go`)
- Access admin UI at http://localhost:3000/admin (after implementing FE-101+)

## Support

For issues with Docker Compose setup:
1. Check this troubleshooting guide
2. Review service logs: `docker-compose logs`
3. Verify Docker Desktop resource limits
4. Ensure all prerequisites are met
