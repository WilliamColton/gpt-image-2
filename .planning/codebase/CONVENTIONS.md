# Coding Conventions

**Analysis Date:** 2026-05-24

## Naming Patterns

**Files:**
- Frontend React component files use PascalCase with `.tsx`: `src/App.tsx`, `src/components/LoginModal.tsx`, `src/components/InputBar.tsx`, `src/components/ui/dialog.tsx` is the exception for shadcn/Radix primitives where lowercase filenames are used.
- Frontend non-component modules use lowercase or camelCase `.ts`: `src/store.ts`, `src/types.ts`, `src/lib/backendApi.ts`, `src/lib/maskPreprocess.ts`, `src/admin/moneyFormat.ts`.
- Frontend tests are co-located and named `*.test.ts` or `*.test.tsx`: `src/store.test.ts`, `src/lib/backendApi.test.ts`, `src/admin/AdminDashboard.test.tsx`.
- Backend packages and source files use lowercase package directories and lowercase Go filenames: `backend-go/service/auth.go`, `backend-go/handler/generate.go`, `backend-go/database/models.go`.
- Backend test files use Go's `_test.go` suffix and often include the feature under test in the filename: `backend-go/service/openai_failover_test.go`, `backend-go/handler/generate_billing_test.go`, `backend-go/database/models_test.go`.

**Functions:**
- React components are PascalCase functions or constants and are default exported for page/modal-style components: `src/components/LoginModal.tsx`, `src/components/RegisterModal.tsx`, `src/admin/AdminPage.tsx`.
- Custom hooks use the `useX` prefix: `src/hooks/useCloseOnEscape.ts`, local helpers such as `useIsMobile` in `src/components/InputBar.tsx`.
- Frontend utility and API functions use camelCase verbs: `buildUrl`, `request`, `loginWithPassword`, `streamTaskStatus` in `src/lib/backendApi.ts`; `adminUpdatePricingConfig`, `adminGetBillingSummary` in `src/admin/adminApi.ts`.
- Zustand state actions use imperative camelCase names: `setAuthUser`, `setSettings`, `showToast`, `clearMaskDraft`, `markAnnouncementSeen` in `src/store.ts`.
- Backend exported handlers/services use PascalCase when called across packages: `AuthLogin`, `AuthMe`, `GenerateImage` in `backend-go/handler/*.go`; `LoginWithPassword`, `RegisterUser`, `CheckQuotaAndCreateTask` in `backend-go/service/*.go`.
- Backend package-local helpers use lower camelCase: `failTask`, `saveGeneratedImagesWithAttribution`, `buildBillingInput` in `backend-go/handler/generate.go`.
- Go test helpers use descriptive names and call `t.Helper()`: setup helpers in `backend-go/service/auth_test.go`, `backend-go/service/billing_test.go`, and handler setup helpers in `backend-go/handler/admin_handler_test.go`.

**Variables:**
- TypeScript variables, props, state, and event handlers use camelCase: `authUser`, `seenAnnouncementUpdatedAt`, `handlePasswordLogin`, `confirmPassword` in `src/components/LoginModal.tsx` and `src/components/MigrationModal.tsx`.
- React state setters use `setX`: `setLoading`, `setError`, `setShowChangelog` in `src/components/LoginModal.tsx` and `src/App.tsx`.
- Fixed module constants use UPPER_SNAKE_CASE: `API_BASE_URL`, `ADMIN_TOKEN_KEY` in `src/admin/adminApi.ts`; `POLL_INTERVAL` in `src/store.ts`.
- TypeScript object/DTO fields mirror backend JSON names. Use camelCase for application fields (`usedCount`, `unlimitedQuota`, `needsMigration`) and preserve API-required snake_case fields inside request parameter types (`output_format`, `output_compression`) in `src/types.ts`.
- Go local variables are short only when scope is small (`c`, `r`, `db`, `err`); use descriptive names for persisted values such as `updatedUser`, `needsMigration`, `resetUsedCount` in `backend-go/handler/auth.go` and `backend-go/handler/admin.go`.
- Go database flags represented as integers use explicit conversion at the edge; `UnlimitedQuota int` is stored on `database.User` in `backend-go/database/models.go`, while API DTOs expose booleans through service/handler responses.

