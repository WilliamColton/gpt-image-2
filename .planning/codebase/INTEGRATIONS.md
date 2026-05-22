# External Integrations

**Analysis Date:** 2026-05-22

## APIs & External Services

**OpenAI-Compatible Image Generation APIs:**
- OpenAI-compatible Images API - Backend generates and edits images through the OpenAI SDK in `backend-go/service/openai.go`.
  - SDK/Client: `github.com/openai/openai-go/v3` imported by `backend-go/service/openai.go` and versioned in `backend-go/go.mod`.
  - Endpoints: `/v1/images/generations` via `client.Images.Generate` and `/v1/images/edits` via `client.Images.Edit` in `backend-go/service/openai.go`.
  - Auth: per-endpoint `apiEndpoints[].apiKey` stored in backend `config.json` and managed by admin endpoints in `backend-go/handler/admin.go`; values are secrets and must not be committed or documented.
  - Base URLs: per-endpoint `apiEndpoints[].baseUrl` loaded by `backend-go/config/config.go`; concrete values may exist in gitignored `backend-go/config.json` and are intentionally excluded from this map.
  - Failover: `backend-go/service/openai.go` calls endpoints in priority order from `config.GetEndpointPool()` and moves to the next endpoint after errors.
  - Concurrency: optional `apiEndpoints[].maxConcurrency` enforced per base URL in `backend-go/service/queue.go`.
  - Admin management: `GET /api/admin/config/endpoints` and `PUT /api/admin/config/endpoints` in `backend-go/main.go` and `backend-go/handler/admin.go` update the endpoint pool and persist it through `backend-go/config/config.go`.

**Frontend-to-Backend API:**
- Go backend API - The React SPA and admin dashboard call the Gin API for auth, config, image upload/download, task execution, announcements, feedback, changelog, and admin functions.
  - SDK/Client: browser `fetch` wrappers in `src/lib/backendApi.ts` and `src/admin/adminApi.ts`; direct task/image fallback calls also exist in `src/store.ts`.
  - Auth: Bearer JWT from `localStorage` keys `gpt-image-playground-token` in `src/lib/backendApi.ts` and `gpt-image-playground-admin-token` in `src/admin/adminApi.ts`.
  - Base URL: `import.meta.env.VITE_BACKEND_URL` with fallback `http://localhost:3001` in `src/lib/backendApi.ts`, `src/admin/adminApi.ts`, and `src/store.ts`.
  - Public endpoints: `/api/health`, `/api/announcement`, `/api/changelog/latest`, and `/api/changelog` registered in `backend-go/main.go`.
  - Auth endpoints: `/api/auth/login`, `/api/auth/me`, and `/api/auth/redeem` registered in `backend-go/main.go` and implemented by `backend-go/handler/auth.go`.
  - Task endpoints: `/api/generate`, `/api/edit`, `/api/tasks`, `/api/tasks/:id`, and `/api/tasks/:id/stream` registered in `backend-go/main.go` and consumed by `src/lib/backendApi.ts` and `src/store.ts`.
  - Image endpoints: `/api/images`, `/api/images/:id`, and image query-token access in `src/lib/backendApi.ts`, `src/store.ts`, and `backend-go/handler/images.go`.
  - Admin endpoints: `/api/admin/login`, `/api/admin/users`, `/api/admin/codes`, `/api/admin/config/endpoints`, `/api/admin/announcement`, `/api/admin/feedback`, and `/api/admin/changelog` registered in `backend-go/main.go` and consumed by `src/admin/adminApi.ts`.

**Local Development Proxy:**
- Vite dev proxy - Optional same-origin proxy for browser CORS avoidance during local development.
  - SDK/Client: Vite `server.proxy` configured in `vite.config.ts`; config normalization in `src/lib/devProxy.ts`.
  - Auth: forwards whatever headers the browser request sends; no separate proxy credential handling in `vite.config.ts`.
  - Config: `dev-proxy.config.json` contains `enabled`, `prefix`, `target`, `changeOrigin`, and `secure`; `.gitignore` excludes `dev-proxy.config.json`.
  - Scope: only active for `vite serve` because `vite.config.ts` loads it when `command === 'serve'`.

**Deployment and Hosting References:**
- Static frontend hosting - `README.md` documents Vercel, GitHub Pages, Nginx, Netlify, and GitHub Container Registry deployment options for static assets.
  - SDK/Client: not applicable; build output comes from `npm run build` in `package.json` and Vite configuration in `vite.config.ts`.
  - Auth: not applicable for static hosting itself; app auth uses backend JWT via `src/lib/backendApi.ts`.
  - Repository artifacts: no `vercel.json`, `.github/workflows/`, root `Dockerfile`, or `docker-compose*.yml` is detected in the current checkout.

