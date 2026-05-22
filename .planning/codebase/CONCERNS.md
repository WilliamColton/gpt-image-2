# Codebase Concerns

**Analysis Date:** 2026-05-22

## Tech Debt

**Backend configuration defaults and secret handling:**
- Issue: Backend startup accepts missing or malformed `config.json` and keeps insecure defaults for `JWTSecret` and `AdminApikey`. Endpoint API keys are persisted through admin updates without using the available encryption helpers, while `util.HashApikey`, `util.EncryptApikey`, and `util.DecryptApikey` are not wired into config persistence.
- Files: `backend-go/config/config.go`, `backend-go/util/crypto.go`, `backend-go/handler/admin.go`, `backend-go/.gitignore`
- Impact: A deployment that starts without a valid config can expose an admin login protected by known defaults. Persisted endpoint credentials remain recoverable from the runtime config file. Invalid JSON silently falls back to defaults, making misconfiguration hard to detect.
- Fix approach: Fail startup when production config uses default secrets; validate JSON unmarshal errors in `backend-go/config/config.go`; store endpoint API keys encrypted at rest or load them from environment/secret storage; add config tests that assert default secrets are rejected outside test/dev mode.

**Task persistence uses untyped JSON and ignores decode errors:**
- Issue: Task params, image IDs, actual params, and revised prompts are stored as JSON strings and decoded with ignored `json.Unmarshal` errors. `TasksUpdate` also accepts arbitrary map payloads and casts fields manually.
- Files: `backend-go/service/task.go`, `backend-go/handler/tasks.go`, `backend-go/database/models.go`
- Impact: Corrupt task rows or malformed client updates degrade silently into zero-value task state. Quota accounting in `CountPendingImages` can undercount or overcount when `params_json` is invalid. Future schema changes are fragile because validation is distributed and mostly implicit.
- Fix approach: Introduce typed request DTOs for task updates in `backend-go/handler/tasks.go`; return and log JSON decode errors in `backend-go/service/task.go`; consider first-class columns for frequently queried fields such as `n`, `status`, and image IDs, or wrap JSON serialization in a tested helper.

**Remote and local image caching paths are duplicated:**
- Issue: `getImageUrl` and `getRemoteImageDataUrl` independently construct authenticated image URLs using the same backend base and query token pattern.
- Files: `src/lib/backendApi.ts`, `src/store.ts`
- Impact: Changes to image auth, token placement, or backend base URL must be applied in multiple places. Divergence can create broken previews, stale caches, or token exposure regressions.
- Fix approach: Export and use one image URL builder from `src/lib/backendApi.ts`; keep token placement centralized; add a small test around URL construction in `src/store.test.ts` or a focused API utility test.

**Large feature modules carry multiple responsibilities:**
- Issue: Several files combine UI rendering, data loading, gesture handling, and business workflows in a single module.
- Files: `src/components/MaskEditorModal.tsx`, `src/store.ts`, `src/admin/AdminDashboard.tsx`, `src/components/InputBar.tsx`, `src/components/DetailModal.tsx`, `backend-go/service/openai.go`, `backend-go/service/auth.go`, `backend-go/handler/generate.go`
- Impact: Changes have high blast radius and are hard to review. State bugs in one workflow can affect unrelated UI behavior, and test coverage tends to target only public outcomes rather than internal edge cases.
- Fix approach: Split `src/components/MaskEditorModal.tsx` into canvas/history/gesture hooks and presentational toolbar pieces; split `src/store.ts` into auth/session, task execution, image cache, and UI slices; split `src/admin/AdminDashboard.tsx` tab panels into separate components; separate OpenAI request construction, failover, and response conversion in `backend-go/service/openai.go`.

