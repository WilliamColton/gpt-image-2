# Coding Conventions

**Analysis Date:** 2026-05-22

## Naming Patterns

**Files:**
- React feature components use PascalCase filenames and default component exports, such as `src/App.tsx`, `src/components/TaskCard.tsx`, `src/components/InputBar.tsx`, and `src/admin/AdminDashboard.tsx`.
- Shared UI primitives under `src/components/ui/` use lowercase or kebab-case filenames with named PascalCase exports, such as `src/components/ui/button.tsx`, `src/components/ui/status-badge.tsx`, and `src/components/ui/app-dialog.tsx`.
- TypeScript utility modules use camelCase filenames that match the exported domain, such as `src/lib/maskPreprocess.ts`, `src/lib/viewportTransform.ts`, `src/lib/backendApi.ts`, and `src/lib/devProxy.ts`.
- TypeScript tests are co-located with implementation files and use `.test.ts`, such as `src/lib/mask.test.ts`, `src/lib/maskPreprocess.test.ts`, `src/lib/viewportTransform.test.ts`, `src/lib/db.test.ts`, and `src/store.test.ts`.
- Go source files use lowercase package-oriented names under `backend-go/`, such as `backend-go/service/openai.go`, `backend-go/handler/generate.go`, and `backend-go/database/models.go`; tests use `_test.go`, such as `backend-go/service/image_test.go` and `backend-go/handler/images_test.go`.

**Functions:**
- TypeScript functions and hooks use camelCase; hooks start with `use`, such as `useCloseOnEscape` in `src/hooks/useCloseOnEscape.ts` and `useIsMobile` in `src/components/InputBar.tsx`.
- React components use PascalCase function names and return JSX directly, such as `App` in `src/App.tsx`, `TaskCard` in `src/components/TaskCard.tsx`, and `ConfirmDialog` in `src/components/ConfirmDialog.tsx`.
- Store actions and utility functions use verb-first camelCase names, such as `ensureImageCached`, `warmImageContentCache`, `submitTask`, and `updateTaskLocal` in `src/store.ts`.
- Backend Go exported functions use PascalCase for cross-package APIs, such as `Load` in `backend-go/config/config.go`, `GenerateImage` in `backend-go/handler/generate.go`, and `CallImagesGenerations` in `backend-go/service/openai.go`.
- Backend Go private helpers use lower camelCase, such as `setEndpoints` in `backend-go/config/config.go`, `executeImageGeneration` in `backend-go/handler/generate.go`, and `convertImagesResponse` in `backend-go/service/openai.go`.

**Variables:**
- TypeScript locals use camelCase, such as `latestChangelog` in `src/App.tsx`, `filteredTasks` in `src/components/InputBar.tsx`, and `actualParamsByImage` in `src/store.ts`.
- TypeScript constants that represent fixed configuration use UPPER_SNAKE_CASE, such as `DEFAULT_SETTINGS` and `DEFAULT_PARAMS` in `src/types.ts`, `API_MAX_IMAGES` in `src/components/InputBar.tsx`, and `DEFAULT_MASK_WORKING_MAX_EDGE` in `src/lib/maskPreprocess.ts`.
- Module-level mutable state is explicit and narrowly named, such as `imageCache`, `imageContentFetches`, `pollTimer`, and `activeStreams` in `src/store.ts`.
- Go package globals use PascalCase only when intentionally exported, such as `App` and `ApiEndpoints` in `backend-go/config/config.go`; private synchronization variables use lower camelCase, such as `endpointsMu` and `persistMu` in `backend-go/config/config.go`.

**Types:**
- TypeScript interfaces and type aliases use PascalCase, such as `AppSettings`, `TaskParams`, `TaskRecord`, `StoredImage`, `BugFeedback`, and `ChangelogEntry` in `src/types.ts`.
- Component-local props commonly use a local `interface Props`, such as `src/components/TaskCard.tsx` and `src/admin/AdminDashboard.tsx`; shared primitive props use exported `*Props`, such as `ButtonProps` in `src/components/ui/button.tsx` and `InputProps` in `src/components/ui/input.tsx`.
- Discriminated string unions define state-like values, such as `TaskStatus`, `ThemeMode`, `BugFeedbackCategory`, and `BugFeedbackStatus` in `src/types.ts`.
- Go structs use PascalCase type names with JSON/GORM tags, such as `Config` and `ApiEndpoint` in `backend-go/config/config.go`, `Task` and `Image` in `backend-go/database/models.go`, and `GeneratedImage` in `backend-go/service/openai.go`.

