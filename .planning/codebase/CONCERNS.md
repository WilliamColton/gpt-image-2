# Codebase Concerns

**Analysis Date:** 2026-05-24

## Security Concerns

### Hardcoded Default Secrets in Config

- **Issue:** The backend config defaults contain hardcoded weak secrets that become live if `config.json` is missing.
- **Files:** `backend-go/config/config.go:90-91`
- **Code:** `JWTSecret: "change-me"`, `AdminApikey: "change-me-admin-apikey"`
- **Risk:** If `config.json` is absent, the backend starts with these defaults. Anyone who knows the defaults can forge JWTs and log into the admin dashboard.
- **Fix approach:** Require `config.json` to be present and exit with an error if missing, rather than falling back to unsafe defaults. Use `os.ReadFile` error check at line 99-101 -- currently it returns `nil` (no error) when the file is missing, meaning defaults are used.

### CORS: All Origins Allowed

- **Issue:** CORS is configured with `AllowAllOrigins: true` for all routes.
- **Files:** `backend-go/main.go:41-46`
- **Risk:** Any origin can make authenticated requests to the API, enabling CSRF attacks where malicious sites make requests on behalf of authenticated users.
- **Current mitigation:** CORS `AllowCredentials` is `false`, so browser credentials are not included. However, the API relies on `Authorization` header bearer tokens which are manually attached by JS, so this protection is partial at best.
- **Fix approach:** Restrict to known frontend origins only, or at minimum maintain a configurable allowlist.

### Token in Query String for Image URLs

- **Issue:** Image URLs include the JWT token as a query parameter.
- **Files:** `src/store.ts:484`, `src/lib/backendApi.ts:201-202`
- **Code:** `/api/images/${id}?token=${encodeURIComponent(token)}`
- **Risk:** JWT tokens in URLs are logged by proxies, CDNs, and browser histories. Token leakage risk.
- **Fix approach:** Use a short-lived signed image access token instead of the full JWT, or validate via Authorization header only with cookie-based image auth.

### No Input Sanitization on Backend Handlers

- **Issue:** Many `ShouldBindJSON` calls discard errors with `_`, meaning malformed JSON is silently accepted.
- **Files:** `backend-go/handler/auth.go:22,80,102,129,158,176,200`, `backend-go/handler/tasks.go:31`
- **Risk:** Silent binding failures can lead to nil pointer issues or unexpected behavior downstream.
- **Fix approach:** Check `ShouldBindJSON` errors consistently and return 400 Bad Request on malformed input.

### Password Validation: Minimum 8 Characters Only

- **Issue:** Password policy only enforces 8-character minimum; no complexity, no common-password check.
- **Files:** `backend-go/handler/auth.go:107-108`, `backend-go/service/auth.go:562-563`
- **Code:** `len(body.Password) < 8`
- **Risk:** Weak passwords like `password123` are accepted.
- **Fix approach:** Add minimum complexity requirements (uppercase, digit, symbol) or recommend a password strength meter on the client.

### `as string` / `as any` Type Assertions

- **Issue:** Unchecked type assertions used in production code (not just tests).
- **Files:** `src/components/InputBar.tsx:450` (`val as any`), `src/components/SearchBar.tsx:30` (`val as any`), `src/components/Select.tsx:16` (`value: any`)
- **Risk:** Runtime type mismatch can cause crashes without clear error messages.
- **Fix approach:** Define proper discriminated union types for Select values and use narrowing instead of `as any`.

### Public Config Exposes Admin-Critical Settings

- **Issue:** The public `/api/config/public` endpoint exposes `model`, `codexCli`, `apiMode`, `inviteEnabled` to unauthenticated users.
- **Files:** `backend-go/handler/config.go:12-21`
- **Risk:** Information disclosure -- exposes infrastructure details (model name, API mode) to anyone.
- **Fix approach:** Review which fields need to be public; consider removing `model` and `apiMode`.

