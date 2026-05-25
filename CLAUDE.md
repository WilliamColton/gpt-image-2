<!-- GSD:project-start source:PROJECT.md -->
## Project

**GPT Image Playground Stability Program**

GPT Image Playground is an existing React/PWA + Go/Gin backend application for OpenAI-compatible image generation and editing. It already supports user authentication, task submission, image upload/storage, Server-Sent Events task updates, admin management, endpoint failover, quotas, invite flows, billing records, announcements, changelog, and feedback.

This project initialization defines the next milestone: fix known bugs and improve reliability where defects reduce customer experience, increase operational cost, or weaken operational trust.

**Core Value:** Users and admins can rely on image generation, authentication, task lifecycle, quota, and billing workflows to behave correctly without manual recovery or avoidable support cost.

### Constraints

- **Scope**: Use the known bug list from the codebase map as the v1 source of truth — the user chose not to broaden this roadmap into general production hardening.
- **Prioritization**: Balance customer experience, operational cost, and safety/reliability rather than optimizing only one category.
- **Workflow**: Interactive GSD mode with standard phase granularity, parallel execution where independent plans allow it, and research/plan-check/verifier enabled.
- **Tech stack**: Preserve the current React/Vite/Zustand frontend and Go/Gin/GORM/SQLite backend unless a local change is necessary for a bug fix.
- **Persistence**: Do not read or commit `backend-go/config.json`, `dev-proxy.config.json`, runtime SQLite files, uploaded images, `dist/`, `node_modules/`, or `.claude/worktrees/`.
- **Testing**: Prefer co-located Vitest tests for frontend logic and Go `testing` with temporary SQLite/Gin `httptest` for backend service/handler behavior.
- **Security**: Treat auth, admin access, token handling, upload paths, and endpoint API keys as sensitive boundaries; validate at HTTP/API boundaries and avoid exposing secrets.
<!-- GSD:project-end -->

<!-- GSD:stack-start source:codebase/STACK.md -->
## Technology Stack

