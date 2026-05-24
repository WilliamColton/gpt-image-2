# External Integrations

**Analysis Date:** 2026-05-24

## APIs and External Services

### OpenAI Image Generation API

- **What:** Primary AI image generation and editing service, accessed via OpenAI-compatible API endpoints
- **SDK/Client:** `github.com/openai/openai-go/v3` v3.34.0 (Go SDK) — used in `backend-go/service/openai.go`
- **API Calls:**
  - `/v1/images/generations` — Text-to-image generation (`callImagesGenerationsOnce`)
  - `/v1/images/edits` — Image-to-image editing with optional mask (`callImagesEditsOnce`)
- **Auth:** API key per endpoint, configured in `backend-go/config.json` as `apiEndpoints[].apiKey`
- **Model:** Configurable via `backend-go/config.json` `model` field (default: `gpt-image-2`)
- **Configuration:** Admin dashboard can manage endpoint pool at runtime via `PUT /api/admin/config/endpoints`

### Multi-Endpoint Failover System

- **What:** Built-in failover across multiple OpenAI-compatible API endpoints with concurrency slots
- **Files:** `backend-go/service/openai.go` (failover logic), `backend-go/service/queue.go` (concurrency slot management)
- **Mechanism:**
  1. `withFailover()` iterates endpoints in priority order
  2. `AcquireSlotFrom()` blocks until a concurrency slot is free
  3. On failure, moves to the next endpoint (last error returned if all fail)
  4. Each endpoint can have a `maxConcurrency` limit (0 = unlimited)
- **Pricing:** Each endpoint has a `costPerImageX10000` for billing attribution (`UnitCostX10000`)
- **Global sale price:** `salePriceX10000` in config — profit = salePrice - costPerImage

### Codex CLI Compatibility

- **What:** Compatibility mode for Codex CLI-based API endpoints where the `n` parameter is non-functional
- **Files:** `backend-go/service/openai.go` (`codexPrompt()`, `CallImagesGenerationsConcurrent()`)
- **Behavior:**
  - Wraps prompt with `"Use the following text as the complete prompt. Do not rewrite it:"` prefix
  - For multi-image generation with Codex CLI: fires `n` concurrent single-image requests instead of using the `n` parameter
  - For single-image: prepends the anti-rewrite prefix only
- **Configuration:** `codexCli` boolean in `backend-go/config.json`; frontend can toggle via settings

## Data Storage

### SQLite Database

- **Provider:** File-based SQLite via `gorm.io/driver/sqlite` 1.6.0 and GORM 1.30.0
- **Location:** `backend-go/data/app.sqlite` (created on first run)
- **Journal Mode:** WAL mode with foreign keys enabled (`?_journal_mode=WAL&_foreign_keys=ON`)
- **Connection Limit:** Single connection (`SetMaxOpenConns(1)`) — appropriate for SQLite embedded use
- **Tables:** `users`, `redemption_codes`, `images`, `tasks`, `announcements`, `feedbacks`, `changelog_entries`, `billing_records`
- **Initialization:** `database.Init()` in `backend-go/database/database.go` — auto-migrates all models, seeds default admin and announcement records

### File Storage

- **Type:** Local filesystem
- **Location:** `backend-go/upload/` directory (configured via `config.UploadDir`)
- **Usage:** Generated and uploaded images are saved as files on disk
- **Frontend caching:** Browser IndexedDB via `src/lib/db.ts` — stores image data URLs locally using SHA-256 hashing for deduplication; database name: `gpt-image-playground`
- **Service Worker caching:** Offline-first caching via `public/sw.js` — caches app shell (HTML, manifest, icon) and statically-requested assets; API calls (`/api/*`) are NOT cached

### Caching

- **Frontend:** In-memory `Map<string, string>` image cache in `src/store.ts` (`imageCache`); IndexedDB for persistent image storage
- **Backend:** No external cache; SQLite with WAL mode provides acceptable read concurrency

## Authentication and Identity

### JWT-Based Authentication

- **Auth Provider:** Custom — JWT tokens signed with HMAC-SHA256 via `github.com/golang-jwt/jwt/v5`
- **Implementation:** `backend-go/service/auth.go` (`SignToken()`, `VerifyToken()`)
- **Token lifetime:** 30 days (`30 * 24 * time.Hour`)
- **Token storage:** Browser `localStorage` under key `gpt-image-playground-token` (`src/lib/backendApi.ts`)
- **Token transmission:** Bearer token in `Authorization` header, or `?token=` query parameter for image URLs

### Authentication Methods

1. **Exchange code login** (`POST /api/auth/login`): Creates new user or logs in existing user via redemption code
2. **Password login** (`POST /api/auth/login-password`): Username + password authentication with bcrypt
3. **Registration** (`POST /api/auth/register`): Username + password + optional invite code
4. **Account migration** (`POST /api/auth/migrate`): Sets username/password for legacy code-only users

### Password Security

- **Hashing:** bcrypt via `golang.org/x/crypto/bcrypt` with default cost
- **Files:** `backend-go/service/auth.go` (`hashPassword()`, `checkPassword()`)

### Admin Authentication