## Code Style

**Formatting:**
- TypeScript and TSX use two-space indentation, single quotes, no semicolons, and trailing commas in multi-line calls/objects, as shown in `src/App.tsx`, `src/store.ts`, `src/components/ui/button.tsx`, and `vite.config.ts`.
- JSX keeps conditional rendering inline when concise and extracts longer derived state above `return`, such as `duration`, `aggregateActualParams`, and `swipeBgClass` in `src/components/TaskCard.tsx`.
- Tailwind CSS classes are written directly in `className` strings for feature UI, such as `src/components/TaskCard.tsx` and `src/components/InputBar.tsx`; shared primitives combine base classes and overrides through `cn` from `src/lib/utils.ts`.
- Use `cn(...inputs)` from `src/lib/utils.ts` for conditional class merging in reusable UI primitives; it combines `clsx` and `tailwind-merge`.
- Use `class-variance-authority` for reusable variant APIs, following `buttonVariants` in `src/components/ui/button.tsx`.
- Global CSS is reserved for fonts, browser quirks, safe-area utilities, scrollbars, and named animations in `src/index.css`; component-specific visual styling belongs in Tailwind classes near the JSX.
- Go code uses standard gofmt formatting, tab indentation, grouped imports, and explicit error checks, as shown in `backend-go/handler/images_test.go`, `backend-go/service/openai.go`, and `backend-go/config/config.go`.
- Project skill directories are not present: no `SKILL.md` files exist under `.claude/skills/` or `.agents/skills/`.

**Linting:**
- No root ESLint config is present at `.eslintrc*` or `eslint.config.*`; no Biome config is present at `biome.json`.
- No root Prettier config is present at `.prettierrc*`; preserve the existing two-space, single-quote, no-semicolon TypeScript style used in `src/App.tsx`, `src/store.ts`, and `vite.config.ts`.
- TypeScript strict checking is enabled in `tsconfig.json` with `strict: true`, `noFallthroughCasesInSwitch: true`, and `noUncheckedSideEffectImports: true`.
- `tsconfig.json` allows unused locals and unused parameters with `noUnusedLocals: false` and `noUnusedParameters: false`; do not rely on unused-variable errors as a quality gate.
- The current frontend static quality gate is `npm run build` from `package.json`, which runs `tsc -b && vite build`.
- Backend Go quality relies on `go test` and standard `gofmt` conventions for files under `backend-go/`.

## Import Organization

**Order:**
1. Framework/runtime imports first, such as React hooks in `src/App.tsx`, `src/components/InputBar.tsx`, and `src/components/ui/button.tsx`, or Go standard library imports in `backend-go/handler/images_test.go`.
2. Third-party packages next, such as `lucide-react`, Radix UI, `class-variance-authority`, and `zustand` in `src/components/TaskCard.tsx`, `src/components/ui/dialog.tsx`, and `src/store.ts`.
3. Type-only imports are marked with `import type`, such as `TaskRecord` in `src/components/TaskCard.tsx`, `StoredImage` in `src/lib/db.ts`, and `AuthUser` in `src/store.ts`.
4. Internal state, library, and type imports follow third-party imports, such as `useStore` from `src/store.ts`, helpers from `src/lib/`, and types from `src/types.ts`.
5. Local component imports come after state/library imports, such as `Header`, `SearchBar`, and modal imports in `src/App.tsx`, or UI primitive imports in `src/components/InputBar.tsx`.
6. Go imports are grouped as standard library, internal module `gpt-image-playground/backend/...`, then third-party dependencies, as shown in `backend-go/handler/images_test.go` and `backend-go/service/image_test.go`.

**Path Aliases:**
- No TypeScript path aliases are configured in `tsconfig.json`; use relative imports such as `../store`, `../types`, `./lib/backendApi`, and `../../lib/utils`.
- Do not introduce alias imports like `@/components/...` unless `tsconfig.json` and Vite resolution are updated together.
- Backend Go imports use the module path from `backend-go/go.mod`: `gpt-image-playground/backend/...`.

## Error Handling