## Languages
- TypeScript 5.8.3 - React frontend and admin UI in `src/**/*.ts` and `src/**/*.tsx`; compiler settings live in `tsconfig.json`.
- Go 1.25.0 - API backend in `backend-go/**`; module declared in `backend-go/go.mod` as `gpt-image-playground/backend`.
- CSS / Tailwind CSS - application styling in `src/index.css`, Tailwind theme configuration in `tailwind.config.js`, and PostCSS wiring in `postcss.config.js`.
- JavaScript config files - build and UI tooling configuration in `tailwind.config.js`, `postcss.config.js`, and `components.json`.
- JSON configuration - frontend/package metadata in `package.json` and backend runtime configuration schema in `backend-go/config/config.go`.
## Runtime
- Browser SPA/PWA - React is mounted from `src/main.tsx`, service worker registration is controlled in `src/main.tsx`, app-shell caching lives in `public/sw.js`, and PWA metadata lives in `public/manifest.webmanifest`.
- Node.js - required for Vite development/build/test commands in `package.json`; no `engines` field is declared in `package.json`. Local toolchain observed: Node.js v22.21.1.
- Go native HTTP server - Gin server entry point in `backend-go/main.go`; `backend-go/go.mod` declares Go 1.25.0. Local toolchain observed: go1.25.10 linux/amd64.
- npm 11.15.0 - root frontend package manager.
- Lockfile: present at `package-lock.json`.
- Go modules - backend dependency management via `backend-go/go.mod` and `backend-go/go.sum`.
## Frameworks
- React 19.1.0 - SPA rendering and component model in `src/main.tsx`, `src/App.tsx`, `src/components/**`, and `src/admin/**`.
- React DOM 19.1.0 - browser root mounting in `src/main.tsx`.
- Vite 6.3.2 - frontend dev server and static build, configured in `vite.config.ts`.
- Gin 1.10.1 - Go HTTP routing, middleware, and JSON APIs in `backend-go/main.go` and `backend-go/handler/**`.
- GORM 1.30.0 with SQLite driver 1.6.0 - persistence layer in `backend-go/database/database.go` and models in `backend-go/database/models.go`.
- Zustand 5.0.5 - frontend state management and persistence in `src/store.ts`.
- Tailwind CSS 3.4.17 - utility-first styling configured in `tailwind.config.js` and consumed from `src/index.css`.
- Radix UI primitives - dialog, alert dialog, dropdown, label, popover, scroll area, select, separator, switch, tabs, and tooltip components under `src/components/ui/**`.
- shadcn-style component setup - UI aliases and Tailwind integration declared in `components.json`; generated/maintained UI components live in `src/components/ui/**`.
- Vitest 4.1.5 - frontend unit tests under `src/**/*.test.ts` and `src/**/*.test.tsx`; commands defined in `package.json`.
- Go standard `testing` package - backend tests under `backend-go/**/*_test.go`.
- In-memory / temp SQLite testing - backend tests use `gorm.io/driver/sqlite` in files such as `backend-go/database/models_test.go` and `backend-go/handler/auth_handler_test.go`.
- TypeScript build mode - `npm run build` executes `tsc -b && vite build` from `package.json`.
- @vitejs/plugin-react 4.4.1 - React transform plugin configured in `vite.config.ts`.
- PostCSS 8.5.3 and Autoprefixer 10.4.21 - CSS processing configured in `postcss.config.js`.
- tailwindcss-animate 1.0.7 - Tailwind animation plugin loaded from `tailwind.config.js`.
- Vite dev proxy - optional local proxy loaded by `vite.config.ts` through `src/lib/devProxy.ts` and `dev-proxy.config.json`.
## Key Dependencies
- `github.com/openai/openai-go/v3` 3.34.0 - OpenAI-compatible image generation/edit client in `backend-go/service/openai.go`.
- `github.com/gin-gonic/gin` 1.10.1 - backend routing and request handling in `backend-go/main.go` and `backend-go/handler/**`.
- `gorm.io/gorm` 1.30.0 and `gorm.io/driver/sqlite` 1.6.0 - SQLite ORM layer in `backend-go/database/database.go`.
- `github.com/golang-jwt/jwt/v5` 5.2.2 - HS256 JWT signing and verification in `backend-go/service/auth.go` and `backend-go/middleware/middleware.go`.
- `golang.org/x/crypto` 0.52.0 - bcrypt password hashing in `backend-go/service/auth.go`.
- `zustand` 5.0.5 - persisted frontend app state in `src/store.ts`.
- `react` 19.1.0 and `react-dom` 19.1.0 - core UI runtime in `src/main.tsx` and `src/App.tsx`.
- `github.com/gin-contrib/cors` 1.7.6 - permissive CORS middleware configured in `backend-go/main.go`.
- `sonner` 2.0.1 - toast notifications wired through `src/store.ts` and rendered by `src/components/ui/sonner.tsx`.
- `lucide-react` 0.468.0 - icon set used by React components in `src/components/**` and `src/admin/**`.
- `class-variance-authority`, `clsx`, and `tailwind-merge` - class composition utilities for UI components in `src/components/ui/**` and `src/lib/utils.ts`.
- Browser IndexedDB API - image persistence helper in `src/lib/db.ts`.
- Browser Cache API and Service Worker API - PWA app-shell caching in `public/sw.js`.
- Browser `localStorage` - auth token storage in `src/lib/backendApi.ts` and `src/admin/adminApi.ts`; Zustand persistence key configured in `src/store.ts`.
## Configuration
- Frontend backend URL is configured with optional `VITE_BACKEND_URL`; it is read in `src/lib/backendApi.ts`, `src/admin/adminApi.ts`, and `src/store.ts`, with a development default pointing at the local Go backend.
- Vite built-ins `import.meta.env.PROD` and `import.meta.env.BASE_URL` control production-only service worker registration in `src/main.tsx`.
- Backend configuration is file-based, not environment-variable-based. `backend-go/config/config.go` loads `backend-go/config.json` from the backend working directory.
- `backend-go/config.json` is present and ignored by git according to `.gitignore`; treat it as the secrets/config store for `jwtSecret`, `adminApikey`, `apiEndpoints[].apiKey`, endpoint base URLs, model, API mode, timeout, pricing, and invite settings.
- `.env`, `.env.*`, and `*.env` files are not detected in the repository scan.
- `dev-proxy.config.json` is present and ignored by git according to `.gitignore`; `vite.config.ts` reads it only for `vite serve` and uses `src/lib/devProxy.ts` to normalize proxy settings.
- Project-specific skill directories are not detected at `.claude/skills/` or `.agents/skills/`.
- Frontend scripts in `package.json`: `npm run dev` starts Vite, `npm run build` runs TypeScript build plus Vite build, `npm run preview` serves the production build, and `npm test` runs Vitest once.
- TypeScript compiler options in `tsconfig.json`: `target` ES2020, DOM libs enabled, `moduleResolution` `bundler`, `jsx` `react-jsx`, and `strict` mode enabled.
- Vite configuration in `vite.config.ts`: React plugin, relative `base: './'`, optional dev proxy, and `__DEV_PROXY_CONFIG__` define injection.
- Tailwind configuration in `tailwind.config.js`: class-based dark mode, content paths `index.html` and `src/**/*.{js,ts,jsx,tsx}`, CSS-variable color tokens, and `tailwindcss-animate` plugin.
- shadcn-style aliases in `components.json`: `components` -> `src/components`, `utils` -> `src/lib/utils`, `ui` -> `src/components/ui`, `lib` -> `src/lib`, and `hooks` -> `src/hooks`.
- Backend dependency and build configuration lives in `backend-go/go.mod` and `backend-go/go.sum`; no Makefile, Dockerfile, or CI workflow files are detected in the repository scan.
## Platform Requirements
- Install frontend dependencies with npm from the repository root containing `package.json` and `package-lock.json`.
- Run frontend dev server with `npm run dev`; Vite serves `index.html` and routes `/admin` to the lazy-loaded admin bundle in `src/main.tsx`.
- Use Go 1.25.x for backend development from `backend-go/`; the backend entry point is `backend-go/main.go`.
- Provide a local `backend-go/config.json` before using authenticated/admin/OpenAI-backed flows; the schema is defined by `backend-go/config/config.go`.
- Runtime directories are created by `backend-go/util/paths.go`: `backend-go/data/` for SQLite and `backend-go/upload/` for image files.
- Frontend builds to static assets in `dist/` via `npm run build`; `README.md` documents static hosting options, but no deployment manifest is detected in the repository scan.
- Backend runs as a Gin HTTP server on the configured port from `backend-go/config/config.go` and `backend-go/config.json`.
- Persist backend runtime storage: SQLite files under `backend-go/data/` and uploaded/generated images under `backend-go/upload/`.
- Configure at least one OpenAI-compatible endpoint in `apiEndpoints` for image generation; endpoint scheduling and failover are implemented in `backend-go/service/openai.go` and `backend-go/service/queue.go`.
- No external cache, object store, queue service, CI workflow, Dockerfile, or hosted deployment configuration is detected in the repository scan.
<!-- GSD:stack-end -->

<!-- GSD:conventions-start source:CONVENTIONS.md -->
## Conventions