**Frontend state mutation during upload flow:**
- Issue: `submitTask` mutates `orderedInputImages` entries directly after upload by assigning `img.id` and `img.dataUrl`.
- Files: `src/store.ts`
- Impact: Direct mutation can bypass Zustand change detection for `inputImages`, causing UI or mask draft state to drift from the uploaded IDs. It is fragile when the same image object is referenced by other local state.
- Fix approach: Build a new uploaded image array immutably, then call `setInputImages` and `updateTaskLocal` with the final IDs. Add a test that verifies `inputImages` reflects backend IDs after upload.

**Configuration surface is inconsistent with documentation:**
- Issue: README describes static frontend API URL/API key query-parameter workflows and Responses API support, while the active TypeScript types restrict `ApiMode` to `images` and the React client now defaults to backend-mediated auth and generation.
- Files: `README.md`, `src/types.ts`, `src/store.ts`, `src/lib/backendApi.ts`, `backend-go/handler/generate.go`
- Impact: Operators and future implementers can follow stale setup instructions, causing missing backend configuration, broken Responses API expectations, or insecure API key handling assumptions.
- Fix approach: Align README with the current backend-first flow; document `VITE_BACKEND_URL`, `backend-go/config.json`, admin endpoint management, and the current `images`-only frontend API mode; remove or reintroduce Responses API support consistently.

**Build output and source version labels are out of sync:**
- Issue: The package version and service worker cache name do not match current feature state. Production build emitted `package.json` version `0.2.15` while `public/sw.js` uses cache `gpt-image-playground-v0.1.5`.
- Files: `package.json`, `public/sw.js`
- Impact: PWA clients can retain stale app-shell assets longer than intended, and user-visible version/changelog behavior can become confusing.
- Fix approach: Tie service-worker cache names to the package/version during build or update `public/sw.js` alongside releases. Add a release checklist item or test that fails when cache version lags package version.

## Known Bugs

**Redeeming the same unused code concurrently can create an extra user without quota:**
- Symptoms: Two concurrent `LoginWithCode` calls can both create a user before one wins the `used_by IS NULL` update. The losing request still signs a token and returns a user even when `RowsAffected == 0` is only logged.
- Files: `backend-go/service/auth.go`
- Trigger: Submit the same unused redemption code in parallel to `/api/auth/login`.
- Workaround: Avoid parallel redemption attempts for the same code; manually delete the orphaned user if it appears.

**Quota checks are not atomic with task creation:**
- Symptoms: Multiple simultaneous generation requests can all pass `CheckQuota` before queued tasks are persisted or used counts are incremented, exceeding the user's quota.
- Files: `backend-go/handler/generate.go`, `backend-go/service/auth.go`, `backend-go/service/task.go`
- Trigger: Send concurrent `/api/generate` or `/api/edit` requests for the same user when remaining quota is low.
- Workaround: Configure endpoint `maxConcurrency` conservatively; manually adjust `used_count` or quota through admin after overruns.

**Uploaded images are fully read into memory despite multipart limits:**
- Symptoms: Large uploads or many concurrent uploads allocate full image buffers in process memory after `FormFile`, and generated base64 data URLs are also fully decoded before writing.
- Files: `backend-go/main.go`, `backend-go/handler/images.go`, `backend-go/service/image.go`, `backend-go/service/openai.go`
- Trigger: Upload a large multipart image or generate/edit multiple high-resolution images concurrently.
- Workaround: Keep reverse-proxy body limits low and avoid high parallelism until streaming size validation is added.

**SSE timeout can be shorter than generation timeout:**
- Symptoms: The task stream disconnects after 10 minutes even though OpenAI calls can run for 30 minutes. The frontend falls back to polling after stream errors, but users may see delayed status updates or stale queued/running state if polling also fails.
- Files: `backend-go/handler/tasks.go`, `backend-go/service/openai.go`, `src/lib/backendApi.ts`, `src/store.ts`
- Trigger: Long-running image generation or edit tasks that exceed 10 minutes.
- Workaround: Leave the page open so polling continues; refresh after long generations to bootstrap the latest task state from `/api/tasks`.