## Performance Concerns

### SQLite with Single Connection

- **Issue:** SQLite is configured with `SetMaxOpenConns(1)` (single writer).
- **Files:** `backend-go/database/database.go:26`
- **Risk:** All database operations (reads + writes) serialize on a single connection. Under concurrent request load, this becomes a bottleneck.
- **Improvement path:** Use WAL mode with multiple readers (already enabled via `_journal_mode=WAL`). Increase `MaxOpenConns` to allow concurrent reads while writes are serialized by SQLite's built-in WAL semantics.

### SSE Polls Database Every 1 Second

- **Issue:** The Server-Sent Events (SSE) task stream endpoint polls the database every second while a task is running.
- **Files:** `backend-go/handler/tasks.go:162`
- **Code:** `pollTicker := time.NewTicker(1 * time.Second)`
- **Risk:** With many concurrent SSE connections (one per active task per user), the polling frequency could overwhelm the single-threaded SQLite connection.
- **Improvement path:** Consider reducing to 2-3 second intervals, or replace with an in-memory pub/sub for task status updates avoiding DB round-trips.

### Large Component Files

- **Issue:** Several React component files are large, indicating high complexity and potential for poor re-render behavior.
- **Files:**
  - `src/admin/AdminDashboard.tsx` -- 1,574 lines
  - `src/store.ts` -- 936 lines
  - `src/components/InputBar.tsx` -- 703 lines
  - `backend-go/service/auth.go` -- 723 lines
- **Risk:** God components/files are hard to maintain, test, and reason about. Large zustand stores can cause unnecessary re-renders across unrelated components.
- **Fix approach:** Split `AdminDashboard.tsx` into tab-specific sub-components. Break `store.ts` into domain-specific slices (auth, tasks, images, UI). Split `InputBar.tsx` into composable controls.

### No API Response Caching

- **Issue:** All frontend API calls use `cache: 'no-store'` and no ETag or cache-control headers are set on the backend.
- **Files:** `src/lib/backendApi.ts:41` (`cache: 'no-store'` on every `fetch`), `src/lib/backendApi.ts:145,152,162,173` (public endpoints also no-store)
- **Risk:** Repeated fetches of announcements, changelogs, and config cause unnecessary network and backend load.
- **Fix approach:** Add ETag/Last-Modified support on backend. Use `stale-while-revalidate` caching strategies for public endpoints. Remove `cache: 'no-store'` where unnecessary.

## Technical Debt

### Skipped Test Suite

- **Issue:** An entire test `describe` block for image cache behavior is skipped with a TODO comment.
- **Files:** `src/store.test.ts:275`
- **Code:** `describe.skip('image cache behavior in store — TODO: update for current store implementation', () => {`
- **Risk:** Image caching logic in `store.ts` has no test coverage, despite being critical for performance.
- **Fix approach:** Update the tests to match current store implementation and re-enable.

### `.bak` Backup Files Committed/Left in Source

- **Issue:** Two `.bak` backup files exist in the source tree from previous edits.
- **Files:** `src/admin/AdminDashboard.tsx.bak` (1,462 lines), `src/admin/AdminDashboard.tsx.bak2` (1,462 lines)
- **Risk:** Dead code and confusion about which file is canonical. These are identical copies of a previous version.
- **Fix approach:** Delete `.bak` and `.bak2` files. Ensure `.gitignore` covers `*.bak` if they tend to accumulate.

### Empty Catch Blocks Silently Swallowing Errors