## Data Storage

**Databases:**
- SQLite via GORM - Backend persistent relational store for users, redemption codes, images, tasks, announcements, feedback, and changelog entries.
  - Connection: `filepath.Join(config.App.DataDir, "app.sqlite")` in `backend-go/database/database.go`; no database env var is used.
  - Client: GORM with `gorm.io/driver/sqlite` in `backend-go/database/database.go`.
  - Tables: model definitions in `backend-go/database/models.go` and AutoMigrate in `backend-go/database/database.go`.
  - Runtime settings: WAL mode and foreign keys enabled in `backend-go/database/database.go`; max open connections set to 1 for SQLite compatibility.
- Browser IndexedDB - Client-side generated/uploaded image cache named `gpt-image-playground` with object store `images` in `src/lib/db.ts`.
  - Connection: browser `indexedDB.open(DB_NAME, DB_VERSION)` in `src/lib/db.ts`.
  - Client: native IndexedDB API in `src/lib/db.ts`; no IndexedDB wrapper package is used.
- Browser localStorage - Client-side persisted Zustand state and JWT storage.
  - Connection: Zustand `persist` name `gpt-image-playground` in `src/store.ts`; JWT keys in `src/lib/backendApi.ts` and `src/admin/adminApi.ts`.
  - Client: native localStorage through Zustand middleware and explicit token helper functions.

**File Storage:**
- Local filesystem only - Backend stores uploaded, mask, and generated image files under `config.App.UploadDir` in `backend-go/service/image.go`.
  - Path resolution: user-specific upload directories from `backend-go/util/paths.go` and safe read/delete paths in `backend-go/service/image.go`.
  - Metadata: image file path, MIME type, size, SHA-256, source, and creation time stored in SQLite model `database.Image` in `backend-go/database/models.go`.
  - Public access: authenticated `GET /api/images/:id` streams local files from `backend-go/handler/images.go`; frontend appends JWT query tokens in `src/lib/backendApi.ts` and `src/store.ts`.

**Caching:**
- Service worker Cache API - Static app shell cache in `public/sw.js`, registered only in production by `src/main.tsx`.
- Browser memory cache - `imageCache` and `imageContentFetches` Maps in `src/store.ts` cache data URLs and in-flight image fetches.
- Browser IndexedDB image cache - `src/lib/db.ts` persists image data URLs and hashes for reuse across sessions.
- External cache: None detected; no Redis, Memcached, CDN SDK, or hosted cache client appears in `package.json`, `backend-go/go.mod`, `src/`, or `backend-go/`.

## Authentication & Identity

**Auth Provider:**
- Custom JWT and redemption-code authentication - No external identity provider is detected.
  - Implementation: `backend-go/service/auth.go` signs HS256 tokens with `jwtSecret`, verifies tokens, creates users from redemption codes, supports quota redemption, and manages admin/user roles.
  - User login: `/api/auth/login` in `backend-go/handler/auth.go` accepts redemption codes and returns `{ token, user }`; frontend uses `loginWithCode` in `src/lib/backendApi.ts`.
  - Admin login: `/api/admin/login` in `backend-go/handler/admin.go` compares submitted admin API key to `config.App.AdminApikey` and returns an admin JWT.
  - Middleware: `backend-go/middleware/middleware.go` enforces Bearer JWTs, query-token image access, disabled-user rejection, and admin role checks.
  - Token storage: frontend user token stored under `gpt-image-playground-token` in `src/lib/backendApi.ts`; admin token stored under `gpt-image-playground-admin-token` in `src/admin/adminApi.ts`.
  - Secrets: `jwtSecret` and `adminApikey` are backend config secrets loaded by `backend-go/config/config.go`; `backend-go/config.json` is gitignored and must not be committed.

## Monitoring & Observability

**Error Tracking:**
- None detected - No Sentry, Datadog, OpenTelemetry, Prometheus, or hosted error tracking package appears in `package.json`, `backend-go/go.mod`, `src/`, or `backend-go/`.

**Logs:**
- Backend structured logs - Go `log/slog` initialized in `backend-go/log/log.go` and used in `backend-go/main.go`, `backend-go/handler/`, `backend-go/service/`, and `backend-go/middleware/`.
- Request logging - `backend-go/middleware/logger.go` logs method, path, status, latency, IP, and user ID for Gin requests.
- Gin default logs - `gin.Default()` in `backend-go/main.go` includes Gin's default logger and recovery middleware.
- Frontend diagnostics - Service worker registration failures are logged with `console.error` in `src/main.tsx`; user-facing notifications use Sonner through `src/components/Toast.tsx`.