**Production build has a chunk-size warning:**
- Symptoms: `npm run build` completes but warns that `dist/assets/index-*.js` is larger than 500 kB after minification.
- Files: `vite.config.ts`, `src/main.tsx`, `src/App.tsx`, `src/components/MaskEditorModal.tsx`, `src/store.ts`
- Trigger: Run `npm run build`.
- Workaround: Current build still succeeds. Use route/component-level dynamic imports for heavy modals and canvas/editor code to reduce the initial app chunk.

## Security Considerations

**Secrets are present in an ignored backend config file:**
- Risk: The runtime backend config file exists locally and contains sensitive configuration. Even though it is ignored by git, accidental copying into logs, docs, screenshots, or deployment images would expose admin/API credentials.
- Files: `backend-go/config.json`, `backend-go/.gitignore`, `backend-go/config/config.go`
- Current mitigation: `backend-go/config.json` is listed in `.gitignore`; this audit does not quote any secret values.
- Recommendations: Move secrets to environment variables or a deployment secret manager; keep only non-secret defaults in checked-in examples; add startup checks that reject default secrets and optionally reject plaintext endpoint API keys.

**JWT tokens are stored in localStorage and sent in image query strings:**
- Risk: XSS or browser extension compromise can read user/admin JWTs. Image URLs include `?token=...`, which can leak through browser history, copied URLs, reverse-proxy logs, and referrers.
- Files: `src/lib/backendApi.ts`, `src/admin/adminApi.ts`, `src/store.ts`, `backend-go/middleware/middleware.go`, `backend-go/handler/images.go`
- Current mitigation: React escapes rendered content by default; public/admin text content is rendered as text, not HTML. Admin APIs require bearer tokens. Image access checks ownership in `ReadImageFileForUser`.
- Recommendations: Prefer secure HttpOnly same-site cookies for browser sessions or short-lived signed image URLs. Remove query-token fallback from `AuthMiddleware` once image delivery supports cookie/header auth safely. Add `Referrer-Policy` and cache-control headers for authenticated images.

**Default admin and JWT secrets are unsafe:**
- Risk: If `backend-go/config.json` is missing, malformed, or incomplete, the backend uses known default `JWTSecret` and `AdminApikey` values.
- Files: `backend-go/config/config.go`, `backend-go/handler/admin.go`, `backend-go/service/auth.go`
- Current mitigation: A local ignored config file can override defaults.
- Recommendations: Require explicit non-default secrets at startup; generate an initial admin secret once and print only a setup instruction; fail on malformed config instead of silently continuing.

**Admin login lacks rate limiting and lockout:**
- Risk: `/api/admin/login` compares a single static admin key and has no rate limiting, lockout, IP throttling, or audit counters.
- Files: `backend-go/handler/admin.go`, `backend-go/main.go`, `backend-go/middleware/logger.go`
- Current mitigation: A failed login returns unauthorized and writes a warning log without the attempted key.
- Recommendations: Add per-IP and global throttling around admin login; add configurable lockout/backoff; consider rotating admin credentials and using hashed verification.

**CORS is fully open for backend APIs:**
- Risk: Any origin can call authenticated backend APIs if it has a valid token. Open CORS also increases the blast radius of leaked tokens.
- Files: `backend-go/main.go`, `backend-go/handler/images.go`
- Current mitigation: `AllowCredentials` is false and auth tokens are bearer/query tokens rather than cookies.
- Recommendations: Make allowed origins configurable; use a strict allowlist in production; remove per-image `Access-Control-Allow-Origin: *` if authenticated images move to cookies or signed URLs.

**Endpoint configuration accepts arbitrary outbound targets:**
- Risk: Admin endpoint pool accepts any syntactically valid URL. A compromised admin token can configure internal-network or metadata-service targets and make the backend issue outbound requests.
- Files: `backend-go/handler/admin.go`, `backend-go/service/openai.go`, `backend-go/config/config.go`
- Current mitigation: Endpoint edits require admin auth and basic URL parsing.
- Recommendations: Restrict schemes to `https` by default; optionally block private/link-local IP ranges; add explicit operator override for non-public endpoints.

