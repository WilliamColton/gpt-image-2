# Codebase Concerns

**Analysis Date:** 2026-05-24

## Tech Debt

**Large stateful frontend modules:**
- Issue: Core UI and state management live in very large files with many responsibilities and implicit coupling between auth, task submission, image caching, IndexedDB, polling, SSE, and modals.
- Files: `src/admin/AdminDashboard.tsx`, `src/components/MaskEditorModal.tsx`, `src/store.ts`, `src/components/InputBar.tsx`
- Impact: Changes to admin features, task lifecycle, image caching, or mask editing require navigating large components and increase regression risk.
- Fix approach: Split `src/store.ts` into auth/task/image-cache/settings slices, split `src/admin/AdminDashboard.tsx` by tab, and extract mask editor canvas/history/gesture logic from `src/components/MaskEditorModal.tsx` into focused hooks.

**Loose task payload typing:**
- Issue: Task persistence accepts loosely typed JSON and `interface{}`/`map[string]interface{}` payloads instead of validated DTOs.
- Files: `backend-go/service/models.go`, `backend-go/handler/tasks.go`, `backend-go/handler/generate.go`, `src/types.ts`
- Impact: Invalid `params`, task status, metadata, and output image IDs can enter `tasks.params_json`, `tasks.output_image_ids_json`, and UI state without server-side schema guarantees.
- Fix approach: Use concrete Go structs for client-editable task patches, validate `TaskParams`, and reserve status/output/actual metadata fields for server-owned updates.

**Runtime configuration persistence mixes secrets and mutable admin state:**
- Issue: API endpoint keys, pricing, invite settings, JWT secret, and admin key are stored in one JSON file and rewritten directly from request handlers.
- Files: `backend-go/config/config.go`, `backend-go/handler/admin.go`, `.gitignore`
- Impact: Admin UI writes can overwrite unrelated configuration fields, failed writes only log errors, and secret-bearing configuration is tightly coupled to normal app settings.
- Fix approach: Separate secrets from mutable app settings, return persistence errors to handlers, write config with a temp-file-plus-rename flow, and store secret material outside the admin-editable JSON document.

**Implicit database migrations:**
- Issue: Startup uses GORM `AutoMigrate` as the migration system.
- Files: `backend-go/database/database.go`, `backend-go/database/models.go`, `backend-go/database/models_test.go`
- Impact: Schema changes lack versioned migration history, rollback steps, deployment ordering, and explicit data backfills.
- Fix approach: Add a migration tool and checked-in migration files for schema changes; keep `AutoMigrate` for tests only or remove it from production startup.

**No ownership cleanup model:**
- Issue: User, task, image, and billing records are not connected by foreign keys or cleanup services.
- Files: `backend-go/database/models.go`, `backend-go/service/auth.go`, `backend-go/service/task.go`, `backend-go/service/image.go`, `src/store.ts`
- Impact: Deleting users or tasks leaves orphaned tasks, images, uploaded files, billing rows, and browser-cached records.
- Fix approach: Define explicit retention rules, add cleanup services for user/task deletion, reference-count image usage, and keep billing retention intentionally separate from user deletion.

**Dead credential helper code:**
- Issue: API key hashing/encryption helpers are not called anywhere.
- Files: `backend-go/util/crypto.go`
- Impact: The presence of unused encryption helpers suggests API keys are intended to be protected, but endpoint keys remain plain runtime configuration values.
- Fix approach: Either remove unused helpers or wire them into endpoint persistence and retrieval with a key-rotation plan.

## Known Bugs

**Concurrent first-use redemption code login can create extra users:**
- Symptoms: Two concurrent requests with the same unused redemption code can both create user rows; only one updates `redemption_codes.used_by`, but the loser still receives a signed token and quota.
- Files: `backend-go/service/auth.go`, `backend-go/handler/auth.go`
- Trigger: Concurrent `POST /api/auth/login` requests using the same unused code.
- Workaround: Not detected.
- Fix approach: Wrap code lookup, user creation, and code marking in one transaction; mark the code first with `used_by IS NULL`, check `RowsAffected`, and create the user only after the claim succeeds.

