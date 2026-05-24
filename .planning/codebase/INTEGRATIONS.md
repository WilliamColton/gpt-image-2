# External Integrations

**Analysis Date:** 2026-05-24

## APIs & External Services

**OpenAI-compatible image APIs:**
- OpenAI Images API-compatible endpoints - used for text-to-image generation and image editing.
  - SDK/Client: `github.com/openai/openai-go/v3` configured in `backend-go/service/openai.go`.
  - Calls: `client.Images.Generate` and `client.Images.Edit` in `backend-go/service/openai.go`.
  - Auth: endpoint API keys are configured as `apiEndpoints[].apiKey` in `backend-go/config.json` and modeled by `backend-go/config/config.go`.
  - Base URL: endpoint `baseUrl` values are runtime-configurable through admin APIs in `backend-go/handler/admin.go` and frontend admin forms in `src/admin/AdminDashboard.tsx`.
  - Failover: endpoints are sorted by priority in `backend-go/config/config.go` and tried in order by `withFailover` in `backend-go/service/openai.go`.
  - Concurrency limits: per-endpoint `maxConcurrency` is enforced by `backend-go/service/queue.go`.
  - Billing attribution: successful generation captures endpoint base URL and cost snapshots in `backend-go/handler/generate.go`, `backend-go/service/billing.go`, and `backend-go/database/models.go`.

**Frontend-to-backend API:**
- Go backend REST API - used by the React app for auth, config, announcements, changelogs, feedback, image upload/download, task management, generation requests, and admin operations.
  - Client: browser `fetch` wrappers in `src/lib/backendApi.ts` and `src/admin/adminApi.ts`.
  - Auth: Bearer JWT from `localStorage` keys managed by `src/lib/backendApi.ts` and `src/admin/adminApi.ts`.
  - Base URL: optional `VITE_BACKEND_URL`; fallback is the local backend default in `src/lib/backendApi.ts`, `src/admin/adminApi.ts`, and `src/store.ts`.
  - Routes: registered in `backend-go/main.go` under `/api/**`.

**Server-Sent Events:**
- Task status streaming - used for generation progress/status updates.
  - Server endpoint: `GET /api/tasks/:id/stream` in `backend-go/main.go`, implemented by `TaskStream` in `backend-go/handler/tasks.go`.
  - Client: streaming `fetch` and manual event parsing in `src/lib/backendApi.ts`.
  - Auth: Bearer JWT header from `src/lib/backendApi.ts`; backend validates through `backend-go/middleware/middleware.go`.

**Local development proxy:**
- Vite dev proxy - optional local-only proxy for routing same-origin frontend requests to an OpenAI-compatible target during development.
  - Implementation: `vite.config.ts` reads `dev-proxy.config.json` and normalizes it with `src/lib/devProxy.ts`.
  - Auth: no proxy-level auth is implemented; credentials are handled by the proxied API request target if used.
  - Scope: only active for `vite serve`; not included in static production builds.

## Data Storage

**Databases:**
- SQLite local database
  - Connection: file path is built as `config.App.DataDir + /app.sqlite` in `backend-go/database/database.go`.
  - Client: GORM via `gorm.io/gorm` and `gorm.io/driver/sqlite` in `backend-go/database/database.go`.
  - Mode: SQLite WAL and foreign keys are enabled in the connection string in `backend-go/database/database.go`.
  - Connection pool: `SetMaxOpenConns(1)` in `backend-go/database/database.go`.
  - Schema: auto-migrated models in `backend-go/database/models.go` include users, redemption codes, images, tasks, announcements, feedback, changelog entries, and billing records.
  - Runtime files: `backend-go/data/` is ignored by `.gitignore` and should be persisted in production.
- Browser IndexedDB
  - Connection: `indexedDB.open('gpt-image-playground', 1)` in `src/lib/db.ts`.
  - Store: `images` object store in `src/lib/db.ts`.
  - Purpose: local image data URLs and deduplication by SHA-256/fallback hash in `src/lib/db.ts` and `src/store.ts`.

**File Storage:**
- Local filesystem only
  - Backend image files are written under `config.App.UploadDir` by `backend-go/service/image.go`.
  - Upload path safety is enforced by `ResolveUploadPath` in `backend-go/util/paths.go`.
  - Runtime directory creation is handled by `backend-go/util/paths.go` from `backend-go/main.go`.
  - `backend-go/upload/` is ignored by `.gitignore` and should be persisted in production.
- Browser Cache API
  - PWA app-shell and static GET responses are cached by `public/sw.js`.
  - API paths under `/api/` are explicitly excluded by `public/sw.js`.

**Caching:**
- In-memory frontend image cache in `src/store.ts` via `imageCache` and `imageContentFetches` maps.
- Browser IndexedDB persistent image cache in `src/lib/db.ts`.
- Browser Service Worker cache in `public/sw.js` for app shell/static assets.
- No Redis, Memcached, CDN SDK, or external cache service is detected.

## Authentication & Identity

**Auth Provider:**
- Custom first-party authentication
  - Implementation: JWT tokens are signed and verified with `github.com/golang-jwt/jwt/v5` in `backend-go/service/auth.go`.
  - Secret: `jwtSecret` in `backend-go/config.json`, modeled by `backend-go/config/config.go`.
  - User middleware: `AuthMiddleware` and `AdminMiddleware` in `backend-go/middleware/middleware.go` validate Bearer tokens.
  - Token transport: frontend APIs attach `Authorization: Bearer <token>` in `src/lib/backendApi.ts` and `src/admin/adminApi.ts`; image URLs can also pass `token` query parameter for `<img>`/file loading through `backend-go/middleware/middleware.go`.
  - User token storage: `gpt-image-playground-token` in browser `localStorage` via `src/lib/backendApi.ts`.
  - Admin token storage: `gpt-image-playground-admin-token` in browser `localStorage` via `src/admin/adminApi.ts`.