**Uploaded content type is trusted from multipart headers:**
- Risk: A client can upload non-image bytes with an image content type. The backend persists the bytes and serves them with the supplied MIME.
- Files: `backend-go/handler/images.go`, `backend-go/service/image.go`
- Current mitigation: Routes require authentication and image ownership. Browser rendering uses image tags for normal display paths.
- Recommendations: Detect MIME from content with `http.DetectContentType` or an image decoder; reject unsupported types; normalize stored extensions and response headers from detected MIME.

## Performance Bottlenecks

**SQLite is constrained to one open connection:**
- Problem: All DB operations share one open connection. Long-running writes, admin list operations, task polling, and image metadata operations serialize.
- Files: `backend-go/database/database.go`, `backend-go/service/task.go`, `backend-go/service/auth.go`, `backend-go/handler/tasks.go`
- Cause: `sqlDB.SetMaxOpenConns(1)` protects SQLite write behavior but limits concurrent request throughput.
- Improvement path: Keep WAL enabled but tune connection pools with separate read/write behavior, add indexes for common filters, and avoid polling-heavy query patterns where possible.

**Task stream polling loads whole task rows once per second per client:**
- Problem: Every open SSE connection polls SQLite every second and unmarshals full task payloads until completion.
- Files: `backend-go/handler/tasks.go`, `backend-go/service/task.go`
- Cause: No in-memory task notification bus exists; the stream handler watches database state.
- Improvement path: Add an event channel keyed by task ID that publishes status updates from `executeImageGeneration`; use DB polling only as a recovery fallback.

**Frontend bootstrapping can eagerly fetch and persist every done output image:**
- Problem: On login/bootstrap, all task image IDs are cached or fetched, and done outputs are warmed asynchronously from the backend into IndexedDB.
- Files: `src/store.ts`, `src/lib/db.ts`, `src/lib/backendApi.ts`
- Cause: `bootstrapBackendSession` iterates every task and calls `setCacheFromIdbOrRemote`; done outputs are queued through `warmImageContentCache` without pagination or visibility checks.
- Improvement path: Fetch images lazily when cards enter the viewport; paginate `/api/tasks`; cap concurrent image warming; add IndexedDB quota/error visibility to the UI.

**Canvas mask editor stores large undo snapshots:**
- Problem: Every stroke pushes full-canvas `ImageData` into undo history, capped at 40 entries.
- Files: `src/components/MaskEditorModal.tsx`, `src/lib/maskPreprocess.ts`
- Cause: Undo/redo is implemented with full pixel snapshots for simplicity.
- Improvement path: Store stroke commands or dirty rectangles; lower the cap based on canvas size; show memory-friendly behavior for high-resolution mask targets.

**Concurrent image generation can spawn many goroutines and large buffers:**
- Problem: Multi-image Codex CLI paths start one goroutine per requested image, each holding input image bytes and output base64 in memory.
- Files: `backend-go/service/openai.go`, `backend-go/handler/generate.go`, `backend-go/service/queue.go`
- Cause: `CallImagesGenerationsConcurrent` and `CallImagesEditsConcurrent` parallelize request fan-out directly based on `n`.
- Improvement path: Bound `n` server-side; enforce endpoint and per-user concurrency before goroutine fan-out; stream/decode outputs directly to files when possible.

## Fragile Areas