**Deleting a running task does not cancel generation and can recreate the task:**
- Symptoms: A task deleted while its goroutine is still generating can be reinserted when `executeImageGeneration` calls `UpsertTask` after the external API returns.
- Files: `backend-go/handler/generate.go`, `backend-go/service/task.go`, `src/store.ts`
- Trigger: Delete a queued or running task from the UI while the backend goroutine is waiting for an endpoint slot or OpenAI response.
- Workaround: Avoid deleting in-progress tasks.
- Fix approach: Add cancellation/tombstone state, check task existence before saving outputs, and pass request/task context into endpoint acquisition and OpenAI calls.

**Backend service tests fail on `needsMigration` JSON omission:**
- Symptoms: `go test ./...` in `backend-go` fails `TestAuthUserJSON_OmitsNeedsMigrationWhenFalse` because `needsMigration` is serialized when false.
- Files: `backend-go/service/models.go`, `backend-go/service/models_test.go`
- Trigger: Running `go test ./...` from `backend-go`.
- Workaround: Not detected.
- Fix approach: Add `omitempty` to `AuthUser.NeedsMigration` or update the test and API contract to require explicit false.

**Frontend test command fails before Vitest runs in this checkout:**
- Symptoms: `npm test` exits during Rollup startup because the native optional package for the Linux x64 GNU platform is missing from `node_modules`.
- Files: `package.json`, `package-lock.json`
- Trigger: Running `npm test` with the current dependency install state.
- Workaround: Reinstall dependencies with npm so optional dependencies are restored.
- Fix approach: Document the reinstall requirement, keep lockfile and npm version consistent, and add CI that installs from scratch.

**Invite reward update can silently fail after user creation:**
- Symptoms: Invited user registration can succeed while inviter quota reward update fails without returning an error.
- Files: `backend-go/service/auth.go`
- Trigger: Database error during inviter quota update after `database.DB.Create(newUser)` succeeds.
- Workaround: Not detected.
- Fix approach: Perform invited-user creation and inviter reward update in one transaction and return an error if either write fails.

**Admin middleware does not verify the admin subject against database state:**
- Symptoms: A valid admin-role JWT remains sufficient for admin access without loading an admin row or checking admin account status.
- Files: `backend-go/middleware/middleware.go`, `backend-go/handler/admin.go`, `backend-go/database/database.go`
- Trigger: Any request to `/api/admin/*` with a valid signed role claim of `admin`.
- Workaround: Rotate the JWT secret to invalidate outstanding tokens.
- Fix approach: Store admin identities, verify the subject exists and is active on each admin request, and include token version/revocation data.

## Security Considerations

**Default secrets are accepted at startup:**
- Risk: Missing or invalid `config.json` leaves the application running with placeholder JWT/admin secret values.
- Files: `backend-go/config/config.go`, `backend-go/handler/admin.go`, `backend-go/service/auth.go`
- Current mitigation: `backend-go/config.json` is ignored by Git in `.gitignore`.
- Recommendations: Fail startup when required secrets are missing or placeholders, validate JSON unmarshal errors, and require strong secret values through environment or secure config management.

**Admin key and user tokens are stored in `localStorage`:**
- Risk: Any successful XSS or malicious browser extension can read long-lived user and admin tokens.
- Files: `src/lib/backendApi.ts`, `src/admin/adminApi.ts`, `src/store.ts`
- Current mitigation: Not detected.
- Recommendations: Prefer HttpOnly, Secure, SameSite cookies or a short-lived access token plus refresh-token flow; clear browser storage on logout and user switch.

**Image URLs carry bearer tokens in the query string:**
- Risk: Tokens can leak through browser history, logs, screenshots, referrers, cache keys, and copied image URLs.
- Files: `src/lib/backendApi.ts`, `src/store.ts`, `backend-go/middleware/middleware.go`, `backend-go/handler/images.go`
- Current mitigation: Image responses use authenticated backend routes and `cache: no-store` in some fetch callers.
- Recommendations: Remove `?token=` authentication, fetch images with `Authorization` headers and object URLs, or issue short-lived signed image URLs scoped to one image.