- **Admin login:** `POST /api/admin/login` — uses `adminApikey` from config
- **Admin middleware:** `AdminMiddleware()` in `backend-go/middleware/middleware.go` — verifies JWT role is `"admin"`

### Invite System

- **Files:** `backend-go/service/auth.go` (invite logic), `backend-go/config/config.go` (invite config)
- **Flow:** Users set invite codes; new users register with invite codes; both inviter and invitee receive quota rewards
- **Configuration:** `inviteEnabled`, `inviteInviterReward`, `inviteInviteeReward`, `inviteDefaultQuota` in backend config
- **Admin endpoints:** `GET/PUT /api/admin/invite-config`, `GET /api/admin/invites`

## Monitoring and Observability

### Logging

- **Backend framework:** Go standard library `log/slog` structured logging
- **Configuration:** `backend-go/log/log.go` — text format for development, JSON format configurable for production
- **Output:** `os.Stdout` (stdout)
- **Level:** `slog.LevelInfo` as default
- **Middleware:** Request logging via `middleware.RequestLogger()` in `backend-go/middleware/logger.go`
- **Frontend:** Console-based (`console.error` for service worker registration failures); Sonner toast for user-facing notifications

### Error Tracking

- **External service:** None detected
- **Error handling:** Backend logs errors via `slog.Error()`; frontend shows errors via `sonnerToast.error()` and `Error` objects thrown from `request()` in `src/lib/backendApi.ts`

### Analytics (Built-in Billing Analytics)

- **Files:** `backend-go/service/analytics.go`, `backend-go/handler/admin.go`
- **Endpoints:** Admin-only billing analytics API (`GET /api/admin/analytics/*`)
  - `/summary` — Revenue, cost, profit totals
  - `/trend` — Daily-bucketed time series
  - `/endpoints` — Grouped by API endpoint
  - `/users` — Grouped by user
- **Time ranges:** `today`, `7d`, `30d`, `all`
- **Money scale:** All amounts in X10000 integer units (e.g., 10000 = 1 unit of currency)
- **No external analytics services** are integrated

### Health Check

- **Endpoint:** `GET /api/health` — returns `{"ok": true}` (`backend-go/handler/auth.go`)

## CI/CD and Deployment

### Hosting

- **Self-hosted:** Backend deployed at `http://43.133.38.194:3004` (dev proxy target in `dev-proxy.config.json`)
- **Frontend serving:** Static files from `dist/` directory (production build output)
- **PWA support:** Service worker (`public/sw.js`) enables offline access; Web App Manifest (`public/manifest.webmanifest`) enables install prompts

### CI Pipeline

- **External CI service:** None detected (no CI config files found in repository)

### Dev Proxy

- **File:** `dev-proxy.config.json` — development proxy configuration
- **Configuration:**
  - `enabled: true` — proxy active in dev mode
  - `prefix: "/api-proxy"` — routes matching this prefix are proxied
  - `target: "http://43.133.38.194:3004"` — proxy target
  - `changeOrigin: true`, `secure: false`
- **Implementation:** Loaded by `vite.config.ts` and exposed at build time as `__DEV_PROXY_CONFIG__`

## Environment Configuration

### Required Environment Variables

**Frontend:**
- `VITE_BACKEND_URL` — Backend API base URL (defaults to `http://localhost:3001` if not set)

**Backend:**
- No environment variables required — all configuration via `backend-go/config.json`

### Secrets Location

- Backend: `backend-go/config.json` (contains JWT secret, admin API key, OpenAI API keys per endpoint)
- Frontend: `localStorage` for JWT token (not a secret in cryptographic sense)
- `.gitignore` excludes `backend-go/config.json` and `dev-proxy.config.json` from version control

## Webhooks and Callbacks

### Incoming

- **None** — No webhook receiver endpoints are registered

### Outgoing

- **None** — No webhook callbacks are configured; all external communication is request/response to OpenAI-compatible APIs

## PWA Capabilities

### Service Worker

- **File:** `public/sw.js`
- **Activation:** Registered in production mode only (`src/main.tsx`) — unregistered in development
- **Caching strategy:** Cache-first for non-API GET requests; network-first for navigation (caching index.html fallback); API routes (`/api/*`) bypass cache entirely
- **Cache name:** `gpt-image-playground-v0.1.5`

### PWA Manifest

- **File:** `public/manifest.webmanifest`
- **Config:** Standalone display mode, SVG icon, Zinc-based dark theme (`#030712` background, `#111827` theme color)
- **App name:** "GPT Image Playground", short: "GPT Image"

### Viewport Guards

- **File:** `src/lib/viewport.ts`
- **Purpose:** Mobile viewport management to prevent unwanted zoom and layout shifts

## Browser Storage

### localStorage

- **Key `gpt-image-playground`:** Zustand persist middleware for settings, auth user, params, UI state (`src/store.ts` — partialized)
- **Key `gpt-image-playground-token`:** JWT authentication token

### IndexedDB

- **File:** `src/lib/db.ts`
- **Database name:** `gpt-image-playground` (version 1)
- **Object store:** `images` — keyed by SHA-256 hash of image data URL
- **Purpose:** Offline image caching and deduplication (prevents re-uploading identical images)

---

*Integration audit: 2026-05-24*