**Quota and redemption-code accounting:**
- Files: `backend-go/service/auth.go`, `backend-go/handler/generate.go`, `backend-go/service/task.go`, `backend-go/database/models.go`
- Why fragile: Quota checks, task insertion, generated image counts, and redemption-code use are separate DB operations without transactions that cover the full invariant.
- Safe modification: Use database transactions for code redemption and task reservation. Add tests that run concurrent login/generation requests and assert one code maps to one user and quota cannot be exceeded.
- Test coverage: Existing Go tests cover endpoint pool sorting and image ownership/deletion, but no tests cover concurrent `LoginWithCode`, `RedeemForUser`, `CheckQuota`, or `IncrementUsedCount` races.

**Backend task lifecycle:**
- Files: `backend-go/handler/generate.go`, `backend-go/service/openai.go`, `backend-go/service/task.go`, `backend-go/handler/tasks.go`, `src/store.ts`
- Why fragile: A client-supplied task ID is inserted as queued, a goroutine mutates a shared task pointer, SSE polls DB state, and the frontend separately patches local task state. Failures during image save can still produce a `done` task with fewer output images than requested.
- Safe modification: Treat backend task state transitions as authoritative; write explicit transition functions; store requested output count; mark partial success distinctly or fail when any output save fails.
- Test coverage: `src/store.test.ts` covers basic success/error UI flow, but Go tests do not cover `GenerateImage`, failed image saves, partial concurrent results, or SSE timeout behavior.

**Admin dashboard:**
- Files: `src/admin/AdminDashboard.tsx`, `src/admin/adminApi.ts`, `backend-go/handler/admin.go`, `backend-go/config/config.go`
- Why fragile: One React component owns all admin tabs, modal state, endpoint edits, user/code deletion, feedback, changelog, and announcement flows. Endpoint API keys are kept in component state and round-tripped back to the backend.
- Safe modification: Split each tab into isolated components with typed form state and dedicated API hooks. Add optimistic-state rollback for destructive actions and scrub endpoint API keys from logs/errors.
- Test coverage: No frontend tests target admin UI flows or admin API client behavior.

**Image storage lifecycle:**
- Files: `backend-go/service/image.go`, `backend-go/service/task.go`, `src/store.ts`, `src/lib/db.ts`, `backend-go/database/models.go`
- Why fragile: Deleting a task removes only task rows; generated/uploaded/mask image records and files remain unless explicit image deletion is used. Deleting a user removes the user row but not referenced tasks/images/files. Frontend deletion only clears memory cache for orphan candidates and does not request remote image deletion.
- Safe modification: Define ownership and retention rules. Add foreign keys or explicit cascading cleanup jobs for `users`, `tasks`, `images`, and upload files. Ensure task deletion either preserves shared images intentionally or deletes unreferenced server images.
- Test coverage: Go tests cover direct image deletion, but not task deletion cleanup, user deletion cleanup, or orphaned upload files.

**PWA and service worker caching:**
- Files: `src/main.tsx`, `public/sw.js`, `package.json`, `vite.config.ts`
- Why fragile: Service worker cache version is manual and app-shell cache strategy is hand-written. The service worker skips `/api/` only for same-origin paths and caches other same-origin GET assets opportunistically.
- Safe modification: Update cache names with every release or use a Vite PWA plugin that fingerprints precache assets. Verify admin route and backend/static deployments under the same origin.
- Test coverage: No tests exercise service-worker install/activate/fetch behavior.

**Canvas and mobile gesture handling:**
- Files: `src/components/MaskEditorModal.tsx`, `src/components/InputBar.tsx`, `src/lib/viewportTransform.ts`, `src/lib/canvasImage.ts`
- Why fragile: Pointer capture, pinch zoom, undo snapshots, drag/drop, paste handling, textarea resizing, and mobile collapse behavior rely on document/window listeners and manual refs.
- Safe modification: Keep browser event cleanup colocated with hooks; add focused tests around pure math helpers; manually verify touch, mouse, keyboard, and PWA display modes after changes.
- Test coverage: Pure viewport and mask preprocessing utilities have tests, but DOM gesture behavior and modal lifecycle are not covered.

## Scaling Limits