- **Issue:** Extensive use of empty `catch {}` blocks with no error logging.
- **Files:**
  - `src/store.ts:112` -- `catch { /* ignore poll errors */ }`
  - `src/store.ts:128` -- `catch { /* ignore */ }`
  - `src/store.ts:142,162` -- `catch { /* ignore IDB errors */ }`
  - `src/store.ts:197` -- `.catch(() => {})`
  - `src/store.ts:199` -- `catch { /* fall through */ }`
  - `src/store.ts:530,535` -- `catch { /* ignore - backend unreachable */ }`
  - `src/lib/backendApi.ts:47,155,166,176,285` -- `catch { }` / `catch { /* ignore parse errors */ }`
  - `src/components/DetailModal.tsx:148` -- `.catch(() => {})`
  - `src/components/InputBar.tsx:140` -- `.catch(() => {})`
  - `src/components/Lightbox.tsx:82` -- `.catch(() => {})`
  - `src/components/SettingsModal.tsx:50,55,96` -- `.catch(() => {})`
- **Risk:** Real errors are silently ignored making debugging impossible. The IndexedDB errors are particularly concerning -- if IDB fails, image caching silently degrades with no user feedback.
- **Fix approach:** Add `console.warn` or structured logging in error paths. Create a production-safe log collector. Surface persistent errors (e.g., 3 consecutive IDB failures) to the user as a degraded-experience notification.

### Unused Import in Go

- **Issue:** Blank import of `golang.org/x/crypto/bcrypt` in models.go, unused.
- **Files:** `backend-go/service/models.go:7`
- **Code:** `_ "golang.org/x/crypto/bcrypt"`
- **Impact:** Minor. The bcrypt package is used in `auth.go` but the blank import here is unnecessary. The compiler ignores it but it indicates a refactoring artifact.

### JSON Parse Errors Discarded in SSE Stream

- **Issue:** The SSE stream parser silently ignores JSON parse errors.
- **Files:** `src/lib/backendApi.ts:285`
- **Code:** `} catch { /* ignore parse errors */ }`
- **Risk:** If the backend sends malformed SSE data, the client silently misses updates with no error reporting.

## Architecture Concerns

### Admin Routes via URL Pathname, Not React Router

- **Issue:** The frontend admin UI is selected by checking `window.location.pathname` directly in `main.tsx` rather than using a router-based solution.
- **Files:** `src/main.tsx:23-25`
- **Code:** `const isAdminRoute = window.location.pathname === '/admin' || window.location.pathname.startsWith('/admin/')`
- **Risk:** Two completely separate React trees are mounted with no shared context, no route transitions, and no code splitting for the main app vs admin. The admin page loads heavy components (AdminDashboard at 1,574 lines) even on quick admin login.
- **Fix approach:** Use react-router or similar to unify the routing. Lazy-load both App and AdminPage consistently.

### Duplicated API Request Logic

- **Issue:** The `request()` and `adminRequest()` functions in `backendApi.ts` and `adminApi.ts` are near-duplicates (auth header, error handling, content-type logic).
- **Files:** `src/lib/backendApi.ts:33-53`, `src/admin/adminApi.ts:49-68`
- **Risk:** Security fixes or error handling improvements must be duplicated. Divergence risk.
- **Fix approach:** Extract a shared `createApiClient(baseUrl, tokenKey)` factory function.

### Duplicated Business Types

- **Issue:** Backend Go types and frontend TypeScript types are independently defined with no shared source of truth.
- **Files:** `backend-go/service/models.go` (Go types), `src/types.ts` (TS types)
- **Risk:** Drift between backend API contract and frontend type expectations. Silent type mismatch bugs.
- **Fix approach:** Generate TypeScript types from Go structs using a tool, or at minimum add integration contract tests.

### Go `interface{}` Usage in Domain Models

- **Issue:** The `TaskRecord` struct uses `interface{}` for `Params`, `ActualParams`, `ActualParamsByImage`, `RevisedPromptByImage` fields.
- **Files:** `backend-go/service/models.go:102-105`
- **Risk:** No compile-time type safety for task parameters. JSON unmarshaling into `interface{}` loses all type checking. A typo in a param field name will silently pass through to the API.
- **Fix approach:** Define concrete structs for `TaskParams`, `ActualParams`, etc., and use `json.RawMessage` if flexibility is needed, or proper struct types.