**Types:**
- TypeScript interfaces and types use PascalCase: `AppSettings`, `TaskParams`, `TaskRecord`, `StoredImage`, `Announcement`, `AdminUser`, `PricingConfigResponse` in `src/types.ts` and `src/admin/adminApi.ts`.
- Prefer union literal types for constrained values: `AnalyticsRange = 'today' | '7d' | '30d' | 'all'` in `src/admin/adminApi.ts`; task status and settings mode unions in `src/types.ts`.
- Type-only imports use `import type` when only compile-time symbols are needed: `import type { TaskRecord } from './types'` in `src/store.test.ts` and `import type { Announcement, BugFeedback } from '../types'` in `src/admin/adminApi.ts`.
- Go structs use PascalCase names and explicit `json`/`gorm` tags: `User`, `RedemptionCode`, `Task`, `BillingRecord` in `backend-go/database/models.go`.
- Go request bodies in handlers are usually local anonymous structs with JSON tags near the endpoint that consumes them: `AuthLogin`, `AuthRegister`, `AuthChangePassword` in `backend-go/handler/auth.go`.

## Code Style

**Formatting:**
- Frontend TypeScript/TSX uses single quotes, no semicolons, two-space indentation, and trailing commas for multiline calls/objects. Match examples in `src/App.tsx`, `src/store.ts`, `src/lib/backendApi.ts`, and `src/admin/adminApi.ts`.
- JSX favors Tailwind utility strings directly on elements. Shared primitive components merge classes through `cn` from `src/lib/utils.ts`; use this for reusable UI primitives in `src/components/ui/*.tsx`.
- shadcn/Radix UI primitives use `React.forwardRef`, `displayName`, `Slot`, `class-variance-authority`, and `VariantProps`; follow `src/components/ui/button.tsx` and `src/components/ui/dialog.tsx` for new primitives.
- Go code should be `gofmt` formatted, with standard library imports first, internal project imports second, and third-party imports last. Follow `backend-go/handler/auth.go` and `backend-go/main.go`.
- Money values use fixed-point integer arithmetic (`x10000`) and string/integer conversion helpers. Do not use `parseFloat` for money logic; use `src/admin/moneyFormat.ts` and `backend-go/service/money.go`.

**Linting:**
- No ESLint config file is detected in the project root; rely on TypeScript, Vitest, and code review conventions for frontend style.
- No Prettier config file is detected in the project root; preserve the existing handwritten formatting style in nearby files.
- TypeScript strict mode is enabled in `tsconfig.json`, but `noUnusedLocals` and `noUnusedParameters` are disabled. Do not treat unused locals as acceptable in new code unless they are intentionally reserved for compatibility.
- `tsconfig.json` uses `moduleResolution: "bundler"`, `verbatimModuleSyntax: true`, `jsx: "react-jsx"`, `noFallthroughCasesInSwitch: true`, and `noUncheckedSideEffectImports: true`; write imports and side-effect modules accordingly.
- Backend linting is standard Go compiler/test enforcement. Use `go test ./...` under `backend-go/` as the primary backend quality gate.

## Import Organization

**Order:**
1. Framework/runtime imports first: React/Vitest/browser-related packages in frontend files (`react`, `vitest`, `zustand`), standard library packages in Go files (`net/http`, `strings`, `testing`).
2. Third-party libraries next: Radix, Sonner, Gin, GORM, JWT, OpenAI SDK.
3. Internal project imports after third-party imports: `gpt-image-playground/backend/...` in Go; relative `../types`, `./lib/backendApi`, `./components/...` in TypeScript.
4. Type-only imports are separated with `import type` when possible: `src/store.test.ts`, `src/admin/adminApi.ts`, `src/lib/backendApi.test.ts`.
5. Test files import Vitest helpers first, then source modules under test, then mocked dependencies: `src/store.test.ts`, `src/lib/backendApi.test.ts`, `src/admin/adminApi.test.ts`.

