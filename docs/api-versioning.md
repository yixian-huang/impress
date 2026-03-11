# API Versioning Strategy

## Current State
All routes use implicit v1 via path prefixes: `/public/*`, `/admin/*`, `/auth/*`.

## Strategy
- **No prefix change now**: Existing routes remain as-is (no `/api/v1/` prefix retrofit).
- **Plugin routes**: `GET/POST /api/v1/plugins/{pluginId}/*` (introduced in Phase 3).
- **Breaking changes**: If a v2 is ever needed, new routes at `/api/v2/*` while v1 remains supported for 2 major releases.
- **Additive changes**: New fields in responses are NOT breaking. Clients should ignore unknown fields.
- **Deprecation**: Deprecated fields get a `deprecated` note in Swagger docs for 1 release before removal.

## Response Envelope
Existing pattern (direct JSON) remains. No wrapper envelope.

## Versioning Headers
- `X-API-Version: 1` response header added to all API responses (optional, for client debugging).