### Large Zustand Store (936 Lines)

- **Issue:** The zustand store in `store.ts` is a God object containing auth, settings, task management, UI state, and image caching.
- **Files:** `src/store.ts`
- **Risk:** Any state update triggers re-render checks for all store consumers. Tight coupling: image cache logic is intermixed with task submission logic.
- **Fix approach:** Split into domain stores: `useAuthStore`, `useTaskStore`, `useImageCacheStore`, `useUIStore`. Use zustand slices or separate store instances.

### Tightly Coupled Auth and Store Initialization

- **Issue:** `bootstrapBackendSession()` in `store.ts` fetches auth user, tasks, and config in one function with inline image caching logic.
- **Files:** `src/store.ts:443-471`
- **Risk:** Hard to test individual behaviors. Auth refresh cannot be done without also refetching tasks and warming caches.
- **Fix approach:** Separate concerns: `refreshAuth()`, `refreshTasks()`, `warmImageCaches()` as independent functions called by an orchestrator.

## Error Handling Gaps

### Missing Error Boundary for React App

- **Issue:** No React Error Boundary is in place for the main app or admin app.
- **Files:** `src/main.tsx` (no error boundary wrapping App or AdminPage)
- **Risk:** An unhandled render error crashes the entire application with a white screen.
- **Fix approach:** Add a `<ErrorBoundary>` component wrapping both App and AdminPage.

### Backend Handler: Binding Errors Silently Discarded

- **Issue:** Multiple handlers use `_ = c.ShouldBindJSON(&body)` discarding the error.
- **Files:** `backend-go/handler/auth.go:22,80,102,129,158,176,200`, `backend-go/handler/tasks.go:31`
- **Risk:** If JSON is malformed, the handler proceeds with zero-value struct fields, potentially causing confusing downstream errors.
- **Fix approach:** Check the error and return 400 Bad Request with a clear message.

### Frontend Store: `FileReader.onerror` with No User Feedback

- **Issue:** `FileReader.onerror` callbacks in file-to-dataUrl conversion reject with a generic error but generate no user-facing toast.
- **Files:** `src/store.ts:193,924,933`
- **Risk:** A corrupted file upload silently fails.
- **Fix approach:** Pass error messages to the toast system.

### `saveGeneratedImagesWithAttribution` Skips Failed Saves Silently

- **Issue:** When saving generated images fails, the error is logged but the caller is not informed how many succeeded vs failed.
- **Files:** `backend-go/handler/generate.go:218-231`
- **Risk:** A task that generates 4 images but only saves 1 completes as "done" with only 1 output. The user doesn't know 3 images were lost.
- **Fix approach:** Return a count of saved/skipped images. Add a warning field to the task if any images failed to save.

### `.catch(() => {})` Without Recovery

- **Issue:** Multiple `.catch(() => {})` calls in store.ts discard errors without any recovery action.
- **Files:** `src/store.ts:197,645,661`
- **Code:** `putImage({...}).catch(() => {})`
- **Risk:** If IndexedDB writes fail, images are cached in memory but lost on page refresh. No retry mechanism.
- **Fix approach:** Log errors and implement a retry queue for failed IDB writes.

## Browser Compatibility Concerns

### International Fonts Loaded from External CDNs

- **Issue:** CSS imports Chinese fonts from external CDNs (`fontsapi.zeoseven.com`, `cdn.jsdelivr.net`).
- **Files:** `src/index.css:1-2`
- **Risk:** If these CDNs are unavailable (China GFW, DNS issues, service discontinuation), the UI font falls back unexpectedly. These are external dependencies outside project control.
- **Fix approach:** Bundle fonts locally or provide a self-hosted fallback font stack. Add `font-display: swap` for graceful degradation.

### Service Worker Registration with No Fallback