**Path Aliases:**
- Use relative imports in application code. `tsconfig.json` does not define `paths`, so avoid alias imports such as `@/components/...` in new code.
- `components.json` declares shadcn aliases (`components`, `utils`, `ui`, `lib`, `hooks`), but these are not mirrored into TypeScript resolution. Treat them as generator metadata, not usable import paths.
- Go imports use the module path `gpt-image-playground/backend/...` for cross-package imports, as shown in `backend-go/handler/auth.go` and `backend-go/main.go`.

## Error Handling

**Patterns:**
- Frontend API clients centralize HTTP handling in request wrappers. Use `request<T>` in `src/lib/backendApi.ts` and `adminRequest<T>` in `src/admin/adminApi.ts` for authenticated JSON calls.
- Frontend request wrappers should parse backend errors as JSON first (`payload.error` or `payload.message`), fall back to response text, and throw `new Error(message)`. Match `src/lib/backendApi.ts` and `src/admin/adminApi.ts`.
- UI form handlers should catch unknown errors with `err instanceof Error ? err.message : String(err)` and display them through local error state or `showToast`. Follow `src/components/LoginModal.tsx`, `src/components/RegisterModal.tsx`, and `src/store.ts`.
- Public frontend reads that are non-critical may degrade gracefully to `null` or empty arrays. Use this pattern only for optional content such as public announcement/changelog/config reads in `src/lib/backendApi.ts`.
- Long-running frontend task flows should update local task state immediately, then transition to done/error through SSE or polling. Use `submitTask`, `streamTaskStatus`, and task update logic in `src/store.ts`.
- Backend handlers validate request bodies at the edge and return `gin.H{"error": message}` with an appropriate HTTP status. Follow `backend-go/handler/auth.go`, `backend-go/handler/admin.go`, and `backend-go/handler/generate.go`.
- Backend services return explicit errors with user-facing Chinese messages for validation/auth failures and wrap lower-level failures with `fmt.Errorf` where context is useful. Follow `backend-go/service/auth.go`, `backend-go/service/openai.go`, and `backend-go/service/task.go`.
- Backend async generation should persist task failure state instead of only returning errors to the original request. Use `failTask` and billing/task persistence helpers in `backend-go/handler/generate.go`.

## Logging

**Framework:**
- Frontend: minimal direct console logging; use user-visible toasts or local error state for expected failures.
- Backend: Go `log/slog`, initialized through `backend-go/log/log.go` and used by handlers/middleware.

**Patterns:**
- Backend request logging belongs in middleware. `backend-go/middleware/logger.go` emits `slog.Info("request", ...)` with request metadata.
- Backend handlers log expected operational failures at warning level with contextual fields such as `user_id`, `username`, and `error`: `backend-go/handler/auth.go`, `backend-go/handler/admin.go`.
- Do not log secrets, bearer tokens, passwords, invite codes, API keys, or raw image data. This applies to API clients in `src/lib/backendApi.ts`, admin flows in `src/admin/adminApi.ts`, and backend config/auth code in `backend-go/config/config.go` and `backend-go/service/auth.go`.
- Frontend `console.error` is reserved for exceptional bootstrapping or debugging paths, such as service worker registration in `src/main.tsx` and image/modal failures in `src/components/DetailModal.tsx`.

## Comments

**When to Comment:**
- Add comments for non-obvious user interaction constraints, browser workarounds, persistence behavior, or money precision rules. Examples: unclosable migration modal in `src/components/MigrationModal.tsx`, money precision warning in `src/admin/moneyFormat.ts`, and task/image cache behavior in `src/store.ts`.
- Keep comments close to the code they explain. Prefer short explanatory comments over restating function names.
- Chinese comments and user-facing Chinese messages are part of the existing style in auth, UI, and generation flows. Preserve the language style of nearby code.
- Do not leave TODO/FIXME comments without a concrete issue or follow-up plan. Existing skipped tests in `src/store.test.ts` demonstrate an area that requires explicit maintenance before re-enabling.