**Open CORS policy:**
- Risk: Any origin can call public and authenticated API routes; bearer tokens in browser storage remain the main protection boundary.
- Files: `backend-go/main.go`
- Current mitigation: `AllowCredentials` is false.
- Recommendations: Restrict origins by deployment environment, add security headers, and document trusted frontend origins.

**No rate limiting or lockout:**
- Risk: Admin key login, password login, registration, redemption-code login, feedback creation, and task generation can be brute-forced or spammed.
- Files: `backend-go/handler/admin.go`, `backend-go/handler/auth.go`, `backend-go/handler/feedback.go`, `backend-go/handler/generate.go`, `backend-go/middleware/middleware.go`
- Current mitigation: Quota checks limit successful generation for normal users.
- Recommendations: Add per-IP and per-account throttling middleware, exponential backoff, account lockout for password auth, and separate limits for admin login and generation submissions.

**Image upload trusts client metadata:**
- Risk: Arbitrary bytes can be saved and served with a caller-provided MIME type; large uploads can consume memory and disk.
- Files: `backend-go/main.go`, `backend-go/handler/images.go`, `backend-go/service/image.go`
- Current mitigation: Gin multipart memory is set to 50 MiB and upload paths are constrained through `backend-go/util/paths.go`.
- Recommendations: Add `http.MaxBytesReader`, decode and validate image content, whitelist MIME types, enforce dimensions and per-user storage quota, and reject SVG or unsupported formats unless sanitized.

**Mutable endpoint configuration can expose API keys to admin UI:**
- Risk: Admin pages fetch and render endpoint API keys, and the keys are persisted as plain JSON config values.
- Files: `backend-go/handler/admin.go`, `backend-go/config/config.go`, `src/admin/AdminDashboard.tsx`, `src/admin/adminApi.ts`
- Current mitigation: Admin routes require an admin JWT; API key inputs render as password fields unless toggled visible.
- Recommendations: Redact keys in GET responses, support write-only key updates, store encrypted keys or external secret references, and audit key reads.

**JWT verification does not pin the expected signing method:**
- Risk: Token verification relies on library defaults rather than an explicit algorithm policy.
- Files: `backend-go/service/auth.go`
- Current mitigation: Tokens are signed with HMAC SHA-256 by `SignToken`.
- Recommendations: Check `t.Method` explicitly, validate required claims and issuer/audience if added, and keep token lifetime configurable.

**Proxy trust is not configured:**
- Risk: Client IP logging and future IP-based rate limits can be incorrect behind proxies.
- Files: `backend-go/main.go`, `backend-go/middleware/logger.go`
- Current mitigation: Not detected.
- Recommendations: Set Gin trusted proxies for deployment and use a known proxy header policy.

**Secret-like runtime files are present:**
- Risk: Local runtime configuration and dev proxy files can contain credentials or internal endpoints.
- Files: `backend-go/config.json`, `dev-proxy.config.json`, `.gitignore`
- Current mitigation: Both files are ignored by Git in `.gitignore`.
- Recommendations: Keep these files unquoted in docs/logs, validate permissions, and provide non-secret example files.

## Performance Bottlenecks

**Unpaginated list endpoints:**
- Problem: Admin and user list APIs load entire tables into memory and return full JSON arrays.
- Files: `backend-go/service/auth.go`, `backend-go/service/task.go`, `backend-go/service/feedback.go`, `backend-go/service/changelog.go`, `backend-go/handler/admin.go`, `backend-go/handler/tasks.go`
- Cause: Queries use `Find` with ordering but no `Limit`/`Offset` or cursor pagination.
- Improvement path: Add paginated API contracts, indexes for filter/sort fields, and frontend virtualized tables for large lists.

**Invite list uses an N+1 count query:**
- Problem: Listing invite codes runs one count query per invite owner.
- Files: `backend-go/service/auth.go`, `backend-go/handler/admin.go`
- Cause: `ListInvites` loads owners then counts `invited_by` per owner in a loop.
- Improvement path: Use a grouped aggregate query (`GROUP BY invited_by`) and join/merge counts in one database round trip.