**Single-process backend with local SQLite and local uploads:**
- Current capacity: One process, one SQLite database under `backend-go/data/`, one local upload tree under `backend-go/upload/`, one DB open connection.
- Limit: Horizontal scaling creates split-brain uploads and config state. A single instance also concentrates all long-running image generation goroutines and SSE streams.
- Scaling path: Move images to object storage, tasks/config to a shared database, and task execution to a worker queue. Use a shared pub/sub or task status table for SSE updates.

**Unbounded task and admin list responses:**
- Current capacity: `/api/tasks`, admin user list, code list, feedback list, and changelog list return full collections.
- Limit: Large installations create slow queries, large JSON responses, and expensive frontend rendering.
- Scaling path: Add pagination, filters, and cursor-based listing to `backend-go/handler/tasks.go`, `backend-go/handler/admin.go`, `backend-go/handler/feedback.go`, and `backend-go/handler/changelog.go`.

**Browser IndexedDB stores base64 data URLs:**
- Current capacity: Images are stored as base64 strings in IndexedDB and in memory cache.
- Limit: Base64 overhead and duplicated memory pressure can hit browser storage quotas or mobile memory limits for large histories.
- Scaling path: Store Blobs instead of data URLs in `src/lib/db.ts`; use object URLs for previews; lazy-load and evict old entries with visible user controls.

**Endpoint concurrency is global per base URL only:**
- Current capacity: `MaxConcurrency` limits are keyed by `baseURL` across the process.
- Limit: No per-user concurrency, queue length, request timeout policy, or cancellation handling exists. A few users can occupy all endpoint slots for long-running jobs.
- Scaling path: Add per-user limits, global queue depth, cancellation, and task priority. Expose queue position in `TaskRecord` and SSE updates.

## Dependencies at Risk

**React 19 with Radix/shadcn-style wrappers:**
- Risk: The UI uses React 19 and multiple Radix primitives. Some admin and modal code is large and relies on portal/dialog behavior.
- Impact: Dependency updates can change focus handling, strict-mode behavior, or portal event semantics across `src/components/ui/*`, `src/components/MaskEditorModal.tsx`, and `src/admin/AdminDashboard.tsx`.
- Migration plan: Keep UI primitive wrappers thin and covered by smoke tests. Upgrade Radix packages together and manually verify modal stacking, Escape handling, and admin dialogs.

**Vite 6 build chunking:**
- Risk: Production build already warns about the main chunk exceeding 500 kB.
- Impact: Initial load can degrade on slow networks, especially because the main app includes heavy history/image/editor logic.
- Migration plan: Lazy-load heavy modals (`src/components/MaskEditorModal.tsx`, `src/components/DetailModal.tsx`, `src/components/Lightbox.tsx`) and consider manual chunks for Radix/lucide/vendor modules in `vite.config.ts`.

**SQLite/GORM as the backend persistence layer:**
- Risk: SQLite with GORM and one open connection is simple but limits concurrency and operational observability.
- Impact: Admin list operations, polling streams, and generation updates contend on one database file. Migration complexity increases as JSON fields accumulate.
- Migration plan: Keep repository methods behind service functions, add migrations, and design a Postgres-compatible schema before adding queue workers or multi-instance deployment.

**OpenAI SDK and image API assumptions:**
- Risk: `backend-go/service/openai.go` depends on `github.com/openai/openai-go/v3` image response fields and manually compensates for Codex CLI multi-image behavior.
- Impact: SDK or API changes can break request construction, actual-params reporting, or failover handling.
- Migration plan: Add unit tests around request parameter mapping and response conversion. Isolate SDK calls behind an interface for fake-client tests.

## Missing Critical Features

**No automated backend config validation command:**
- Problem: Operators cannot run a safe validation that checks missing secrets, endpoint URLs, and default admin/JWT values before starting the server.
- Blocks: Reliable production deployment and CI validation of backend configuration.