- Password authentication
  - Implementation: bcrypt hashing and comparison in `backend-go/service/auth.go` using `golang.org/x/crypto/bcrypt`.
  - Login/register/change-password/admin-reset flows are routed from `backend-go/main.go` to handlers in `backend-go/handler/auth.go` and `backend-go/handler/admin.go`.
- Redemption-code and invite-code identity flows
  - Redemption codes are modeled in `backend-go/database/models.go` and implemented in `backend-go/service/auth.go`.
  - Invite codes and invite reward settings are modeled in `backend-go/database/models.go`, configured in `backend-go/config/config.go`, and exposed through handlers in `backend-go/handler/auth.go` and `backend-go/handler/admin.go`.
- Admin bootstrap
  - A default admin user is created by `initAdmin` in `backend-go/database/database.go`.
  - Admin login uses `adminApikey` from `backend-go/config.json` through `AdminLogin` in `backend-go/handler/admin.go`.

## Monitoring & Observability

**Error Tracking:**
- None detected. No Sentry, Datadog, Honeycomb, OpenTelemetry, Prometheus, or external error tracking SDK is present in `package.json`, `backend-go/go.mod`, or source imports.

**Logs:**
- Backend uses Go `log/slog` initialized by `backend-go/log/log.go` and called throughout `backend-go/handler/**`, `backend-go/service/**`, and `backend-go/database/**`.
- Request logging middleware is implemented in `backend-go/middleware/logger.go` and registered in `backend-go/main.go`.
- Gin default logger/recovery is also included by `gin.Default()` in `backend-go/main.go`.
- Frontend service worker registration errors are logged with `console.error` in `src/main.tsx`.

## CI/CD & Deployment

**Hosting:**
- Frontend static hosting is supported through Vite build output in `dist/`; `README.md` documents static deployment and hosted demo options.
- Backend hosting is a standalone Go HTTP service from `backend-go/main.go` that listens on `config.App.Port`.
- PWA hosting requires serving `public/manifest.webmanifest`, `public/pwa-icon.svg`, `public/sw.js`, and built Vite assets.
- No repository deployment manifest is detected for Docker, Vercel, Netlify, Cloudflare, systemd, Kubernetes, or compose in the current scan.

**CI Pipeline:**
- None detected. No `.github/workflows/**`, GitLab CI, CircleCI, or other CI configuration is present in the repository scan.

## Environment Configuration

**Required env vars:**
- No required runtime environment variables are detected in the source scan.
- Optional frontend build/runtime variable: `VITE_BACKEND_URL` used by `src/lib/backendApi.ts`, `src/admin/adminApi.ts`, and `src/store.ts`.
- Vite built-ins used: `import.meta.env.PROD` and `import.meta.env.BASE_URL` in `src/main.tsx`.
- Backend configuration is loaded from `backend-go/config.json`, not environment variables, by `backend-go/config/config.go`.

**Required backend config fields:**
- `jwtSecret` - JWT signing secret used by `backend-go/service/auth.go` and `backend-go/middleware/middleware.go`.
- `adminApikey` - admin login secret used by `backend-go/handler/admin.go`.
- `apiEndpoints[].baseUrl` - OpenAI-compatible endpoint base URL used by `backend-go/service/openai.go`.
- `apiEndpoints[].apiKey` - OpenAI-compatible endpoint API key used by `backend-go/service/openai.go`.
- `apiEndpoints[].maxConcurrency` - optional endpoint concurrency limit enforced by `backend-go/service/queue.go`.
- `apiEndpoints[].priority` - optional endpoint ordering field sorted by `backend-go/config/config.go`.
- `apiEndpoints[].costPerImageX10000` and `salePriceX10000` - billing inputs used by `backend-go/handler/generate.go` and `backend-go/service/billing.go`.
- `model`, `apiMode`, `timeout`, `codexCli`, and invite settings - app behavior exposed by `backend-go/handler/config.go` and configured in `backend-go/config/config.go`.

**Secrets location:**
- Backend secrets and endpoint keys are stored in `backend-go/config.json`; `.gitignore` marks this file as ignored because it contains secrets.
- Local dev proxy configuration is stored in `dev-proxy.config.json`; `.gitignore` marks this file as ignored.
- No `.env`, `.env.*`, `*.env`, `.npmrc`, credential, certificate, or private key files are detected in the repository scan.

## Webhooks & Callbacks

**Incoming:**
- None detected. There are no webhook routes in `backend-go/main.go` or webhook handlers in `backend-go/handler/**`.
- Server-Sent Events are provided by `GET /api/tasks/:id/stream` in `backend-go/handler/tasks.go`; this is a client-initiated stream, not a webhook callback.

**Outgoing:**
- None detected. The backend makes outbound OpenAI-compatible SDK calls from `backend-go/service/openai.go`, but no outbound webhook/callback delivery is implemented.
- No payment, OAuth, email, object storage, or messaging provider callback integration is detected.

---

*Integration audit: 2026-05-24*