## Naming Patterns
- Frontend React component files use PascalCase with `.tsx`: `src/App.tsx`, `src/components/LoginModal.tsx`, `src/components/InputBar.tsx`, `src/components/ui/dialog.tsx` is the exception for shadcn/Radix primitives where lowercase filenames are used.
- Frontend non-component modules use lowercase or camelCase `.ts`: `src/store.ts`, `src/types.ts`, `src/lib/backendApi.ts`, `src/lib/maskPreprocess.ts`, `src/admin/moneyFormat.ts`.
- Frontend tests are co-located and named `*.test.ts` or `*.test.tsx`: `src/store.test.ts`, `src/lib/backendApi.test.ts`, `src/admin/AdminDashboard.test.tsx`.
- Backend packages and source files use lowercase package directories and lowercase Go filenames: `backend-go/service/auth.go`, `backend-go/handler/generate.go`, `backend-go/database/models.go`.
- Backend test files use Go's `_test.go` suffix and often include the feature under test in the filename: `backend-go/service/openai_failover_test.go`, `backend-go/handler/generate_billing_test.go`, `backend-go/database/models_test.go`.
- React components are PascalCase functions or constants and are default exported for page/modal-style components: `src/components/LoginModal.tsx`, `src/components/RegisterModal.tsx`, `src/admin/AdminPage.tsx`.
- Custom hooks use the `useX` prefix: `src/hooks/useCloseOnEscape.ts`, local helpers such as `useIsMobile` in `src/components/InputBar.tsx`.
- Frontend utility and API functions use camelCase verbs: `buildUrl`, `request`, `loginWithPassword`, `streamTaskStatus` in `src/lib/backendApi.ts`; `adminUpdatePricingConfig`, `adminGetBillingSummary` in `src/admin/adminApi.ts`.
- Zustand state actions use imperative camelCase names: `setAuthUser`, `setSettings`, `showToast`, `clearMaskDraft`, `markAnnouncementSeen` in `src/store.ts`.
- Backend exported handlers/services use PascalCase when called across packages: `AuthLogin`, `AuthMe`, `GenerateImage` in `backend-go/handler/*.go`; `LoginWithPassword`, `RegisterUser`, `CheckQuotaAndCreateTask` in `backend-go/service/*.go`.
- Backend package-local helpers use lower camelCase: `failTask`, `saveGeneratedImagesWithAttribution`, `buildBillingInput` in `backend-go/handler/generate.go`.
- Go test helpers use descriptive names and call `t.Helper()`: setup helpers in `backend-go/service/auth_test.go`, `backend-go/service/billing_test.go`, and handler setup helpers in `backend-go/handler/admin_handler_test.go`.
- TypeScript variables, props, state, and event handlers use camelCase: `authUser`, `seenAnnouncementUpdatedAt`, `handlePasswordLogin`, `confirmPassword` in `src/components/LoginModal.tsx` and `src/components/MigrationModal.tsx`.
- React state setters use `setX`: `setLoading`, `setError`, `setShowChangelog` in `src/components/LoginModal.tsx` and `src/App.tsx`.
- Fixed module constants use UPPER_SNAKE_CASE: `API_BASE_URL`, `ADMIN_TOKEN_KEY` in `src/admin/adminApi.ts`; `POLL_INTERVAL` in `src/store.ts`.
- TypeScript object/DTO fields mirror backend JSON names. Use camelCase for application fields (`usedCount`, `unlimitedQuota`, `needsMigration`) and preserve API-required snake_case fields inside request parameter types (`output_format`, `output_compression`) in `src/types.ts`.
- Go local variables are short only when scope is small (`c`, `r`, `db`, `err`); use descriptive names for persisted values such as `updatedUser`, `needsMigration`, `resetUsedCount` in `backend-go/handler/auth.go` and `backend-go/handler/admin.go`.
- Go database flags represented as integers use explicit conversion at the edge; `UnlimitedQuota int` is stored on `database.User` in `backend-go/database/models.go`, while API DTOs expose booleans through service/handler responses.
- TypeScript interfaces and types use PascalCase: `AppSettings`, `TaskParams`, `TaskRecord`, `StoredImage`, `Announcement`, `AdminUser`, `PricingConfigResponse` in `src/types.ts` and `src/admin/adminApi.ts`.
- Prefer union literal types for constrained values: `AnalyticsRange = 'today' | '7d' | '30d' | 'all'` in `src/admin/adminApi.ts`; task status and settings mode unions in `src/types.ts`.
- Type-only imports use `import type` when only compile-time symbols are needed: `import type { TaskRecord } from './types'` in `src/store.test.ts` and `import type { Announcement, BugFeedback } from '../types'` in `src/admin/adminApi.ts`.
- Go structs use PascalCase names and explicit `json`/`gorm` tags: `User`, `RedemptionCode`, `Task`, `BillingRecord` in `backend-go/database/models.go`.
- Go request bodies in handlers are usually local anonymous structs with JSON tags near the endpoint that consumes them: `AuthLogin`, `AuthRegister`, `AuthChangePassword` in `backend-go/handler/auth.go`.
## Code Style
- Frontend TypeScript/TSX uses single quotes, no semicolons, two-space indentation, and trailing commas for multiline calls/objects. Match examples in `src/App.tsx`, `src/store.ts`, `src/lib/backendApi.ts`, and `src/admin/adminApi.ts`.
- JSX favors Tailwind utility strings directly on elements. Shared primitive components merge classes through `cn` from `src/lib/utils.ts`; use this for reusable UI primitives in `src/components/ui/*.tsx`.
- shadcn/Radix UI primitives use `React.forwardRef`, `displayName`, `Slot`, `class-variance-authority`, and `VariantProps`; follow `src/components/ui/button.tsx` and `src/components/ui/dialog.tsx` for new primitives.
- Go code should be `gofmt` formatted, with standard library imports first, internal project imports second, and third-party imports last. Follow `backend-go/handler/auth.go` and `backend-go/main.go`.
- Money values use fixed-point integer arithmetic (`x10000`) and string/integer conversion helpers. Do not use `parseFloat` for money logic; use `src/admin/moneyFormat.ts` and `backend-go/service/money.go`.
- No ESLint config file is detected in the project root; rely on TypeScript, Vitest, and code review conventions for frontend style.
- No Prettier config file is detected in the project root; preserve the existing handwritten formatting style in nearby files.
- TypeScript strict mode is enabled in `tsconfig.json`, but `noUnusedLocals` and `noUnusedParameters` are disabled. Do not treat unused locals as acceptable in new code unless they are intentionally reserved for compatibility.
- `tsconfig.json` uses `moduleResolution: "bundler"`, `verbatimModuleSyntax: true`, `jsx: "react-jsx"`, `noFallthroughCasesInSwitch: true`, and `noUncheckedSideEffectImports: true`; write imports and side-effect modules accordingly.
- Backend linting is standard Go compiler/test enforcement. Use `go test ./...` under `backend-go/` as the primary backend quality gate.
## Import Organization
- Use relative imports in application code. `tsconfig.json` does not define `paths`, so avoid alias imports such as `@/components/...` in new code.
- `components.json` declares shadcn aliases (`components`, `utils`, `ui`, `lib`, `hooks`), but these are not mirrored into TypeScript resolution. Treat them as generator metadata, not usable import paths.
- Go imports use the module path `gpt-image-playground/backend/...` for cross-package imports, as shown in `backend-go/handler/auth.go` and `backend-go/main.go`.
## Error Handling
- Frontend API clients centralize HTTP handling in request wrappers. Use `request<T>` in `src/lib/backendApi.ts` and `adminRequest<T>` in `src/admin/adminApi.ts` for authenticated JSON calls.
- Frontend request wrappers should parse backend errors as JSON first (`payload.error` or `payload.message`), fall back to response text, and throw `new Error(message)`. Match `src/lib/backendApi.ts` and `src/admin/adminApi.ts`.
- UI form handlers should catch unknown errors with `err instanceof Error ? err.message : String(err)` and display them through local error state or `showToast`. Follow `src/components/LoginModal.tsx`, `src/components/RegisterModal.tsx`, and `src/store.ts`.
- Public frontend reads that are non-critical may degrade gracefully to `null` or empty arrays. Use this pattern only for optional content such as public announcement/changelog/config reads in `src/lib/backendApi.ts`.
- Long-running frontend task flows should update local task state immediately, then transition to done/error through SSE or polling. Use `submitTask`, `streamTaskStatus`, and task update logic in `src/store.ts`.
- Backend handlers validate request bodies at the edge and return `gin.H{"error": message}` with an appropriate HTTP status. Follow `backend-go/handler/auth.go`, `backend-go/handler/admin.go`, and `backend-go/handler/generate.go`.
- Backend services return explicit errors with user-facing Chinese messages for validation/auth failures and wrap lower-level failures with `fmt.Errorf` where context is useful. Follow `backend-go/service/auth.go`, `backend-go/service/openai.go`, and `backend-go/service/task.go`.
- Backend async generation should persist task failure state instead of only returning errors to the original request. Use `failTask` and billing/task persistence helpers in `backend-go/handler/generate.go`.
## Logging
- Frontend: minimal direct console logging; use user-visible toasts or local error state for expected failures.
- Backend: Go `log/slog`, initialized through `backend-go/log/log.go` and used by handlers/middleware.
- Backend request logging belongs in middleware. `backend-go/middleware/logger.go` emits `slog.Info("request", ...)` with request metadata.
- Backend handlers log expected operational failures at warning level with contextual fields such as `user_id`, `username`, and `error`: `backend-go/handler/auth.go`, `backend-go/handler/admin.go`.
- Do not log secrets, bearer tokens, passwords, invite codes, API keys, or raw image data. This applies to API clients in `src/lib/backendApi.ts`, admin flows in `src/admin/adminApi.ts`, and backend config/auth code in `backend-go/config/config.go` and `backend-go/service/auth.go`.
- Frontend `console.error` is reserved for exceptional bootstrapping or debugging paths, such as service worker registration in `src/main.tsx` and image/modal failures in `src/components/DetailModal.tsx`.
## Comments
- Add comments for non-obvious user interaction constraints, browser workarounds, persistence behavior, or money precision rules. Examples: unclosable migration modal in `src/components/MigrationModal.tsx`, money precision warning in `src/admin/moneyFormat.ts`, and task/image cache behavior in `src/store.ts`.
- Keep comments close to the code they explain. Prefer short explanatory comments over restating function names.
- Chinese comments and user-facing Chinese messages are part of the existing style in auth, UI, and generation flows. Preserve the language style of nearby code.
- Do not leave TODO/FIXME comments without a concrete issue or follow-up plan. Existing skipped tests in `src/store.test.ts` demonstrate an area that requires explicit maintenance before re-enabling.
- Use JSDoc/TSDoc for exported helpers where units, invariants, or precision matter. `src/admin/moneyFormat.ts` documents fixed-point money behavior and should be the model for similar helpers.
- Most local React components and backend handlers do not use doc comments; keep them self-explanatory through names and local structure.
- Go exported symbols may omit comments in this codebase, but add comments when a symbol's behavior is not obvious or when it encodes a cross-package invariant.
## Function Design
- Prefer small pure helpers for reusable logic: mask helpers in `src/lib/mask.ts`, viewport geometry in `src/lib/viewportTransform.ts`, money helpers in `src/admin/moneyFormat.ts` and `backend-go/service/money.go`.
- Large UI orchestration components exist (`src/components/InputBar.tsx`, `src/App.tsx`) but new logic should be extracted into hooks, utilities, or store actions when it can be tested separately.
- Backend handlers should remain thin: bind/validate input, call `service`, log, return JSON. Put business logic in `backend-go/service/*.go`.
- Backend generation orchestration in `backend-go/handler/generate.go` uses helper functions for persistence, attribution, and billing. Follow that extraction pattern when adding generation-side behavior.
- TypeScript API functions should accept explicit primitive/domain parameters and build JSON bodies internally: `loginWithPassword(username, password)`, `adminUpdateQuota(userId, delta, resetUsedCount, mode)` in `src/lib/backendApi.ts` and `src/admin/adminApi.ts`.
- Use object payload types when data is already a domain DTO or when many fields travel together: `ChangelogEntryPayload` in `src/types.ts` and `src/admin/adminApi.ts`.
- Go handlers should bind JSON into local structs and pass validated values to service functions. Do not pass `*gin.Context` into service packages.
- Go service functions should accept IDs/labels and domain values, not HTTP objects. Follow `service.ChangePassword(user.ID, old, new)` from `backend-go/handler/auth.go` into `backend-go/service/auth.go`.
- Frontend request functions return typed `Promise<T>` wrappers around backend JSON and throw on non-2xx responses: `src/lib/backendApi.ts`, `src/admin/adminApi.ts`.
- Frontend optional public reads may return `Promise<T | null>` or fallback arrays when the UI can continue without the data: public announcement/changelog functions in `src/lib/backendApi.ts`.
- Zustand actions generally mutate state and return `void`/`Promise<void>`, except helper functions such as `ensureImageCached` and `getCachedImage` in `src/store.ts`.
- Backend services use Go's `(value, error)` convention. For mutations with no return payload, return only `error` or a small DTO plus `error` as shown in `backend-go/service/auth.go` and `backend-go/service/billing.go`.
- Backend handlers return JSON envelopes consistently: success uses `gin.H{"ok": true}` or typed payloads; failure uses `gin.H{"error": err.Error()}`.
## Module Design
- `src/types.ts` is the central place for shared frontend domain types. Add cross-component DTOs and persisted shape types there instead of duplicating interfaces.
- `src/lib/backendApi.ts` owns user-facing backend API calls and token storage for `gpt-image-playground-token`.
- `src/admin/adminApi.ts` owns admin API DTOs/calls and token storage for `gpt-image-playground-admin-token`.
- `src/store.ts` owns app state, persistence, image cache, task lifecycle, and backend session bootstrapping. Avoid introducing competing global stores.
- `src/lib/*.ts` modules should contain pure or browser utility code that can be tested independently: `src/lib/mask.ts`, `src/lib/maskPreprocess.ts`, `src/lib/db.ts`, `src/lib/viewportTransform.ts`.
- Backend package boundaries are conventional: `backend-go/handler` for HTTP, `backend-go/service` for business logic and external APIs, `backend-go/database` for GORM models/connection, `backend-go/middleware` for auth/logging, `backend-go/config` for runtime configuration.
- No broad frontend barrel export pattern is detected. Import directly from the module that owns the symbol, such as `../types`, `./lib/backendApi`, or `./components/ui/button`.
- Do not add new barrel files unless they solve a concrete import problem; direct imports make ownership clearer in this codebase.
- Backend Go packages naturally export by package name; do not create pass-through packages for handler/service/database symbols.
<!-- GSD:conventions-end -->