## CI/CD & Deployment

**Hosting:**
- Frontend static hosting - Vite output from `dist/` is the deployable SPA artifact, documented in `README.md` and produced by `npm run build` in `package.json`.
- Backend self-hosted Go service - `backend-go/main.go` starts the API server and requires persistent `data/` and `upload/` directories relative to the Go process working directory.
- README-documented platforms - `README.md` references Vercel, GitHub Pages, GHCR Docker images, Nginx, Netlify, and self-hosted reverse proxies.
- Detected repository artifacts - no root `Dockerfile`, `docker-compose*.yml`, `.github/workflows/`, or `vercel.json` is present in the current checkout.

**CI Pipeline:**
- None detected - `.github/workflows/` is not present in the current checkout, and no other CI config file is detected at the repository root.

## Environment Configuration

**Required env vars:**
- `VITE_BACKEND_URL` - Optional but production-relevant frontend build variable used in `src/lib/backendApi.ts`, `src/admin/adminApi.ts`, and `src/store.ts`; defaults to `http://localhost:3001` when absent.
- Backend env vars: Not detected; `backend-go/config/config.go` reads `config.json` instead of `os.Getenv` for app settings.
- `VITE_DEFAULT_API_URL` - Documented in `README.md` for static deployments but not referenced by current source in `src/` or `vite.config.ts`.
- `API_URL` - Documented in `README.md` for Docker static image configuration but no root `Dockerfile` or source reference is detected in the current checkout.

**Backend config file keys:**
- `port` - Backend listen port loaded by `backend-go/config/config.go` and used in `backend-go/main.go`.
- `jwtSecret` - JWT signing secret loaded by `backend-go/config/config.go` and used by `backend-go/service/auth.go`.
- `adminApikey` - Admin login credential loaded by `backend-go/config/config.go` and checked by `backend-go/handler/admin.go`.
- `model` - OpenAI image model loaded by `backend-go/config/config.go` and used in `backend-go/service/openai.go`.
- `apiMode` - Public config field returned by `backend-go/handler/config.go`; current frontend `src/types.ts` supports `images`.
- `timeout` - Public config field returned by `backend-go/handler/config.go` and stored in frontend settings through `src/store.ts`.
- `codexCli` - Public config field returned by `backend-go/handler/config.go`; request-level behavior is passed through `src/store.ts`, `src/lib/backendApi.ts`, and `backend-go/handler/generate.go`.
- `apiEndpoints` - OpenAI-compatible endpoint pool loaded and persisted by `backend-go/config/config.go`, edited by `backend-go/handler/admin.go`, and consumed by `backend-go/service/openai.go`.

**Secrets location:**
- Backend secrets file - `backend-go/config.json` is present and excluded by `.gitignore`; it contains environment-specific backend configuration and must not be read into documentation, quoted, or committed.
- API endpoint credentials - `apiEndpoints[].apiKey` values are backend secrets managed through `backend-go/handler/admin.go` and persisted by `backend-go/config/config.go`.
- JWT tokens - Browser tokens are stored in localStorage by `src/lib/backendApi.ts` and `src/admin/adminApi.ts`; avoid logging or exposing those values in UI changes.
- Dev proxy config - `dev-proxy.config.json` is gitignored and used only by `vite.config.ts` during development; treat environment-specific target values as local configuration.

## Webhooks & Callbacks

**Incoming:**
- Webhooks: None detected - No webhook-specific routes or signature verification handlers appear in `backend-go/main.go`, `backend-go/handler/`, or `src/`.
- Server-Sent Events: `GET /api/tasks/:id/stream` in `backend-go/main.go` and `backend-go/handler/tasks.go` streams task status updates to `streamTaskStatus` in `src/lib/backendApi.ts`; this is an internal app callback channel, not a third-party webhook.
- Public callbacks: None detected beyond normal public GET endpoints `/api/announcement`, `/api/changelog/latest`, and `/api/changelog` in `backend-go/main.go`.

**Outgoing:**
- OpenAI-compatible HTTPS API calls - `backend-go/service/openai.go` sends image generation/edit requests to configured `apiEndpoints[].baseUrl` values with `apiEndpoints[].apiKey` credentials.
- Outgoing webhooks: None detected - no code posts callbacks to third-party webhook URLs in `backend-go/` or `src/`.
- Browser network calls - Frontend `fetch` calls target the configured backend base URL in `src/lib/backendApi.ts`, `src/admin/adminApi.ts`, and `src/store.ts`; service worker `public/sw.js` ignores `/api/` requests and does not proxy external services.

---

*Integration audit: 2026-05-22*