**SSE streams poll the database every second:**
- Problem: Each active task stream reads the task row once per second until completion or timeout.
- Files: `backend-go/handler/tasks.go`, `backend-go/service/task.go`, `backend-go/database/database.go`
- Cause: Streaming is implemented as per-connection polling on SQLite, which is limited to one open connection.
- Improvement path: Publish task status changes through an in-process broker for single-instance mode or external pub/sub for distributed mode; reduce polling frequency as fallback.

**Frontend bootstrap eagerly walks all tasks and image IDs:**
- Problem: Session bootstrap loads all tasks and iterates every input, mask, and output image to seed caches.
- Files: `src/store.ts`, `src/lib/backendApi.ts`, `src/lib/db.ts`
- Cause: `bootstrapBackendSession` fetches all tasks and warms generated output images without pagination or viewport-based lazy loading.
- Improvement path: Paginate task loading, warm only visible thumbnails, and lazy-load image data on demand.

**Browser image cache stores base64 data URLs indefinitely:**
- Problem: Fetched backend images are converted to base64 data URLs and persisted to IndexedDB.
- Files: `src/store.ts`, `src/lib/db.ts`
- Cause: `fetchAndCacheImage` stores full image data in both memory cache and IndexedDB without eviction.
- Improvement path: Store Blob/object URLs where possible, add size/age-based eviction, and namespace cached images by user.

**Mask editor undo history can consume large memory:**
- Problem: Up to 40 full `ImageData` snapshots are retained per mask-editing session.
- Files: `src/components/MaskEditorModal.tsx`, `src/lib/maskPreprocess.ts`
- Cause: `pushUndoSnapshot` stores complete canvas buffers for each stroke.
- Improvement path: Store compressed deltas or region snapshots, lower history depth for large canvases, and expose memory-aware limits.

**Generation goroutines can wait indefinitely for endpoint slots:**
- Problem: Each submitted task starts a goroutine before acquiring an endpoint concurrency slot.
- Files: `backend-go/handler/generate.go`, `backend-go/service/queue.go`, `backend-go/service/openai.go`
- Cause: `AcquireSlotFrom` loops until a slot is available and has no context cancellation or global queue bound.
- Improvement path: Add a bounded worker queue, per-user queue limits, context cancellation, and persisted queued state.

## Fragile Areas

**Task lifecycle is split across frontend optimistic state, backend DB, SSE, and polling:**
- Files: `src/store.ts`, `src/lib/backendApi.ts`, `backend-go/handler/generate.go`, `backend-go/handler/tasks.go`, `backend-go/service/task.go`
- Why fragile: The frontend creates optimistic tasks, uploads images, submits generation, starts SSE, falls back to polling, and independently allows task patch/delete operations.
- Safe modification: Treat the backend as the owner of lifecycle fields, keep frontend writes to user preferences such as favorites, and add state-machine tests for queued/running/done/error/delete transitions.
- Test coverage: Some frontend store tests exist in `src/store.test.ts`; image cache tests are skipped and cancellation/delete race tests are not detected.

**Billing, quota, image save, and task finalization are not transactional:**
- Files: `backend-go/handler/generate.go`, `backend-go/service/billing.go`, `backend-go/service/auth.go`, `backend-go/service/image.go`, `backend-go/service/task.go`
- Why fragile: Generated images are saved, billing rows are inserted, used count is incremented, and task status is updated through separate writes; errors after image save are logged without rollback.
- Safe modification: Add idempotent finalization with retryable steps, unique billing constraints, reconciliation jobs, and explicit partial-failure states.
- Test coverage: Billing helper tests exist in `backend-go/handler/generate_billing_test.go` and `backend-go/service/billing_test.go`; end-to-end failure recovery tests are not detected.

**Endpoint failover and limiter state are in memory:**
- Files: `backend-go/service/openai.go`, `backend-go/service/queue.go`, `backend-go/config/config.go`, `backend-go/handler/admin.go`
- Why fragile: Limiters are keyed by base URL, refreshed by deleting a `sync.Map`, and are unaware of external backend instances.
- Safe modification: Use endpoint IDs, version limiter configuration, avoid deleting active semaphores during in-flight requests, and externalize concurrency control for multi-instance deployments.
- Test coverage: Failover tests exist in `backend-go/service/openai_failover_test.go`; limiter reconfiguration and in-flight refresh tests are not detected.