- **Issue:** Service worker registration catches errors and logs them, but offers no degraded-mode UI.
- **Files:** `src/main.tsx:11-14`
- **Risk:** If `sw.js` is 404 or registration fails, PWA offline mode and caching silently fail.
- **Fix approach:** Expose service worker status to the store and surface in settings UI.

## Database/Migration Concerns

### GORM AutoMigrate Used in Production

- **Issue:** Database schema is created/modified via `GORM AutoMigrate` at startup in production.
- **Files:** `backend-go/database/database.go:28`
- **Risk:** GORM AutoMigrate can silently add columns but won't remove them or handle complex migrations (renames, type changes). No migration versioning means rollback is impossible.
- **Fix approach:** Use GORM's migrator or a migration tool (golang-migrate) for versioned, reviewable migrations.

### No Database Index on `created_at` for Analytics Queries

- **Issue:** Analytics queries filter on `created_at` range, but the `BillingRecord` model only indexes `created_at` in the GORM tag (line 118) which GORM AutoMigrate may not actually create as a standalone index for range queries.
- **Files:** `backend-go/database/models.go:118` (`CreatedAt int64 -- gorm:"not null;index"`)
- **Risk:** Full table scans on growing billing data for analytics queries.
- **Fix approach:** Verify the index exists in SQLite. Add composite index on `(created_at, user_id)` for user-specific analytics if needed.

### Upload Directories with Orphaned User Data

- **Issue:** The `upload/` directory contains subdirectories (e.g., `8uW63bt9D_Xf6KNEwWHiF`) that appear to be user-upload folders.
- **Files:** `backend-go/upload/`
- **Risk:** If users are deleted but their uploaded images are not cleaned up from the filesystem, disk usage grows unboundedly.
- **Fix approach:** Implement a cleanup job that finds upload directories with no matching database records and removes them.

## Dependency Concerns

### Go 1.25.0 -- Future/Unstable Version

- **Issue:** The Go module specifies `go 1.25.0`, which is not a stable release at time of writing.
- **Files:** `backend-go/go.mod:3`
- **Risk:** May not be installable in all environments. Future Go toolchain behavior changes could break compilation.
- **Fix approach:** Pin to the latest stable Go release (e.g., 1.23.x or 1.24.x depending on release timeline).

### No `package-lock.json` or `yarn.lock` Committed

- **Issue:** Only `package.json` is present. No lockfile detected for npm/yarn/pnpm.
- **Files:** No `package-lock.json`, `yarn.lock`, or `pnpm-lock.yaml` found in repository root.
- **Risk:** Non-deterministic installs across environments. CI and production could get different dependency versions.
- **Fix approach:** Commit the lockfile. Use `npm ci` instead of `npm install` in CI.

### External Font CDN Dependencies

- **Issue:** Fonts loaded from external CDN URLs with no integrity hashes.
- **Files:** `src/index.css:1-2`
- **Risk:** Supply chain risk -- if the CDN is compromised or the font package version changes, injected malicious CSS or font files could affect the app.
- **Fix approach:** Pin to specific versions with `@version` in URL. Add `integrity` hashes. Consider self-hosting.

### Tailwind CSS v3 (Not v4)

- **Issue:** Tailwind CSS v3.4.17 in use; v4 has been released with significant performance improvements.
- **Files:** `package.json:41`
- **Risk:** Not critical, but missing out on v4's improved build performance and CSS-first configuration.
- **Fix approach:** Evaluate migration to Tailwind v4 when convenient; not urgent.

## Configuration Management Concerns

### `config.json` with API Keys in Plaintext

- **Issue:** The `backend-go/config.json` file contains API keys in plaintext (visible per `.gitignore` exclusion but stored on disk unencrypted).
- **Files:** `backend-go/config.json` (excluded from git via `.gitignore`)
- **Risk:** Plaintext API keys on disk can be exposed via backup systems, server access, or accidental commits of the file.
- **Fix approach:** Support environment variables as overrides for sensitive fields (`apiKey`, `jwtSecret`, `adminApikey`). Prefer env vars over file-based secrets.