**JSDoc/TSDoc:**
- Use JSDoc/TSDoc for exported helpers where units, invariants, or precision matter. `src/admin/moneyFormat.ts` documents fixed-point money behavior and should be the model for similar helpers.
- Most local React components and backend handlers do not use doc comments; keep them self-explanatory through names and local structure.
- Go exported symbols may omit comments in this codebase, but add comments when a symbol's behavior is not obvious or when it encodes a cross-package invariant.

## Function Design

**Size:**
- Prefer small pure helpers for reusable logic: mask helpers in `src/lib/mask.ts`, viewport geometry in `src/lib/viewportTransform.ts`, money helpers in `src/admin/moneyFormat.ts` and `backend-go/service/money.go`.
- Large UI orchestration components exist (`src/components/InputBar.tsx`, `src/App.tsx`) but new logic should be extracted into hooks, utilities, or store actions when it can be tested separately.
- Backend handlers should remain thin: bind/validate input, call `service`, log, return JSON. Put business logic in `backend-go/service/*.go`.
- Backend generation orchestration in `backend-go/handler/generate.go` uses helper functions for persistence, attribution, and billing. Follow that extraction pattern when adding generation-side behavior.

**Parameters:**
- TypeScript API functions should accept explicit primitive/domain parameters and build JSON bodies internally: `loginWithPassword(username, password)`, `adminUpdateQuota(userId, delta, resetUsedCount, mode)` in `src/lib/backendApi.ts` and `src/admin/adminApi.ts`.
- Use object payload types when data is already a domain DTO or when many fields travel together: `ChangelogEntryPayload` in `src/types.ts` and `src/admin/adminApi.ts`.
- Go handlers should bind JSON into local structs and pass validated values to service functions. Do not pass `*gin.Context` into service packages.
- Go service functions should accept IDs/labels and domain values, not HTTP objects. Follow `service.ChangePassword(user.ID, old, new)` from `backend-go/handler/auth.go` into `backend-go/service/auth.go`.

**Return Values:**
- Frontend request functions return typed `Promise<T>` wrappers around backend JSON and throw on non-2xx responses: `src/lib/backendApi.ts`, `src/admin/adminApi.ts`.
- Frontend optional public reads may return `Promise<T | null>` or fallback arrays when the UI can continue without the data: public announcement/changelog functions in `src/lib/backendApi.ts`.
- Zustand actions generally mutate state and return `void`/`Promise<void>`, except helper functions such as `ensureImageCached` and `getCachedImage` in `src/store.ts`.
- Backend services use Go's `(value, error)` convention. For mutations with no return payload, return only `error` or a small DTO plus `error` as shown in `backend-go/service/auth.go` and `backend-go/service/billing.go`.
- Backend handlers return JSON envelopes consistently: success uses `gin.H{"ok": true}` or typed payloads; failure uses `gin.H{"error": err.Error()}`.

## Module Design

**Exports:**
- `src/types.ts` is the central place for shared frontend domain types. Add cross-component DTOs and persisted shape types there instead of duplicating interfaces.
- `src/lib/backendApi.ts` owns user-facing backend API calls and token storage for `gpt-image-playground-token`.
- `src/admin/adminApi.ts` owns admin API DTOs/calls and token storage for `gpt-image-playground-admin-token`.
- `src/store.ts` owns app state, persistence, image cache, task lifecycle, and backend session bootstrapping. Avoid introducing competing global stores.
- `src/lib/*.ts` modules should contain pure or browser utility code that can be tested independently: `src/lib/mask.ts`, `src/lib/maskPreprocess.ts`, `src/lib/db.ts`, `src/lib/viewportTransform.ts`.
- Backend package boundaries are conventional: `backend-go/handler` for HTTP, `backend-go/service` for business logic and external APIs, `backend-go/database` for GORM models/connection, `backend-go/middleware` for auth/logging, `backend-go/config` for runtime configuration.

**Barrel Files:**
- No broad frontend barrel export pattern is detected. Import directly from the module that owns the symbol, such as `../types`, `./lib/backendApi`, or `./components/ui/button`.
- Do not add new barrel files unless they solve a concrete import problem; direct imports make ownership clearer in this codebase.
- Backend Go packages naturally export by package name; do not create pass-through packages for handler/service/database symbols.

---

*Convention analysis: 2026-05-24*