**Config loading swallows missing and malformed config:**
- Files: `backend-go/config/config.go`, `backend-go/config/config_test.go`, `backend-go/main.go`
- Why fragile: Missing config returns success, malformed JSON is ignored, and the app proceeds with defaults.
- Safe modification: Return structured validation errors, distinguish development defaults from production requirements, and test invalid JSON and placeholder-secret startup.
- Test coverage: Config tests cover endpoint sorting and persistence; startup validation for missing/malformed secret-bearing config is not detected.

**Admin dashboard holds many independent flows in one component:**
- Files: `src/admin/AdminDashboard.tsx`, `src/admin/adminApi.ts`, `src/admin/AdminDashboard.test.tsx`
- Why fragile: Users, codes, endpoint config, pricing, analytics, announcements, feedback, changelog, invite config, and password reset state all share one render tree and many local states.
- Safe modification: Extract tab components with isolated data hooks and colocated tests for each admin workflow.
- Test coverage: Focused tests exist for API calls and money formatting; component-level coverage for most tabs is not detected.

**Path safety depends on stored relative paths remaining valid:**
- Files: `backend-go/service/image.go`, `backend-go/util/paths.go`, `backend-go/database/models.go`
- Why fragile: Image records store file paths in the database; if paths are edited, moved, or generated incorrectly, file reads fail or rely on resolver guards.
- Safe modification: Store normalized user/image IDs and derive paths instead of persisting file paths, or add DB constraints and repair tooling.
- Test coverage: Image handler/service tests exist in `backend-go/handler/images_test.go` and `backend-go/service/image_test.go`; corrupted path repair tests are not detected.

## Scaling Limits

**Single-instance SQLite backend:**
- Current capacity: SQLite is configured with WAL and `SetMaxOpenConns(1)`.
- Limit: Write throughput and all DB-backed polling/listing operations are serialized inside one process.
- Scaling path: Move to a network database for multi-instance deployments and keep SQLite as a local/single-user mode.
- Files: `backend-go/database/database.go`, `backend-go/database/models.go`

**Local filesystem uploads:**
- Current capacity: Uploaded and generated files live under local `backend-go/upload/`.
- Limit: Horizontal scaling, backups, and cross-instance image serving break without shared storage.
- Scaling path: Store images in object storage with signed URLs and keep metadata in the database.
- Files: `backend-go/service/image.go`, `backend-go/util/paths.go`, `.gitignore`

**In-memory queue and endpoint limiters:**
- Current capacity: Queued generation work and endpoint semaphores exist only inside one Go process.
- Limit: Multiple backend instances can overrun endpoint concurrency and cannot resume in-memory work after restart.
- Scaling path: Use a persistent job queue with workers and distributed concurrency controls.
- Files: `backend-go/handler/generate.go`, `backend-go/service/queue.go`, `backend-go/service/openai.go`

**Long-running image calls occupy goroutines:**
- Current capacity: External image generation calls use 30-minute contexts and one goroutine per concurrent output in Codex CLI/multi-image paths.
- Limit: Large `n`, many users, or slow endpoints increase goroutine and memory pressure.
- Scaling path: Bound `n` server-side, set per-user and global queue limits, and run jobs in a worker pool.
- Files: `backend-go/service/openai.go`, `backend-go/handler/generate.go`, `backend-go/service/models.go`

**Admin analytics scans billing aggregates directly:**
- Current capacity: Analytics queries aggregate over `billing_records` at request time.
- Limit: Large billing tables make dashboard responses slower and compete with task writes on SQLite.
- Scaling path: Add time-bucket rollups, pagination for breakdowns, and database indexes aligned to analytics filters.
- Files: `backend-go/service/analytics.go`, `backend-go/database/models.go`, `backend-go/handler/admin.go`

## Dependencies at Risk

