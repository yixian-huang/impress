# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

印迹官网 (Blotting Consultancy) is a bilingual (`zh`/`en`) React SPA with a Go/Gin/GORM CMS backend.
- **frontend/** — Vite + React SPA (pnpm workspace member)
- **backend/** — Go/Gin REST API with GORM ORM
- **docs/** — Architecture and API specs
- **scripts/** — Build, deploy, and long-running agent harness

Primary planning docs (read in this order):
1. `docs/development-plan.md` (execution plan for long-running agent)
2. `docs/api-spec.md` (REST API contract)
3. `docs/architecture.md` (layering and delivery architecture)
4. `docs/data-model.md` (page config and translation-state rules)

## Core Commands

Use `pnpm` only (monorepo with `pnpm-workspace.yaml`; `frontend/` is the sole workspace member). Root-level scripts delegate to frontend via `pnpm -C frontend`.

```bash
pnpm install
pnpm dev            # Vite dev server at http://localhost:3000 (proxies /uploads to :8088)
pnpm build          # Production build to frontend/out/
pnpm preview
pnpm lint           # ESLint on frontend/src
pnpm type-check     # tsc --noEmit against tsconfig.app.json
```

### Testing

Vitest is configured with `happy-dom` environment and `@testing-library/react`. Setup file: `frontend/src/test/setup.ts`. Test files: `src/**/*.test.{ts,tsx}`.

```bash
pnpm test           # vitest single run (alias for pnpm -C frontend test:run)
pnpm test:run       # same as above
cd frontend && pnpm test -- src/path/to/file.test.tsx   # run a single test file
```

### Backend

```bash
cd backend && go build -o server ./cmd/server/   # compile
cd backend && go test -v -race ./...             # run all Go tests
cd backend && go vet ./...                       # static analysis
```

Or use the Makefile from the repo root:

```bash
make dev-up         # pnpm install + build-backend + dev (SQLite, recommended)
make dev            # starts backend (:8088) + frontend (:3000); backend must be built
make build-backend  # compile Go binary
make check          # pnpm lint && pnpm type-check
make stop           # kill processes on :8088 and :3000
```

### Docker

```bash
docker compose up                                        # PostgreSQL + backend + frontend
docker compose -f docker-compose.sqlite.yml up           # SQLite variant
```

### Verification

Default verification command: `pnpm lint && pnpm type-check`.
CI quality gate (`.github/workflows/quality-gate.yml`) runs: lint, type-check, frontend tests, `go vet`, `go test -race`, and an integration smoke build.

### Long-running agent harness

```bash
pnpm agent:init
pnpm agent:run
pnpm agent:run -- --max-iterations 10 --model sonnet
```

## Architecture

### Frontend

**Stack:** React 19 + TypeScript + Vite 7 + Tailwind CSS 3 + React Router 7 + i18next + axios

**Provider hierarchy** (see `App.tsx`): `I18nextProvider > BrowserRouter > ThemeProvider > GlobalConfigProvider > AuthProvider > AppRoutes`

Key layers:
- **Routing:** `src/router/config.tsx` — all routes are lazy-loaded. Public pages at `/`, `/about`, etc. Admin panel under `/admin/*`. Dynamic CMS pages under `/p/*`.
- **Pages:** `src/pages/*/page.tsx` — one directory per route. Admin pages in `src/pages/admin/`.
- **API client:** `src/api/http.ts` defines an axios instance; domain modules (`src/api/pages.ts`, `src/api/articles.ts`, etc.) export typed fetch functions. `VITE_API_BASE_URL` env var sets the backend origin.
- **i18n:** `src/i18n/local/{zh,en}/common.ts` — translation resources. Locale detection via `i18next-browser-languagedetector`; `zh` is the fallback for display.
- **Contexts:** `AuthContext` (JWT auth with refresh), `GlobalConfigContext` (fetches published site-wide config per locale).
- **Theme system:** `src/theme/` — CMS-driven dynamic page rendering. `DynamicPage.tsx` fetches a page config by slug from `/public/pages/:slug` and renders its `sections` array via `SectionRenderer`. Theme packages in `src/theme/packages/` (default, modern-dark, warm-earth). Section components in `src/theme/sections/` (HeroSection, CardGridSection, ContactFormSection, etc.).
- **Shared UI:** `src/components/feature/` (Header, Footer, PageHero, ProtectedRoute).
- **Auto-imports:** `unplugin-auto-import` auto-imports React hooks, router helpers, and `useTranslation`/`Trans` — do not add redundant imports for these. Generated file: `frontend/auto-imports.d.ts` (do not edit).
- **Path alias:** `@` → `src` (configured in `vite.config.ts`).
- **Build output:** `frontend/out/` (do not edit).

### Backend

**Stack:** Go 1.24 + Gin + GORM (SQLite or PostgreSQL) + JWT auth

Layered architecture in `backend/`:
- `cmd/server/main.go` — entry point, wires dependencies and routes.
- `internal/handler/` — HTTP handlers grouped by domain (auth, content, article, media, page, public, theme, analytics, backup, auditlog, category, tag, sitemap).
- `internal/service/` — business logic (content_service, validation_service).
- `internal/repository/` — data access layer (interface + `_impl.go` pattern).
- `internal/model/` — GORM models (user, article, content_document, content_version, media, page, category, tag, etc.).
- `internal/middleware/` — JWT auth middleware, rate limiting.
- `internal/seed/` — database seeding.
- `internal/db/` — database connection setup.
- `internal/backup/` — backup functionality.
- `pkg/` — shared packages (apierror, audit, auth, config, logger, metrics).

The backend supports both SQLite (local dev via `DB_DSN=file:...`) and PostgreSQL (Docker/production).

## Project-Specific Constraints

- Keep frontend behavior stable while introducing config-driven rendering.
- Maintain bilingual behavior (`zh` fallback for runtime display).
- Backend env vars: `PORT`, `DB_DSN`, `JWT_SECRET`, `JWT_REFRESH_SECRET`, `ENV`, `UPLOAD_DIR`.
- Frontend build-time defines: `BASE_PATH`, `IS_PREVIEW`, `PROJECT_ID`, `VERSION_ID`, `READDY_AI_DOMAIN`.

## Coding Conventions

- TypeScript + functional React components (`*.tsx`)
- 2-space indentation, double quotes, semicolons
- Tailwind utility-first styles
- `@typescript-eslint/no-explicit-any` is turned off
- Go: standard `go fmt`; repository interface + `_impl.go` pattern
- Run `pnpm lint && pnpm type-check` before finishing significant changes

## Long-Running Agent Protocol

These rules are mandatory when work is driven by `scripts/long-agent.mjs` + `claude` CLI.

### State Files

- `.long-agent/feature_list.json` — long-term backlog.
- `.long-agent/state.json` — iteration/session status.
- `.long-agent/agent-progress.md` — running log.
- `.long-agent/reports/*.json` — per-iteration structured reports.

Do not delete these files during autonomous runs.

### Backlog Mutability Rules

In `.long-agent/feature_list.json`:
- Allowed change: `status` only (`todo`, `in_progress`, `done`, `blocked`, `needs_human`)
- Not allowed: changing `id`, `title`, `description`, `acceptance`, `priority`, `category`, `depends_on`
- Not allowed: reordering or deleting existing features

### Iteration Rules

Each iteration must:
1. Re-orient (`pwd`, `ls -la`, `git status`, inspect related files).
2. Pick one feature only (next incomplete item with dependencies satisfied).
3. Implement only that feature.
4. Run verification command (`pnpm lint && pnpm type-check` unless overridden).
5. Update that feature status based on result.
6. Append concise progress notes.

Status mapping:
- `done`: implemented and verification passes
- `blocked`: technical blocker identified
- `needs_human`: business decision/credentials/approval needed (stop further iterations)

### Commit Discipline

- One focused commit per completed feature.
- Commit message should include feature id, key changes, and verification result.
- Never include generated runtime artifacts from `.long-agent/` in commits.