<!-- GSD:architecture-start source:ARCHITECTURE.md -->
## Architecture

## System Overview
```text
```
## Component Responsibilities
| Component | Responsibility | File |
|-----------|----------------|------|
| Vite entry | Mounts React, installs mobile viewport guards, registers/unregisters service worker, chooses home vs admin route by `window.location.pathname`. | `src/main.tsx` |
| Home app shell | Initializes store, applies theme, opens changelog, renders header/search/grid/input/modals. | `src/App.tsx` |
| Admin shell | Gates `/admin` behind stored admin token and lazy-loads the dashboard. | `src/admin/AdminPage.tsx` |
| Admin dashboard | Owns admin tabs for users, codes, API endpoints, pricing, analytics, announcements, feedback, changelog, invites. | `src/admin/AdminDashboard.tsx` |
| Home components | Render task cards, grid selection, input bar, image detail/lightbox, settings, masks, auth, announcements, changelog, feedback. | `src/components/*.tsx` |
| UI primitives | Wrap Radix/shadcn-style primitives and shared visual components. | `src/components/ui/*.tsx` |
| Client state/actions | Centralizes Zustand state, persistent settings/session, image memory cache, task submission, SSE/polling, local/remote task mutations. | `src/store.ts` |
| Public API client | Encapsulates authenticated user API calls, token storage, task/image endpoints, public config, SSE stream parsing. | `src/lib/backendApi.ts` |
| Admin API client | Encapsulates admin token storage and `/api/admin/*` calls. | `src/admin/adminApi.ts` |
| Browser image store | Stores image data URLs in IndexedDB and hashes data URLs for client-side deduplication. | `src/lib/db.ts` |
| Backend bootstrap/router | Loads config, initializes logger/runtime dirs/database, wires middleware and all routes, starts Gin server. | `backend-go/main.go` |
| Auth/admin middleware | Verifies JWTs, loads active users, places `AuthUser` and `userID` in Gin context, protects admin endpoints. | `backend-go/middleware/middleware.go` |
| Request logger | Logs method/path/status/latency/IP/user ID after each Gin request. | `backend-go/middleware/logger.go` |
| HTTP handlers | Bind/validate request payloads, read path/query params, call services, return JSON/files/SSE. | `backend-go/handler/*.go` |
| Domain services | Own auth, quota, task persistence, image filesystem operations, OpenAI calls, endpoint limiting, billing, analytics, announcements, changelog, feedback. | `backend-go/service/*.go` |
| Database layer | Owns GORM connection, AutoMigrate, bootstrap admin/announcement rows, and persisted model definitions. | `backend-go/database/database.go`, `backend-go/database/models.go` |
| Runtime config | Loads `config.json`, exposes public config values, protects endpoint pool with mutexes, persists endpoint/pricing/invite config changes. | `backend-go/config/config.go` |
| Utilities | Generate IDs, hash/encrypt API keys, and constrain upload paths to the upload root. | `backend-go/util/*.go` |
| PWA shell cache | Caches static shell assets and bypasses `/api/*` requests. | `public/sw.js` |
## Pattern Overview
- The frontend uses a composition shell: `src/main.tsx` selects `src/App.tsx` or `src/admin/AdminPage.tsx`; feature screens compose components from `src/components/`, `src/admin/`, and `src/components/ui/`.
- The frontend uses a centralized action store: business actions such as submission, upload, reuse, deletion, polling, and SSE updates live in `src/store.ts`; components call store actions instead of owning workflow state.
- The backend uses explicit layers: `backend-go/main.go` declares routes, `backend-go/handler/*.go` handles HTTP concerns, `backend-go/service/*.go` handles domain logic, and `backend-go/database/*.go` handles GORM models/connection.
- Long-running image generation is asynchronous: `backend-go/handler/generate.go` creates a queued task, starts a goroutine, updates persisted task state, and `backend-go/handler/tasks.go` streams task status by polling SQLite.
- External image generation uses an endpoint pool with priority, failover, per-endpoint concurrency limits, and cost attribution in `backend-go/service/openai.go` and `backend-go/service/queue.go`.
## Layers
- Purpose: Mount the React app, choose home or admin UI, manage service worker lifecycle, and apply global CSS.
- Location: `index.html`, `src/main.tsx`, `src/index.css`, `public/sw.js`, `public/manifest.webmanifest`
- Contains: HTML root, React root creation, service worker registration, PWA metadata, Tailwind base/theme CSS.
- Depends on: `react`, `react-dom`, `src/App.tsx`, `src/admin/AdminPage.tsx`, `src/lib/viewport.ts`.
- Used by: Browser navigation to `/` and `/admin`.
- Purpose: Render the image generation workspace, task list, image viewers, mask editor, settings/auth dialogs, announcements, changelog, and feedback UI.
- Location: `src/App.tsx`, `src/components/*.tsx`
- Contains: Presentational and interaction-heavy React components such as `src/components/InputBar.tsx`, `src/components/TaskGrid.tsx`, `src/components/TaskCard.tsx`, `src/components/DetailModal.tsx`, `src/components/MaskEditorModal.tsx`.
- Depends on: `src/store.ts`, `src/types.ts`, `src/lib/*.ts`, `src/components/ui/*.tsx`.
- Used by: `src/App.tsx`.
- Purpose: Render management screens for users, redemption codes, endpoints, pricing, billing analytics, announcement, feedback, changelog, password reset, and invite settings.
- Location: `src/admin/AdminPage.tsx`, `src/admin/AdminLogin.tsx`, `src/admin/AdminDashboard.tsx`, `src/admin/adminApi.ts`, `src/admin/moneyFormat.ts`
- Contains: Admin route gate, login form, dashboard tabs, admin-specific API wrapper and money formatting helpers.
- Depends on: `src/store.ts` for theme/toasts, `src/admin/adminApi.ts` for HTTP, `src/components/ui/*.tsx` for controls.
- Used by: `src/main.tsx` when path is `/admin` or `/admin/*`.
- Purpose: Own application state, persisted user settings/session, image cache, task lifecycle, SSE/polling fallback, and high-level operations.
- Location: `src/store.ts`
- Contains: Zustand state shape, `imageCache`, `activeStreams`, polling timer, `initStore`, `bootstrapBackendSession`, `submitTask`, `executeTask`, `reuseConfig`, `editOutputs`, `removeTask`, `addImageFromFile`.
- Depends on: `src/lib/backendApi.ts`, `src/lib/db.ts`, `src/lib/canvasImage.ts`, `src/lib/mask.ts`, `src/lib/size.ts`, `src/types.ts`.
- Used by: `src/App.tsx`, `src/components/*.tsx`, `src/admin/*.tsx`.
- Purpose: Keep browser storage, network calls, image manipulation, masking, sizing, clipboard, viewport, and parameter display reusable.
- Location: `src/lib/*.ts`, `src/lib/*.tsx`, `src/admin/adminApi.ts`
- Contains: `fetch` wrappers, auth/admin token helpers, IndexedDB helpers, image canvas helpers, mask preprocessing, size normalization, viewport transforms, clipboard helpers.
- Depends on: Browser APIs (`fetch`, `localStorage`, `indexedDB`, canvas, clipboard), `src/types.ts`.
- Used by: `src/store.ts`, `src/components/*.tsx`, `src/admin/*.tsx`, `vite.config.ts`.
- Purpose: Register all backend routes, apply CORS/request logging/auth/admin middleware, configure multipart limits.
- Location: `backend-go/main.go`, `backend-go/middleware/*.go`
- Contains: Gin router groups for `/api/auth`, `/api/config`, `/api/images`, `/api/tasks`, `/api/generate`, `/api/edit`, `/api/feedback`, `/api/admin/*`.
- Depends on: `backend-go/config`, `backend-go/database`, `backend-go/handler`, `backend-go/middleware`, `github.com/gin-gonic/gin`, `github.com/gin-contrib/cors`.
- Used by: Backend process started from `backend-go/main.go`.
- Purpose: Translate HTTP requests/responses into service calls; handlers should not own persistence conversions.
- Location: `backend-go/handler/*.go`
- Contains: Auth handlers, image upload/download/delete, task list/update/delete/SSE, generation dispatch, config, admin, announcement, changelog, feedback.
- Depends on: `backend-go/service`, `backend-go/middleware`, `backend-go/config`, `github.com/gin-gonic/gin`.
- Used by: Route declarations in `backend-go/main.go`.
- Purpose: Own domain rules, persistence workflows, external API calls, async task transitions, quota logic, billing, analytics, endpoint limiting, and filesystem image storage.
- Location: `backend-go/service/*.go`
- Contains: `service.TaskRecord` conversions, quota transaction, OpenAI-compatible client calls, endpoint failover, image save/read/delete, JWT/password auth, invitation logic, billing rows, analytics aggregation, announcement/changelog/feedback rules.
- Depends on: `backend-go/database`, `backend-go/config`, `backend-go/util`, OpenAI SDK, GORM.
- Used by: `backend-go/handler/*.go`, `backend-go/middleware/middleware.go`.
- Purpose: Store durable application records and image files.
- Location: `backend-go/database/*.go`, `backend-go/data/app.sqlite`, `backend-go/upload/<userID>/`
- Contains: GORM models for users, redemption codes, images, tasks, announcements, feedback, changelog entries, billing records; SQLite database; uploaded/generated image files.
- Depends on: SQLite/GORM and filesystem access.
- Used by: `backend-go/service/*.go`.
- Purpose: Call OpenAI-compatible image generation/edit APIs through configured endpoint pool.
- Location: `backend-go/service/openai.go`, `backend-go/service/queue.go`, `backend-go/config/config.go`
- Contains: `config.ApiEndpoint`, OpenAI client creation, failover loop, endpoint limiters, Codex CLI compatibility prompt, generated image attribution.
- Depends on: `github.com/openai/openai-go/v3`, `backend-go/config`.
- Used by: `backend-go/handler/generate.go` through `service.CallImagesGenerations*` and `service.CallImagesEdits*`.
## Data Flow
### Primary Request Path
### Admin Configuration and Analytics Path
### Authentication and Session Path
### Image Storage and Cache Path
- Use Zustand in `src/store.ts` for all shared frontend state: auth user, settings, prompt, input images, mask draft, task list, search/filter, selection, modal IDs, announcements, changelog, toasts, and confirm dialogs.
- Persist only durable client preferences/session fragments through `persist()` in `src/store.ts:291`: settings, auth user, params, dismissed Codex prompts, seen announcement timestamp, and dismissed changelog keys.
- Keep image bytes out of persisted Zustand; use `imageCache` in `src/store.ts:44`, IndexedDB in `src/lib/db.ts`, backend files in `backend-go/upload/`, and backend metadata in `backend-go/database/models.go`.
- Treat backend SQLite as authoritative for authenticated task history; `src/store.ts:443` reloads tasks from `getTasks()` after session bootstrap.
- Backend global runtime state lives in `config.App` and `config.ApiEndpoints` in `backend-go/config/config.go`, `database.DB` in `backend-go/database/database.go`, endpoint limiters in `backend-go/service/queue.go`, and the default slog logger in `backend-go/log/log.go`.
## Key Abstractions
- Purpose: Represents a generation/edit job across UI, API, service, and database layers.
- Examples: `src/types.ts:64`, `backend-go/service/models.go:103`, `backend-go/database/models.go:46`, `backend-go/service/task.go:14`
- Pattern: Frontend uses typed `TaskRecord`; backend service converts JSON-friendly `TaskRecord` to GORM `database.Task` with JSON-encoded params/image arrays and metadata.
- Purpose: Carries image generation parameters (`size`, `quality`, `output_format`, `output_compression`, `moderation`, `n`).
- Examples: `src/types.ts:27`, `backend-go/service/models.go:94`, `backend-go/service/openai.go:126`, `src/store.ts:597`
- Pattern: UI normalizes unsupported values before submission; backend passes params to OpenAI-compatible SDK calls and captures actual params returned by the provider.
- Purpose: Separate transient browser input previews, IndexedDB records, and backend-persisted image metadata/files.
- Examples: `src/types.ts:47`, `src/types.ts:92`, `backend-go/service/models.go:83`, `backend-go/database/models.go:33`, `backend-go/service/image.go:42`
- Pattern: Data URLs are local preview/cache payloads; backend image IDs identify uploaded/generated/mask files and are stored in task image ID arrays.
- Purpose: Represent authenticated user state, persisted user records, and admin list rows with quota/status fields.
- Examples: `src/lib/backendApi.ts:6`, `backend-go/service/models.go:35`, `backend-go/database/models.go:3`, `backend-go/service/auth.go:47`, `src/admin/adminApi.ts:6`
- Pattern: JWT `sub` and `role` authorize requests; middleware reloads the persisted user and rejects disabled accounts.
- Purpose: Represents an OpenAI-compatible backend endpoint with API key, priority, max concurrency, and per-image cost.
- Examples: `backend-go/config/config.go:12`, `src/admin/adminApi.ts:27`, `backend-go/service/openai.go:62`, `backend-go/service/queue.go:11`
- Pattern: Admin config updates replace a sorted endpoint pool; generation code acquires a limiter slot, tries endpoints in priority order, and stamps endpoint/cost attribution on generated images.
- Purpose: Immutable per-successful-image accounting snapshot.
- Examples: `backend-go/database/models.go:106`, `backend-go/service/billing.go:10`, `backend-go/service/analytics.go:105`, `backend-go/handler/admin.go:303`
- Pattern: Generation writes one row per saved output image with endpoint, sale, cost, revenue, profit, and user label snapshots; admin analytics aggregate these rows by range.
- Purpose: Support public announcement/changelog display and user bug/feature feedback with admin management.
- Examples: `src/types.ts:101`, `src/types.ts:110`, `src/types.ts:128`, `backend-go/database/models.go:70`, `backend-go/database/models.go:79`, `backend-go/database/models.go:93`, `backend-go/service/announcement.go`, `backend-go/service/changelog.go`, `backend-go/service/feedback.go`
- Pattern: Public read endpoints hydrate home UI at startup; admin endpoints create/update/status-change records through services.
- Purpose: Coordinate overlay visibility and close only the topmost modal on Escape.
- Examples: `src/store.ts:253`, `src/hooks/useCloseOnEscape.ts:7`, `src/components/ConfirmDialog.tsx`, `src/components/MaskEditorModal.tsx`
- Pattern: Global modal IDs/booleans live in Zustand when shared; local modal state is acceptable for component-owned overlays; `useCloseOnEscape` handles stacked Escape behavior.
## Entry Points
- Location: `index.html`
- Triggers: Browser loads the Vite-built SPA.
- Responsibilities: Provides `#root`, PWA metadata links, and module script to `/src/main.tsx`.
- Location: `src/main.tsx`
- Triggers: Vite module load from `index.html`.
- Responsibilities: Installs mobile viewport guards, handles service worker lifecycle, selects home vs admin route, and mounts React.
- Location: `src/App.tsx`
- Triggers: `src/main.tsx` when path is not `/admin`.
- Responsibilities: Initializes store, applies theme, listens for image drag prevention, displays the main workspace and global modals.
- Location: `src/admin/AdminPage.tsx`
- Triggers: `src/main.tsx` when path is `/admin` or `/admin/*`.
- Responsibilities: Checks admin token, renders admin login or dashboard, applies theme.
- Location: `backend-go/main.go`
- Triggers: Go process start in the `backend-go` module.
- Responsibilities: Loads config, initializes logging/directories/database, registers middleware/routes, and runs Gin on `config.App.Port`.
- Location: `public/sw.js`
- Triggers: `src/main.tsx:9` registers it in production.
- Responsibilities: Cache app shell/static GET responses, serve cached `index.html` for navigations, bypass `/api/*`.
- Location: `vite.config.ts`
- Triggers: `npm run dev`, `npm run build`, `npm run preview`.
- Responsibilities: Configure React plugin, relative base, dev proxy normalization from `dev-proxy.config.json`, and compile-time `__DEV_PROXY_CONFIG__` define.
## Architectural Constraints
- **Threading:** Frontend code in `src/*.tsx` and `src/store.ts` runs on the browser event loop; backend HTTP handlers in `backend-go/handler/*.go` run concurrently per request; image generation runs in goroutines started at `backend-go/handler/generate.go:79`; concurrent multi-image provider calls use goroutines in `backend-go/service/openai.go:223` and `backend-go/service/openai.go:324`.
- **SQLite concurrency:** `backend-go/database/database.go:20` enables WAL and foreign keys; `backend-go/database/database.go:25` sets `SetMaxOpenConns(1)`, so backend DB access is serialized through one open SQLite connection.
- **Global state:** `config.App` in `backend-go/config/config.go:81`, `database.DB` in `backend-go/database/database.go:15`, `config.ApiEndpoints` in `backend-go/config/config.go:20`, endpoint limiters in `backend-go/service/queue.go:15`, `imageCache`/`activeStreams` in `src/store.ts:44` and `src/store.ts:70`, and `escStack` in `src/hooks/useCloseOnEscape.ts:7` are module-level state.
- **Config secrets:** `backend-go/config.json` exists as ignored runtime config and contains backend configuration/secrets; never read or commit its contents. Use `backend-go/config/config.go` to understand schema/defaults.
- **Upload path safety:** All backend file reads should resolve relative upload paths through `backend-go/util/paths.go:29`; do not join untrusted paths directly under `backend-go/upload/`.
- **Authentication tokens:** User tokens are stored with key `gpt-image-playground-token` in `src/lib/backendApi.ts:4`; admin tokens are stored with key `gpt-image-playground-admin-token` in `src/admin/adminApi.ts:4`; do not mix user and admin API clients.
- **Circular imports:** No circular dependency chain is detected in the scanned source tree; preserve the one-way flow `components/admin → store/API/lib → backendApi/adminApi` on the frontend and `main → handler/middleware → service → database/config/util` on the backend.
- **Route ownership:** Backend routes are centralized in `backend-go/main.go`; adding a handler without adding a route in `backend-go/main.go` leaves it unreachable.
- **PWA API bypass:** `public/sw.js:27` bypasses paths beginning with `/api/`; new API routes should stay under `/api/` to avoid static shell caching.
- **Agent worktrees:** `.claude/worktrees/` contains agent scratch worktrees, not product source; edit root `src/` and `backend-go/` paths instead.
## Anti-Patterns
### Handler-to-database shortcut
### Component-level HTTP scattering
### Local-only task mutation
### Editing generated/runtime directories
## Error Handling
- Frontend API wrappers in `src/lib/backendApi.ts:34` and `src/admin/adminApi.ts:51` parse `{ error }` or `{ message }` response bodies and throw `Error` instances.
- Store workflows in `src/store.ts` catch backend/session/upload/SSE errors and call `showToast()` or mark tasks as `error`.
- Backend handlers validate request bodies with `c.ShouldBindJSON()` or explicit checks in files such as `backend-go/handler/auth.go`, `backend-go/handler/generate.go`, `backend-go/handler/admin.go`, and `backend-go/handler/feedback.go`.
- Backend services return Go `error` values with user-facing Chinese messages for validation and generic messages for persistence/provider failures.
- Async generation failures call `failTask()` in `backend-go/handler/generate.go:200`, which logs the failure and persists `status = "error"`, `error`, `finishedAt`, and `elapsed`.
- Request/operational logs use `log/slog` initialized in `backend-go/log/log.go`; request logging is applied in `backend-go/main.go:39`.
## Cross-Cutting Concerns
<!-- GSD:architecture-end -->

<!-- GSD:skills-start source:skills/ -->
## Project Skills

No project skills found. Add skills to any of: `.claude/skills/`, `.agents/skills/`, `.cursor/skills/`, `.github/skills/`, or `.codex/skills/` with a `SKILL.md` index file.
<!-- GSD:skills-end -->

<!-- GSD:workflow-start source:GSD defaults -->
## GSD Workflow Enforcement

Before using Edit, Write, or other file-changing tools, start work through a GSD command so planning artifacts and execution context stay in sync.

Use these entry points:
- `/gsd-quick` for small fixes, doc updates, and ad-hoc tasks
- `/gsd-debug` for investigation and bug fixing
- `/gsd-execute-phase` for planned phase work

Do not make direct repo edits outside a GSD workflow unless the user explicitly asks to bypass it.
<!-- GSD:workflow-end -->



<!-- GSD:profile-start -->
## Developer Profile

> Profile not yet configured. Run `/gsd-profile-user` to generate your developer profile.
> This section is managed by `generate-claude-profile` -- do not edit manually.
<!-- GSD:profile-end -->