**Rollup optional native package install state:**
- Risk: Frontend tests cannot start when npm optional dependencies are missing from `node_modules`.
- Impact: Vitest coverage is unavailable in the affected checkout.
- Migration plan: Reinstall with npm in CI and local setup, pin npm behavior, and document recovery in project setup instructions.
- Files: `package.json`, `package-lock.json`

**SQLite as primary persistence:**
- Risk: SQLite with one open connection fits single-instance deployment but becomes a bottleneck for concurrent web workloads.
- Impact: Polling, list endpoints, generation finalization, billing writes, and admin analytics contend for one connection.
- Migration plan: Introduce a database abstraction boundary only where needed, then migrate production deployments to Postgres/MySQL while keeping SQLite tests.
- Files: `backend-go/database/database.go`, `backend-go/database/models.go`

**OpenAI SDK and Images API coupling:**
- Risk: Image parameter structs and response metadata are tightly mapped to the OpenAI Go SDK.
- Impact: Provider changes or SDK changes affect generation, edit, actual parameter extraction, and revised prompt handling.
- Migration plan: Keep provider-specific code behind a service interface and add contract tests for generation/edit behavior.
- Files: `backend-go/service/openai.go`, `backend-go/service/models.go`, `backend-go/handler/generate.go`

**Gin default middleware behavior:**
- Risk: `gin.Default()` installs default logger/recovery behavior and the app also adds a custom request logger.
- Impact: Logs can be duplicated or include formatting not aligned with production structured logging.
- Migration plan: Use `gin.New()`, explicitly install recovery and logging middleware, and set Gin mode from configuration.
- Files: `backend-go/main.go`, `backend-go/middleware/logger.go`, `backend-go/log/log.go`

**Browser storage APIs:**
- Risk: IndexedDB, FileReader, Canvas, Service Worker, and localStorage are central to core behavior.
- Impact: Browser quota limits, private browsing restrictions, or storage eviction can break image reuse and cached generated outputs.
- Migration plan: Treat backend storage as canonical, keep browser caches optional, and add graceful degradation paths.
- Files: `src/lib/db.ts`, `src/store.ts`, `src/components/MaskEditorModal.tsx`, `public/sw.js`

## Missing Critical Features

**Graceful shutdown and job recovery:**
- Problem: In-flight generation goroutines are not tracked for shutdown, cancellation, or restart recovery.
- Blocks: Reliable deployments, rolling restarts, and deterministic task state after process crashes.
- Files: `backend-go/main.go`, `backend-go/handler/generate.go`, `backend-go/service/task.go`

**Server-side parameter validation:**
- Problem: Backend generation requests do not enforce maximum prompt length, maximum `n`, accepted enum values, image count, compression range, or image dimension limits.
- Blocks: Safe public deployment and reliable external API calls.
- Files: `backend-go/handler/generate.go`, `backend-go/service/models.go`, `src/components/InputBar.tsx`, `src/types.ts`

**Storage retention and quota management:**
- Problem: There is no cleanup job or storage quota for uploaded/generated images, orphaned files, browser IndexedDB entries, feedback, changelogs, or old tasks.
- Blocks: Long-running deployments with predictable disk usage and privacy retention requirements.
- Files: `backend-go/service/image.go`, `backend-go/service/task.go`, `src/lib/db.ts`, `src/store.ts`

**Audit logging for admin actions:**
- Problem: Admin mutations update users, config, pricing, invites, announcements, feedback, changelog, and passwords without a durable audit table.
- Blocks: Incident response and accountability for privileged changes.
- Files: `backend-go/handler/admin.go`, `backend-go/handler/announcement.go`, `backend-go/handler/changelog.go`, `backend-go/handler/feedback.go`, `backend-go/database/models.go`

**Configuration validation and secret rotation:**
- Problem: Config loading has no validation report, no required secret checks, and no rotation flow for admin keys, JWT signing keys, or endpoint API keys.
- Blocks: Safe production configuration management.
- Files: `backend-go/config/config.go`, `backend-go/handler/admin.go`, `backend-go/service/auth.go`

