<!-- refreshed: 2026-05-22 -->
# Architecture

**Analysis Date:** 2026-05-22

## System Overview

```text
┌─────────────────────────────────────────────────────────────┐
│                    Browser / PWA Shell                       │
│  `index.html` `public/sw.js` `src/main.tsx`                  │
├───────────────────────┬─────────────────────────────────────┤
│      User App          │              Admin App              │
│  `src/App.tsx`         │  `src/admin/AdminPage.tsx`          │
│  `src/components/*`    │  `src/admin/AdminDashboard.tsx`     │
└───────────┬───────────┴──────────────┬──────────────────────┘
            │                          │
            ▼                          ▼
┌─────────────────────────────────────────────────────────────┐
│                 Client State / Transport Layer               │
│  `src/store.ts` `src/lib/backendApi.ts`                      │
│  `src/admin/adminApi.ts` `src/lib/db.ts`                     │
└───────────────────────────┬─────────────────────────────────┘
                            │ HTTP + SSE `/api/*`
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                    Go API Gateway / Router                   │
│  `backend-go/main.go` `backend-go/middleware/*`              │
└───────────────────────────┬─────────────────────────────────┘
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                Handler + Service Domain Layer                │
│  `backend-go/handler/*` `backend-go/service/*`               │
└───────────────┬───────────────────────────────┬─────────────┘
                │                               │
                ▼                               ▼
┌─────────────────────────────┐   ┌───────────────────────────┐
│      Local Persistence       │   │      External Images API   │
│ `backend-go/database/*`      │   │ `backend-go/service/openai.go` │
│ `backend-go/data/app.sqlite` │   │ endpoint pool in           │
│ `backend-go/upload/`         │   │ `backend-go/config/config.go`  │
└─────────────────────────────┘   └───────────────────────────┘
```

## Component Responsibilities

| Component | Responsibility | File |
|-----------|----------------|------|
| Browser bootstrap | Installs mobile viewport guards, registers/unregisters the service worker, and chooses the user app or admin app based on `/admin` path. | `src/main.tsx` |
| User app shell | Initializes remote session/config/announcement/changelog data, applies theme, mounts global modals, and composes the main generation workspace. | `src/App.tsx` |
| Admin app shell | Applies the same persisted theme and switches between admin login and dashboard. | `src/admin/AdminPage.tsx` |
| Global client store | Owns user session, app settings, prompt/images/mask state, task list, filters, modal state, toast state, image cache orchestration, task submission, SSE fallback polling, and remote task mutations. | `src/store.ts` |
| User API client | Wraps `/api/auth`, `/api/config`, `/api/images`, `/api/tasks`, `/api/generate`, `/api/edit`, `/api/feedback`, announcement, and changelog endpoints; owns user token storage. | `src/lib/backendApi.ts` |
| Admin API client | Wraps `/api/admin/*` endpoints and owns admin token storage. | `src/admin/adminApi.ts` |
| Browser image persistence | Stores locally cached image data URLs in IndexedDB and hashes uploaded data URLs for deduplication. | `src/lib/db.ts` |
| User UI components | Render header, search/filter controls, task grid/cards, input bar, detail/lightbox/mask/settings/login/announcement/changelog/feedback modals, and confirm/toast surfaces. | `src/components/` |
| UI primitive library | Provides Radix/Tailwind primitives shared by user and admin screens. | `src/components/ui/` |
| Admin dashboard | Manages users, redemption codes, API endpoints, announcement, feedback, and changelog records. | `src/admin/AdminDashboard.tsx` |
| Backend bootstrap/router | Loads config, initializes logging/runtime directories/database, installs CORS/request logging, registers routes, and starts Gin. | `backend-go/main.go` |
| Backend middleware | Validates user/admin JWTs and logs request metadata. | `backend-go/middleware/middleware.go`, `backend-go/middleware/logger.go` |
| Backend handlers | Convert HTTP requests/responses to service calls for auth, images, tasks, generation, config, admin, announcement, feedback, and changelog features. | `backend-go/handler/` |
| Backend services | Own domain behavior for auth/quota, image persistence, task persistence, OpenAI image calls, endpoint queuing, announcements, feedback, changelog entries, and service DTOs. | `backend-go/service/` |
| Backend configuration | Loads `config.json`, exposes public app settings, owns endpoint pool sorting/persistence, and guards endpoint pool mutation with mutexes. | `backend-go/config/config.go` |
| Backend database | Opens SQLite through GORM, sets connection policy, runs migrations, and defines persistent models. | `backend-go/database/database.go`, `backend-go/database/models.go` |
| Backend utilities | Generate IDs, hash buffers, and constrain upload paths to the upload root. | `backend-go/util/` |
| PWA cache | Caches the app shell and same-origin GET assets while bypassing `/api/*`. | `public/sw.js` |

## Pattern Overview

**Overall:** Split frontend/backend monolith with a React/Zustand client, Gin HTTP API, service-oriented Go backend, SQLite persistence, and filesystem image blobs.

**Key Characteristics:**
- Keep app-wide browser state in `src/store.ts`; use components in `src/components/` for presentation and isolated local form state.
- Keep all browser backend calls behind `src/lib/backendApi.ts` or `src/admin/adminApi.ts`; stateful task/session side effects belong in `src/store.ts`.
- Keep backend routes centralized in `backend-go/main.go`; handlers in `backend-go/handler/` should call services in `backend-go/service/` rather than accessing external APIs directly.
- Store metadata in SQLite models from `backend-go/database/models.go`; store image bytes under `backend-go/upload/` through `backend-go/service/image.go`.
- Use endpoint pool failover and per-endpoint concurrency limits from `backend-go/config/config.go`, `backend-go/service/queue.go`, and `backend-go/service/openai.go` for image generation.

## Layers

**Browser/PWA shell:**
- Purpose: Own HTML host, PWA manifest/service worker, Tailwind globals, viewport/gesture guards, and top-level route selection.
- Location: `index.html`, `public/`, `src/main.tsx`, `src/index.css`
- Contains: Root DOM, service worker registration, app/admin route switch, global CSS, safe-area helpers, and animation utilities.
- Depends on: `react`, `react-dom`, `src/lib/viewport.ts`, `public/sw.js`
- Used by: User app in `src/App.tsx` and admin app in `src/admin/AdminPage.tsx`

**User application UI:**
- Purpose: Render the image generation workspace, task browsing, image detail/lightbox, settings, login, announcements, changelog, and feedback.
- Location: `src/App.tsx`, `src/components/`
- Contains: Feature components and modals such as `src/components/InputBar.tsx`, `src/components/TaskGrid.tsx`, `src/components/DetailModal.tsx`, `src/components/MaskEditorModal.tsx`, and `src/components/SettingsModal.tsx`.
- Depends on: `src/store.ts`, `src/types.ts`, `src/lib/*`, `src/components/ui/*`
- Used by: Browser bootstrap in `src/main.tsx`

**Admin UI:**
- Purpose: Manage users, redemption codes, endpoint config, announcements, feedback, and changelog entries.
- Location: `src/admin/`
- Contains: Route shell `src/admin/AdminPage.tsx`, login `src/admin/AdminLogin.tsx`, dashboard `src/admin/AdminDashboard.tsx`, and API client `src/admin/adminApi.ts`.
- Depends on: `src/store.ts` for theme/toast, `src/admin/adminApi.ts` for transport, shared primitives in `src/components/ui/`.
- Used by: Lazy route branch in `src/main.tsx`.

**Client state/orchestration:**
- Purpose: Coordinate session bootstrap, task lifecycle, image cache, persisted user preferences, modal state, and task mutations.
- Location: `src/store.ts`
- Contains: Zustand store, `imageCache`, remote image fetch deduplication, active SSE streams, polling fallback, `submitTask`, `executeTask`, task reuse/edit/delete helpers, and image add helpers.
- Depends on: `src/lib/backendApi.ts`, `src/lib/db.ts`, `src/lib/canvasImage.ts`, `src/lib/mask.ts`, `src/lib/size.ts`, `src/types.ts`
- Used by: Most user components in `src/components/` and admin theme/toast consumers in `src/admin/`.

**Client transport and browser storage:**
- Purpose: Isolate HTTP request construction, token storage, IndexedDB image persistence, canvas/mask processing, clipboard, size formatting, and dev proxy normalization.
- Location: `src/lib/`, `src/admin/adminApi.ts`
- Contains: `src/lib/backendApi.ts`, `src/admin/adminApi.ts`, `src/lib/db.ts`, `src/lib/canvasImage.ts`, `src/lib/mask.ts`, `src/lib/maskPreprocess.ts`, `src/lib/viewportTransform.ts`, `src/lib/devProxy.ts`, `src/lib/clipboard.ts`, and `src/lib/size.ts`.
- Depends on: Browser APIs (`fetch`, `localStorage`, `indexedDB`, `canvas`, `crypto.subtle`) and shared types in `src/types.ts`.
- Used by: `src/store.ts`, feature components in `src/components/`, admin UI in `src/admin/`, and Vite config in `vite.config.ts`.

**Backend API gateway:**
- Purpose: Provide the `/api/*` HTTP surface, attach middleware, configure CORS, and route requests to handlers.
- Location: `backend-go/main.go`
- Contains: Route groups for health, public announcement/changelog, auth, config, images, tasks/SSE, generation/edit, feedback, and admin APIs.
- Depends on: `backend-go/config`, `backend-go/database`, `backend-go/handler`, `backend-go/middleware`, `backend-go/log`, `backend-go/util`, Gin, and CORS middleware.
- Used by: Frontend clients in `src/lib/backendApi.ts` and `src/admin/adminApi.ts`.

**Backend handlers:**
- Purpose: Bind/validate HTTP input, fetch authenticated user context, return JSON/files/SSE events, and delegate business logic to services.
- Location: `backend-go/handler/`
- Contains: `backend-go/handler/auth.go`, `backend-go/handler/images.go`, `backend-go/handler/tasks.go`, `backend-go/handler/generate.go`, `backend-go/handler/admin.go`, `backend-go/handler/config.go`, `backend-go/handler/announcement.go`, `backend-go/handler/feedback.go`, and `backend-go/handler/changelog.go`.
- Depends on: `backend-go/service`, `backend-go/middleware`, `backend-go/config`, Gin, and standard HTTP utilities.
- Used by: Route registrations in `backend-go/main.go`.

**Backend service/domain layer:**
- Purpose: Own application behavior independent of HTTP: auth/JWT/quota, task persistence, image files, OpenAI requests, endpoint failover, endpoint concurrency slots, announcement, feedback, changelog, and DTO conversion.
- Location: `backend-go/service/`
- Contains: `backend-go/service/auth.go`, `backend-go/service/image.go`, `backend-go/service/task.go`, `backend-go/service/openai.go`, `backend-go/service/queue.go`, `backend-go/service/announcement.go`, `backend-go/service/feedback.go`, `backend-go/service/changelog.go`, and `backend-go/service/models.go`.
- Depends on: `backend-go/database`, `backend-go/config`, `backend-go/util`, OpenAI Go SDK, GORM, JWT, and standard library concurrency primitives.
- Used by: Handlers in `backend-go/handler/` and middleware in `backend-go/middleware/middleware.go`.

**Persistence/runtime layer:**
- Purpose: Persist application metadata in SQLite and image bytes in filesystem uploads.
- Location: `backend-go/database/`, `backend-go/data/`, `backend-go/upload/`
- Contains: GORM initialization/migrations in `backend-go/database/database.go`, models in `backend-go/database/models.go`, SQLite DB at `backend-go/data/app.sqlite`, and per-user upload folders under `backend-go/upload/`.
- Depends on: `backend-go/config/config.go`, `backend-go/util/paths.go`, GORM SQLite driver.
- Used by: Services in `backend-go/service/`.

## Data Flow

### Primary Request Path

1. Browser bootstraps `src/main.tsx:7`, chooses `src/App.tsx` for non-admin routes at `src/main.tsx:23`, and mounts React at `src/main.tsx:35`.
2. `src/App.tsx:25` calls `initStore()`; `src/store.ts:521` fetches public announcement/changelog and `src/store.ts:529` bootstraps authenticated sessions when a user token exists.
3. User input enters `src/components/InputBar.tsx`; `src/components/InputBar.tsx:217` calls `submitTask()` from `src/store.ts`.
4. `src/store.ts:551` validates prompt/auth/mask state, `src/store.ts:616` creates an optimistic queued `TaskRecord`, `src/store.ts:640` uploads mask images, `src/store.ts:657` uploads input images, and `src/store.ts:669` executes the task.
5. `src/store.ts:693` chooses generate/edit mode; `src/store.ts:702` calls `submitEditTask()` or `src/store.ts:704` calls `submitGenerateTask()` from `src/lib/backendApi.ts`.
6. `src/lib/backendApi.ts:161` posts `/api/generate`; `src/lib/backendApi.ts:169` posts `/api/edit`; both endpoints are registered in `backend-go/main.go:75`.
7. `backend-go/handler/generate.go:24` binds the request, `backend-go/handler/generate.go:38` checks quota, `backend-go/handler/generate.go:62` persists the queued task, and `backend-go/handler/generate.go:68` starts asynchronous generation in a goroutine.
8. `backend-go/handler/generate.go:73` loads endpoint config at `backend-go/handler/generate.go:76`, resolves input/mask image files at `backend-go/handler/generate.go:82`, and dispatches generation/edit calls at `backend-go/handler/generate.go:128`.
9. `backend-go/service/openai.go:140` and `backend-go/service/openai.go:204` call the configured Images API endpoints through failover; `backend-go/service/queue.go:77` gates endpoint concurrency.
10. `backend-go/handler/generate.go:145` saves generated images through `backend-go/service/image.go`, updates quota at `backend-go/handler/generate.go:149`, and writes final task metadata at `backend-go/handler/generate.go:167`.
11. `src/store.ts:708` starts SSE with `src/lib/backendApi.ts:182`; `backend-go/handler/tasks.go:121` streams task updates until completion or error.
12. `src/store.ts:672` warms output image cache on success, updates the task list, and displays toast state consumed by `src/components/Toast.tsx`.

### Session and Auth Flow

1. `src/components/LoginModal.tsx:18` calls `loginWithCode()` from `src/lib/backendApi.ts`.
2. `src/lib/backendApi.ts:62` posts `/api/auth/login`, which is registered in `backend-go/main.go:55`.
3. `backend-go/handler/auth.go:17` validates the code and delegates to `backend-go/service/auth.go:82`.
4. `backend-go/service/auth.go:82` either logs in an existing code owner or creates a new user/redeems the code, then `backend-go/service/auth.go:145` signs a JWT.
5. `src/lib/backendApi.ts:19` stores the user token in `localStorage`; `src/components/LoginModal.tsx:19` calls `bootstrapBackendSession()` from `src/store.ts`.
6. `src/store.ts:447` fetches `/api/auth/me`, `/api/tasks`, and `/api/config/public`; `backend-go/middleware/middleware.go:14` protects these endpoints with JWT validation.

### Image Cache and Storage Flow

1. User file drops/pastes enter `src/components/InputBar.tsx:172` and call `addImageFromFile()` at `src/store.ts:898`.
2. `src/store.ts:901` hashes the data URL through `src/lib/db.ts:48` and stores the data URL in the in-memory `imageCache` from `src/store.ts:43`.
3. On submit, `src/store.ts:657` uploads data URLs through `src/lib/backendApi.ts:124`; `backend-go/handler/images.go:15` receives multipart uploads.
4. `backend-go/service/image.go:42` deduplicates images by SHA-256 per user, writes bytes under `backend-go/upload/`, and creates `images` rows from `backend-go/database/models.go:27`.
5. Components call `ensureImageCached()` from `src/store.ts:131`; it checks IndexedDB through `src/lib/db.ts:38`, falls back to `/api/images/:id` via `src/store.ts:486`, converts blobs to data URLs at `src/store.ts:183`, and persists them with `src/lib/db.ts:42`.

### Admin Management Flow

1. `src/main.tsx:23` detects `/admin`; `src/main.tsx:26` lazy-loads `src/admin/AdminPage.tsx`.
2. `src/admin/AdminPage.tsx:8` reads admin login state from `src/admin/adminApi.ts:97`.
3. `src/admin/AdminLogin.tsx:31` posts admin credentials via `src/admin/adminApi.ts:70`; `backend-go/handler/admin.go:16` compares against backend config and signs an admin JWT.
4. `src/admin/AdminDashboard.tsx:158` loads tab-specific data through functions in `src/admin/adminApi.ts`.
5. Admin routes under `backend-go/main.go:83` require `backend-go/middleware/middleware.go:60`; service mutations update users/codes/config/announcement/feedback/changelog through `backend-go/service/*` and `backend-go/database/models.go`.
6. Endpoint edits call `backend-go/handler/admin.go:184`, persist through `backend-go/config/config.go:112`, refresh limiters through `backend-go/service/queue.go:96`, and return the sorted endpoint pool from `backend-go/config/config.go:27`.

### Announcement, Changelog, and Feedback Flow

1. `src/store.ts:521` fetches public announcement and latest changelog through `src/lib/backendApi.ts:86` and `src/lib/backendApi.ts:96`.
2. `src/App.tsx:42` auto-opens `src/components/ChangelogModal.tsx` when a published changelog has no persisted dismissal key.
3. `src/components/AnnouncementModal.tsx:12` renders public announcement content and persists the seen timestamp through `src/store.ts:262`.
4. `src/components/FeedbackModal.tsx:36` posts feedback through `src/lib/backendApi.ts:117`; `backend-go/handler/feedback.go:15` validates and stores feedback via `backend-go/service/feedback.go:50`.
5. Admin dashboard tabs in `src/admin/AdminDashboard.tsx` manage announcement, feedback status, and changelog records through `src/admin/adminApi.ts:143`, `src/admin/adminApi.ts:154`, and `src/admin/adminApi.ts:165`.

**State Management:**
- Use `src/store.ts` for shared client state, persisted preferences, session/user info, tasks, UI overlays, toasts, and task mutations.
- Use component-local `useState` for transient form state in `src/components/LoginModal.tsx`, `src/components/FeedbackModal.tsx`, `src/admin/AdminLogin.tsx`, and `src/admin/AdminDashboard.tsx`.
- Use `localStorage` only through `src/lib/backendApi.ts` and `src/admin/adminApi.ts` for user/admin JWT tokens and through Zustand persist in `src/store.ts:291` for selected app preferences.
- Use IndexedDB only through `src/lib/db.ts` for browser image data URL caching.
- Use SQLite/GORM only through `backend-go/database/` and service functions in `backend-go/service/`.

## Key Abstractions

**TaskRecord:**
- Purpose: Represents a generation/edit request from queued state through final outputs/error, including prompt, params, input/mask/output image IDs, timing, favorite state, and API-returned metadata.
- Examples: `src/types.ts:62`, `backend-go/service/models.go:67`, `backend-go/database/models.go:40`
- Pattern: Shared DTO shape between frontend and backend, with backend database JSON columns for nested params/image ID arrays in `backend-go/service/task.go:14`.

**TaskParams and AppSettings:**
- Purpose: Represent UI/default generation settings and server-returned public configuration.
- Examples: `src/types.ts:7`, `src/types.ts:25`, `backend-go/service/models.go:40`, `backend-go/service/models.go:58`, `backend-go/handler/config.go:12`
- Pattern: Frontend owns defaults in `src/types.ts`; backend exposes runtime values from `backend-go/config/config.go` through `/api/config/public`.

**Image identity and cache:**
- Purpose: Use image IDs as stable references across tasks, UI, IndexedDB, SQLite metadata, and filesystem upload paths.
- Examples: `src/types.ts:45`, `src/types.ts:90`, `src/store.ts:43`, `src/lib/db.ts:38`, `backend-go/database/models.go:27`, `backend-go/service/image.go:42`
- Pattern: In-memory cache first, IndexedDB second, authenticated backend image URL fallback, and backend SHA-256 deduplication per user.

**Endpoint pool:**
- Purpose: Manage one or more Images API endpoints with optional per-endpoint API key, concurrency limit, priority, and failover.
- Examples: `backend-go/config/config.go:12`, `backend-go/config/config.go:27`, `backend-go/service/queue.go:20`, `backend-go/service/openai.go:60`, `src/admin/adminApi.ts:25`
- Pattern: Admin-managed mutable backend config guarded by mutexes; service calls acquire slots and try endpoints in priority order.

**AuthUser and JWT:**
- Purpose: Represent authenticated users/admins and authorize access to user/admin API surfaces.
- Examples: `src/lib/backendApi.ts:6`, `backend-go/service/models.go:22`, `backend-go/service/auth.go:19`, `backend-go/middleware/middleware.go:14`, `backend-go/middleware/middleware.go:60`
- Pattern: User login uses redemption codes; admin login uses configured admin key; both receive JWTs validated by middleware.

**Modal and overlay state:**
- Purpose: Centralize app-wide overlay visibility while allowing local modal form state.
- Examples: `src/store.ts:252`, `src/store.ts:278`, `src/components/ConfirmDialog.tsx`, `src/components/Toast.tsx`, `src/hooks/useCloseOnEscape.ts`
- Pattern: Global store fields select which major overlay is open; `useCloseOnEscape` maintains a stack so Escape closes only the top registered modal.

**UI primitive layer:**
- Purpose: Provide reusable Tailwind/Radix components for consistent visual and accessibility behavior.
- Examples: `src/components/ui/button.tsx`, `src/components/ui/dialog.tsx`, `src/components/ui/card.tsx`, `src/components/ui/status-badge.tsx`, `src/lib/utils.ts`
- Pattern: Primitive components use `cn()` from `src/lib/utils.ts` to merge Tailwind classes and wrap Radix primitives where appropriate.

## Entry Points

**Frontend browser entry:**
- Location: `src/main.tsx`
- Triggers: Vite module script in `index.html:17`
- Responsibilities: Install mobile viewport guards, manage service worker lifecycle, branch between user and admin React trees, and mount `React.StrictMode`.

**User app entry:**
- Location: `src/App.tsx`
- Triggers: Non-admin route branch in `src/main.tsx:35`
- Responsibilities: Bootstrap store data, apply theme class, prevent page image drag, mount workspace components and global modals.

**Admin app entry:**
- Location: `src/admin/AdminPage.tsx`
- Triggers: `/admin` and `/admin/*` route branch in `src/main.tsx:23`
- Responsibilities: Apply theme, check admin token, switch login/dashboard, and clear admin token on logout.

**Service worker:**
- Location: `public/sw.js`
- Triggers: Production registration in `src/main.tsx:11`
- Responsibilities: Cache app shell, clean old caches, serve navigation fallback from cached `index.html`, and skip `/api/*`.

**Backend server:**
- Location: `backend-go/main.go`
- Triggers: Running the Go binary/module from `backend-go/`
- Responsibilities: Load config, initialize logging/dirs/database, register middleware/routes, and listen on configured port.

**Generation API:**
- Location: `backend-go/handler/generate.go`
- Triggers: `POST /api/generate` and `POST /api/edit` from `backend-go/main.go:75`
- Responsibilities: Validate task request, enforce quota, persist queued task, start generation goroutine, call Images API, save outputs, update task state.

**Task SSE API:**
- Location: `backend-go/handler/tasks.go`
- Triggers: `GET /api/tasks/:id/stream` from `backend-go/main.go:68`
- Responsibilities: Send current task state, poll SQLite every second, emit SSE events on status changes, and heartbeat every 15 seconds.

**Admin API:**
- Location: `backend-go/handler/admin.go`, `backend-go/handler/announcement.go`, `backend-go/handler/feedback.go`, `backend-go/handler/changelog.go`
- Triggers: `/api/admin/*` routes from `backend-go/main.go:81`
- Responsibilities: Authenticate admins, mutate users/codes/endpoint config/content records, and return JSON responses.

## Architectural Constraints

- **Threading:** Browser code in `src/` uses the single-threaded JS event loop with async `fetch`, timers, canvas work, and event listeners. Backend generation runs in goroutines from `backend-go/handler/generate.go:68`; endpoint concurrency is controlled by semaphores in `backend-go/service/queue.go:11`; SQLite is constrained to one open connection in `backend-go/database/database.go:25`.
- **Global state:** Frontend process-global state includes `imageCache` and fetch dedupe maps in `src/store.ts:43`, polling timer in `src/store.ts:52`, active SSE stream controllers in `src/store.ts:69`, and Escape handler stack in `src/hooks/useCloseOnEscape.ts:7`. Backend process-global state includes `config.App` in `backend-go/config/config.go:75`, endpoint pool mutex/state in `backend-go/config/config.go:20`, `database.DB` in `backend-go/database/database.go:15`, endpoint limiters in `backend-go/service/queue.go:15`, and default logger in `backend-go/log/log.go:8`.
- **Circular imports:** Static relative-import scan detects no TypeScript cycles under `src/`; Go package scan detects no package cycles under `backend-go/`. Keep dependencies directed from UI to store/lib and from handler/middleware to service/database/config.
- **Secret-bearing config:** `backend-go/config.json` is ignored and contains backend runtime secrets/API endpoints; never import it into frontend code or quote values from it in documentation. Public config must flow through `backend-go/handler/config.go` and `src/lib/backendApi.ts`.
- **Auth boundary:** User APIs use `backend-go/middleware/middleware.go:14`; admin APIs use `backend-go/middleware/middleware.go:60`; image GET also accepts a `token` query parameter for `<img>`/blob fetches through `src/lib/backendApi.ts:136` and `src/store.ts:486`.
- **CORS boundary:** Backend allows all origins and common headers/methods in `backend-go/main.go:41`; do not rely on CORS as an authorization control.
- **Runtime data:** `backend-go/data/`, `backend-go/upload/`, `dist/`, `node_modules/`, `dev-proxy.config.json`, and `backend-go/config.json` are not source-of-truth code directories.

## Anti-Patterns

### Stateful backend mutations in presentation components

**What happens:** Components can call transport functions directly for isolated form actions, such as `src/components/LoginModal.tsx:18`, `src/components/SettingsModal.tsx:31`, and `src/components/FeedbackModal.tsx:36`.
**Why it's wrong:** Duplicating shared task/session mutations in components bypasses image cache updates, persisted task state, SSE/polling fallback, and toast/confirm conventions centralized in `src/store.ts`.
**Do this instead:** Put shared app mutations in `src/store.ts` and keep components calling store actions. Use `src/lib/backendApi.ts` directly only for isolated form submits that do not mutate shared task/image state.

### New mutable globals outside existing state owners

**What happens:** The architecture already has process globals for store/cache/timers in `src/store.ts`, Escape stack in `src/hooks/useCloseOnEscape.ts`, backend config in `backend-go/config/config.go`, database handle in `backend-go/database/database.go`, and endpoint limiters in `backend-go/service/queue.go`.
**Why it's wrong:** Additional module-level mutable state makes tests and concurrent backend behavior harder to reason about, especially around generation goroutines and SSE/polling.
**Do this instead:** Add frontend shared state to `src/store.ts`; add backend config state to `backend-go/config/config.go`; add database-backed state to `backend-go/database/models.go` plus `backend-go/service/*`; add concurrency controls to `backend-go/service/queue.go`.

### Backend route behavior outside handler/service split

**What happens:** Routes are registered centrally in `backend-go/main.go`; handlers in `backend-go/handler/` perform HTTP binding and services in `backend-go/service/` perform business logic.
**Why it's wrong:** Adding OpenAI calls, file writes, or database mutations directly in `backend-go/main.go` or middleware would bypass reusable service logic and make authorization/error handling inconsistent.
**Do this instead:** Register only route wiring in `backend-go/main.go`, put request/response handling in `backend-go/handler/<feature>.go`, and put reusable domain logic in `backend-go/service/<feature>.go`.

## Error Handling

**Strategy:** Use thrown `Error` values and toast/UI state in the frontend; use structured logs plus JSON error responses and persistent task error state in the backend.

**Patterns:**
- Frontend API clients throw `Error(message)` on non-2xx responses in `src/lib/backendApi.ts:31` and `src/admin/adminApi.ts:48`; components catch and show store toasts or local form errors.
- `src/store.ts:707` prefers SSE updates and `src/store.ts:722` falls back to polling with `src/store.ts:55` when SSE fails.
- Backend handlers return `c.JSON(status, gin.H{"error": ...})` for validation/auth/service failures in files such as `backend-go/handler/auth.go`, `backend-go/handler/generate.go`, and `backend-go/handler/feedback.go`.
- Generation failures call `failTask()` in `backend-go/handler/generate.go:172`, update task `status = "error"`, persist `error`, `finishedAt`, and `elapsed`, and log through `slog`.
- File/image path resolution uses `backend-go/util/paths.go:29` to keep resolved upload paths inside `backend-go/upload/`.

## Cross-Cutting Concerns

**Logging:** Backend initializes the default `slog` logger in `backend-go/log/log.go:10` and attaches request logging middleware in `backend-go/middleware/logger.go:10`. Frontend logs are limited to service worker registration failure in `src/main.tsx:13` and clipboard/image copy exceptions in components such as `src/components/DetailModal.tsx:256`.

**Validation:** Frontend validates prompt/auth/mask readiness in `src/store.ts:551`, image count/type in `src/components/InputBar.tsx:172`, and mask dimensions/coverage in `src/lib/canvasImage.ts:56`. Backend validates JSON/body/path data in `backend-go/handler/*`, quota in `backend-go/service/auth.go:331`, endpoint config in `backend-go/handler/admin.go:184`, feedback text/status in `backend-go/handler/feedback.go:15`, and changelog lengths/published requirements in `backend-go/service/changelog.go:144`.

**Authentication:** User authentication uses redemption-code login in `backend-go/service/auth.go:82`, JWT signing in `backend-go/service/auth.go:19`, bearer/query-token verification in `backend-go/middleware/middleware.go:14`, and token storage in `src/lib/backendApi.ts:15`. Admin authentication uses configured admin key validation in `backend-go/handler/admin.go:16`, admin JWT verification in `backend-go/middleware/middleware.go:60`, and token storage in `src/admin/adminApi.ts:32`.

**Persistence:** Frontend user preferences are persisted by Zustand middleware in `src/store.ts:291`; browser images are cached in IndexedDB through `src/lib/db.ts`; backend metadata is stored through GORM models in `backend-go/database/models.go`; image bytes are stored through `backend-go/service/image.go` under `backend-go/upload/`.

**Styling and accessibility:** Use shared Radix/Tailwind primitives from `src/components/ui/`, utility class merging from `src/lib/utils.ts`, and global font/safe-area/animation definitions in `src/index.css`.

---

*Architecture analysis: 2026-05-22*