### `VITE_BACKEND_URL` -- Hardcoded Fallback to `localhost:3001`

- **Issue:** Both `backendApi.ts` and `adminApi.ts` fall back to `http://localhost:3001` when `VITE_BACKEND_URL` is not set.
- **Files:** `src/lib/backendApi.ts:3`, `src/admin/adminApi.ts:3`
- **Code:** `|| 'http://localhost:3001'`
- **Risk:** In production without the env var set, all API calls go to localhost and silently fail.
- **Fix approach:** In production builds, require `VITE_BACKEND_URL` and fail the build if absent. Only use the fallback in development.

### `config.json` Written with `0644` Permissions (World-Readable)

- **Issue:** Config persistence writes `config.json` with `0644` permissions.
- **Files:** `backend-go/config/config.go:180,225,309`
- **Risk:** The config file contains API keys and JWT secrets; 0644 means any user on the system can read it.
- **Fix approach:** Use `0600` permissions for config files containing secrets.

## Type Safety Concerns

### `any` in Select Component's onChange

- **Issue:** The custom Select component's `onChange` callback passes `value: any` as the parameter type, forcing consumers to cast.
- **Files:** `src/components/Select.tsx:16`
- **Code:** `onChange: (value: any) => void`
- **Risk:** Consumers use `val as any` casts, bypassing all type checking.
- **Fix approach:** Make Select generic: `Select<T>({ value: T, onChange: (v: T) => void, ... })`.

### `noUnusedLocals: false` and `noUnusedParameters: false`

- **Issue:** TypeScript strict mode is enabled, but unused locals and parameters are not flagged.
- **Files:** `tsconfig.json:15-16`
- **Code:** `"noUnusedLocals": false`, `"noUnusedParameters": false`
- **Risk:** Dead variables and parameters accumulate, indicating incomplete refactors or unfinished features.
- **Fix approach:** Enable both flags after a cleanup pass to remove existing unused code.

## Code That Appears Incomplete

### `generateCode` -- No Error on Random Source Failure

- **Issue:** Random byte generation in `generateCode` discards the error from `rand.Int`.
- **Files:** `backend-go/service/auth.go:235`
- **Code:** `n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))`
- **Risk:** If the crypto random source fails, `n` is 0 and all generated codes start with the same character, reducing entropy.
- **Fix approach:** Handle the error or use `crypto/rand` directly.

### `getInviteCode` -- Error is Swallowed

- **Issue:** When `GetInviteCode` fails in the handler, the error is logged but `nil` is returned with 200 OK.
- **Files:** `backend-go/handler/auth.go:218-221`
- **Risk:** Silent failures where the user thinks they have no invite code but the real issue is a database error.
- **Fix approach:** Differentiate "not found" from "database error" in the response.

### `ReplacingUploads` Directory Cleanup Missing

- **Issue:** The upload directory `backend-go/upload/` contains user subdirectories with images, but there is no cleanup mechanism for old uploads.
- **Risk:** Unbounded disk growth. Deleted task images still exist on disk.

## Test Coverage Gaps

### Skipped Test Suites

- **Issue:** One `describe.skip` block for image cache behavior.
- **Files:** `src/store.test.ts:275`
- **What's not tested:** Image caching (memory cache, IndexedDB fallback, remote fetch).
- **Risk:** Image cache bugs silently corrupt or slow down the UI.
- **Priority:** Medium.

### No Frontend Integration Tests for SSE

- **Issue:** SSE streaming logic in `streamTaskStatus()` has no test coverage.
- **Files:** `src/lib/backendApi.ts:246-297`
- **What's not tested:** SSE connection lifecycle, reconnection logic, buffer parsing, error handling.
- **Risk:** SSE failures manifest as "stuck loading" states.

---

*Concerns audit: 2026-05-24*