**No server-side request cancellation:**
- Problem: Once a generation goroutine starts, disconnecting the browser or deleting the task does not cancel the OpenAI request.
- Blocks: Efficient queue management, user-triggered cancellation, and cost control for accidental long-running jobs.

**No structured database migrations:**
- Problem: `AutoMigrate` creates/updates tables automatically without versioned migrations or rollback scripts.
- Blocks: Safe production schema evolution, data backfills, and reliable deployment previews.

**No remote image garbage collection:**
- Problem: Task deletion and user deletion do not clean upload files or image rows based on reference counts.
- Blocks: Long-running deployments with bounded disk usage.

**No CI pipeline detected:**
- Problem: No `.github/workflows/*` files were detected for frontend tests, Go tests, build, lint, or security checks.
- Blocks: Automatic validation of changes before merge/deploy.

## Test Coverage Gaps

**Backend auth, quota, and redemption races:**
- What's not tested: Concurrent code redemption, concurrent quota checks, used-count increments, disabled-user behavior during queued/running jobs, and admin login throttling.
- Files: `backend-go/service/auth.go`, `backend-go/handler/auth.go`, `backend-go/handler/generate.go`
- Risk: Race conditions can create orphan users, exceed quota, or allow expensive jobs to keep running after account changes.
- Priority: High

**Backend generation and failover:**
- What's not tested: `GenerateImage`, `executeImageGeneration`, failover ordering with errors, endpoint limiter behavior, concurrent generation merge behavior, partial image-save failures, and no-endpoint errors.
- Files: `backend-go/handler/generate.go`, `backend-go/service/openai.go`, `backend-go/service/queue.go`
- Risk: Core image generation can fail silently, report incorrect output counts, or overrun endpoint limits.
- Priority: High

**Admin APIs and admin UI:**
- What's not tested: User/code deletion, endpoint save validation, announcement updates, feedback status changes, changelog CRUD, admin token storage, and admin dashboard form behavior.
- Files: `backend-go/handler/admin.go`, `backend-go/handler/announcement.go`, `backend-go/handler/feedback.go`, `backend-go/handler/changelog.go`, `src/admin/AdminDashboard.tsx`, `src/admin/adminApi.ts`
- Risk: High-privilege operations can regress without automated detection.
- Priority: High

**Image lifecycle and cleanup:**
- What's not tested: Task deletion cleanup, user deletion cleanup, orphaned upload files, MIME validation, upload size rejection, and local/remote cache invalidation.
- Files: `backend-go/service/image.go`, `backend-go/service/task.go`, `src/store.ts`, `src/lib/db.ts`
- Risk: Disk usage grows unbounded and stale image references survive deletion.
- Priority: Medium

**PWA/service worker behavior:**
- What's not tested: Service-worker cache invalidation, app-shell fallback, same-origin API exclusion, and production registration/unregistration paths.
- Files: `public/sw.js`, `src/main.tsx`
- Risk: Users can see stale UI or cached assets after deployment.
- Priority: Medium

**Canvas/editor DOM interactions:**
- What's not tested: Mask editor pointer/pinch gestures, undo/redo history memory behavior, brush-size popover, drag-and-drop uploads, paste uploads, and mobile input collapse gestures.
- Files: `src/components/MaskEditorModal.tsx`, `src/components/InputBar.tsx`, `src/lib/viewportTransform.ts`, `src/lib/canvasImage.ts`
- Risk: Mobile and touch regressions can ship without unit-test failures.
- Priority: Medium

**Current automated checks:**
- What's not tested: Linting and formatting are not configured as package scripts; only Vitest and Go tests were detected/run.
- Files: `package.json`, `backend-go/go.mod`, `src/*.test.ts`, `backend-go/**/*_test.go`
- Risk: Style, unused code, and some TypeScript/Go hygiene issues rely on manual review outside `npm test`, `npm run build`, and `go test ./...`.
- Priority: Low

---

*Concerns audit: 2026-05-22*