**Centralized error contract:**
- Problem: Handlers return raw service error strings, generic messages, and success-with-empty responses inconsistently.
- Blocks: Reliable frontend error handling and localization.
- Files: `backend-go/handler/auth.go`, `backend-go/handler/tasks.go`, `backend-go/handler/images.go`, `backend-go/handler/admin.go`, `src/lib/backendApi.ts`, `src/admin/adminApi.ts`

## Test Coverage Gaps

**Skipped image cache tests:**
- What's not tested: Memory cache, IndexedDB restore, backend image fetch, concurrent fetch deduplication, and remote URL fallback in the current store implementation.
- Files: `src/store.test.ts`, `src/store.ts`, `src/lib/db.ts`
- Risk: Image display, reuse, and local caching can regress without test failures.
- Priority: High

**Task deletion/cancellation races:**
- What's not tested: Deleting queued/running tasks, user deletion during generation, and backend completion after local deletion.
- Files: `backend-go/handler/generate.go`, `backend-go/service/task.go`, `src/store.ts`
- Risk: Deleted tasks can reappear, quota and billing can change after deletion, and generated files can become orphaned.
- Priority: High

**Concurrent redemption-code login:**
- What's not tested: Multiple simultaneous first-use logins for the same redemption code.
- Files: `backend-go/service/auth.go`, `backend-go/handler/auth.go`, `backend-go/service/auth_test.go`
- Risk: Quota and access can be granted to extra users.
- Priority: High

**Server-side generation validation:**
- What's not tested: Invalid `TaskParams`, excessive `n`, excessive image count, invalid image IDs, huge prompts, and unsupported output formats.
- Files: `backend-go/handler/generate.go`, `backend-go/service/models.go`, `backend-go/service/openai.go`
- Risk: Bad requests reach external APIs or create expensive backend work.
- Priority: High

**Config failure paths:**
- What's not tested: Malformed `config.json`, placeholder secrets, persistence write failures, file permission issues, and partial config writes.
- Files: `backend-go/config/config.go`, `backend-go/config/config_test.go`, `backend-go/handler/admin.go`
- Risk: The app can run insecurely or report successful admin config changes that are not persisted.
- Priority: High

**Security middleware:**
- What's not tested: CORS origin restrictions, rate limiting, token-in-query rejection, admin token revocation, and proxy trust behavior.
- Files: `backend-go/main.go`, `backend-go/middleware/middleware.go`, `backend-go/middleware/logger.go`, `src/lib/backendApi.ts`
- Risk: Security controls can remain absent or regress silently.
- Priority: High

**Billing finalization failure recovery:**
- What's not tested: Image save success with billing insert failure, used count update failure, task final update failure, duplicate finalization, and reconciliation.
- Files: `backend-go/handler/generate.go`, `backend-go/service/billing.go`, `backend-go/service/auth.go`, `backend-go/service/task.go`
- Risk: Financial analytics and quota state can diverge from generated outputs.
- Priority: Medium

**Admin dashboard tab behavior:**
- What's not tested: Most admin tab component flows, confirmation modals, optimistic state after failures, endpoint key visibility, invite config, changelog, feedback, and analytics rendering.
- Files: `src/admin/AdminDashboard.tsx`, `src/admin/AdminDashboard.test.tsx`, `src/admin/adminApi.ts`
- Risk: Large UI changes can break admin workflows while API unit tests pass.
- Priority: Medium

**Upload validation and storage cleanup:**
- What's not tested: Oversized uploads, invalid MIME/content mismatch, unsupported formats, per-user storage limits, orphan cleanup, and path corruption.
- Files: `backend-go/handler/images.go`, `backend-go/service/image.go`, `backend-go/util/paths.go`, `backend-go/handler/images_test.go`
- Risk: Storage exhaustion, invalid files, and missing images can reach production.
- Priority: Medium

**Frontend service worker cache behavior:**
- What's not tested: App shell caching, cache version migration, offline fallback, API bypass, and stale asset behavior.
- Files: `public/sw.js`, `src/main.tsx`
- Risk: Users can receive stale frontend code or broken offline fallback after deploys.
- Priority: Low

---

*Concerns audit: 2026-05-24*