**Patterns:**
- Frontend API wrappers centralize HTTP error parsing and throw `Error` instances from `request` in `src/lib/backendApi.ts` and `adminRequest` in `src/admin/adminApi.ts`; callers should display `err instanceof Error ? err.message : String(err)`.
- UI async handlers catch errors and route them through the store toast, as in `handleFiles` in `src/components/InputBar.tsx` and admin loaders/actions in `src/admin/AdminDashboard.tsx`.
- Public optional fetches return safe defaults instead of throwing, such as `getPublicAnnouncement`, `getLatestPublicChangelog`, and `getPublicChangelogEntries` in `src/lib/backendApi.ts`.
- Background cache and polling failures are intentionally swallowed when retry/fallback behavior exists, such as `pollRunningTasks`, `setCacheFromIdbOrRemote`, and `fetchAndCacheImage` in `src/store.ts`.
- User-facing validation throws localized `Error` messages from pure helpers when the caller needs a hard stop, such as `validateMaskTarget` and `assertUsableMaskCoverage` in `src/lib/mask.ts`.
- Use `void` or explicit `.catch(() => {})` for deliberate fire-and-forget work, such as `warmImageContentCache` and `putImage(...).catch(() => {})` in `src/store.ts`.
- Backend Gin handlers validate input at the edge and return JSON errors with appropriate HTTP status codes, such as `GenerateImage` in `backend-go/handler/generate.go`, `AuthLogin` in `backend-go/handler/auth.go`, and `ImagesUpload` in `backend-go/handler/images.go`.
- Backend services return errors with contextual messages and log unexpected persistence or API failures with `slog`, such as `ListTasks` in `backend-go/service/task.go`, `SaveImageBuffer` in `backend-go/service/image.go`, and `withFailover` in `backend-go/service/openai.go`.
- Backend startup code may panic after logging unrecoverable initialization failures in `backend-go/main.go`; request handlers should not panic for user or integration errors.

## Logging

**Framework:**
- Frontend: minimal `console.error` only for non-user-facing diagnostics in `src/main.tsx` and `src/components/DetailModal.tsx`; user-visible feedback uses the store toast and Sonner in `src/components/Toast.tsx`.
- Backend: Go `log/slog` initialized by `backend-go/log/log.go` and used throughout `backend-go/handler/`, `backend-go/service/`, `backend-go/config/`, and `backend-go/middleware/`.

**Patterns:**
- Use `useStore((s) => s.showToast)` for frontend action results, matching `src/admin/AdminDashboard.tsx` and `src/components/InputBar.tsx`.
- Backend logs should include structured key-value fields such as `user_id`, `task_id`, `image_id`, `status`, and `error`, matching `backend-go/handler/generate.go`, `backend-go/service/task.go`, and `backend-go/middleware/logger.go`.
- Request logging is middleware-owned in `backend-go/middleware/logger.go`; avoid duplicating routine request logs inside individual handlers.
- Do not log API keys, JWTs, redemption codes in user-facing paths, or full image data URLs; current structured logging in `backend-go/service/openai.go`, `backend-go/service/image.go`, and `backend-go/handler/admin.go` focuses on identifiers and error values.

## Comments

**When to Comment:**
- Use section divider comments for large state modules and type files, following `// ===== Image cache =====`, `// ===== Global polling =====`, and `// ===== Store 类型 =====` in `src/store.ts`, and section comments in `src/types.ts`.
- Use comments to explain browser behavior, async lifecycle, or non-obvious UI interactions, such as swipe selection comments in `src/components/TaskCard.tsx`, paste/drag comments in `src/components/InputBar.tsx`, and ESC stack comments in `src/hooks/useCloseOnEscape.ts`.
- Backend comments document concurrency, failover, and persistence side effects, such as `GetEndpointPool`/`SetEndpoints` in `backend-go/config/config.go`, `withFailover` in `backend-go/service/openai.go`, and task execution comments in `backend-go/handler/generate.go`.
- Keep comments close to the code they explain; avoid restating simple assignments or JSX labels.

**JSDoc/TSDoc:**
- TSDoc is used sparingly for exported fields and utility intent, such as `InputImage`, `TaskRecord`, and `StoredImage` fields in `src/types.ts`, and `streamTaskStatus` in `src/lib/backendApi.ts`.
- Prefer concise field comments for data contracts in `src/types.ts` when the property name does not fully explain storage or API behavior.
- Go exported functions may include line comments when they are package-level APIs, such as `GetEndpointPool`, `SetEndpoints`, and `persistEndpoints` in `backend-go/config/config.go`.

## Function Design

**Size:**
- Pure utility functions should stay small and single-purpose, following `clampViewTransform`, `zoomAtPoint`, and `clientPointToCanvasPoint` in `src/lib/viewportTransform.ts`, and `classifyMaskAlpha` in `src/lib/mask.ts`.
- React feature components may be large when they coordinate UI state and event handlers, such as `src/components/InputBar.tsx` and `src/admin/AdminDashboard.tsx`; keep derived values and handlers named near the state they use.
- Extract shared UI primitives into `src/components/ui/` when behavior/style repeats across feature components, following `src/components/ui/button.tsx`, `src/components/ui/dialog.tsx`, and `src/components/ui/alert-dialog.tsx`.
- Backend handlers should parse/validate requests, call services, and format responses; long-running or persistence-heavy logic belongs in `backend-go/service/`, following `backend-go/handler/generate.go` delegating to `backend-go/service/openai.go`, `backend-go/service/image.go`, and `backend-go/service/task.go`.

**Parameters:**
- Use object parameters for functions with many related values, such as `getPinchTransform(input: { ... })` in `src/lib/viewportTransform.ts`.
- Use explicit typed scalar parameters for simple helper APIs, such as `orderInputImagesForMask(inputImages, targetImageId)` in `src/lib/mask.ts` and `adminUpdateQuota(userId, delta, resetUsedCount, mode)` in `src/admin/adminApi.ts`.
- Use `Partial<T>` for state patch operations, such as `setSettings(s: Partial<AppSettings>)` and `setParams(p: Partial<TaskParams>)` in `src/store.ts`.
- Go service functions pass `userID` and resource IDs explicitly, such as `ListTasks(userID string)`, `GetTask(userID, taskID string)`, and `SaveImageBuffer(userID string, ...)` in `backend-go/service/`.

**Return Values:**
- Async frontend API functions return typed `Promise<T>` values, such as `loginWithCode`, `getTasks`, `uploadImage`, and `submitGenerateTask` in `src/lib/backendApi.ts`.
- Cache/read helpers return `undefined` or `null` for missing optional data when absence is expected, such as `getCachedImage` and `ensureImageCached` in `src/store.ts`, and `getPublicAnnouncement` in `src/lib/backendApi.ts`.
- Pure validation helpers throw for invalid state and otherwise return a typed result or `void`, such as `validateMaskTarget` and `assertUsableMaskCoverage` in `src/lib/mask.ts`.
- Go functions follow `(value, error)` or `error` return conventions, such as `Load` in `backend-go/config/config.go`, `GetTask` in `backend-go/service/task.go`, and `DataURLToBytes` in `backend-go/service/openai.go`.

## Module Design

**Exports:**
- Feature React components use default exports from their files, such as `src/App.tsx`, `src/components/ConfirmDialog.tsx`, `src/components/InputBar.tsx`, and `src/admin/AdminDashboard.tsx`.
- Shared UI primitives use named exports, such as `Button` and `buttonVariants` from `src/components/ui/button.tsx`, `Input` from `src/components/ui/input.tsx`, and `StatusBadge` from `src/components/ui/status-badge.tsx`.
- Utility modules export named pure helpers, such as `calculateMaskWorkingSize` from `src/lib/maskPreprocess.ts`, `clampViewTransform` from `src/lib/viewportTransform.ts`, and `hashDataUrl` from `src/lib/db.ts`.
- Type and constant definitions are centralized in `src/types.ts`; use exported `DEFAULT_SETTINGS` and `DEFAULT_PARAMS` instead of duplicating defaults.
- API boundary modules are separated by role: user-facing backend calls live in `src/lib/backendApi.ts`, while admin calls live in `src/admin/adminApi.ts`.
- Backend packages separate concerns by directory: HTTP handlers in `backend-go/handler/`, services in `backend-go/service/`, database models in `backend-go/database/`, config in `backend-go/config/`, middleware in `backend-go/middleware/`, and utilities in `backend-go/util/`.

**Barrel Files:**
- Barrel files are not used in `src/components/`, `src/lib/`, or `src/admin/`; import directly from the file that owns the symbol.
- `src/components/ui/` does not expose a central `index.ts`; use direct imports such as `../components/ui/button`, `../components/ui/dialog`, and `../components/ui/table`.
- Backend Go packages expose symbols through package imports rather than barrel files; import package paths directly from `gpt-image-playground/backend/...`.

---

*Convention analysis: 2026-05-22*
