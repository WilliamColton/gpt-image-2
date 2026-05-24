# Codebase Structure

**Analysis Date:** 2026-05-24

## Directory Layout

```
gpt_image_playground/
├── .claude/                 # Claude/GSD tooling, commands, hooks, agent worktrees; local tooling, not product source
├── .planning/               # GSD planning/codebase documents
│   └── codebase/            # Generated codebase maps such as `ARCHITECTURE.md` and `STRUCTURE.md`
├── backend-go/              # Go backend module: Gin API, GORM/SQLite persistence, services, runtime data
│   ├── config/              # Runtime config schema/load/persist helpers
│   ├── data/                # Runtime SQLite data directory (`app.sqlite`); generated and ignored
│   ├── database/            # GORM DB initialization and persisted model definitions
│   ├── handler/             # Gin HTTP handlers and backend handler tests
│   ├── log/                 # slog initialization
│   ├── middleware/          # Gin auth/admin/request-logging middleware
│   ├── service/             # Domain logic, OpenAI integration, queueing, auth, task/image/billing services
│   ├── upload/              # Runtime per-user image file storage; generated and ignored
│   ├── util/                # Shared backend ID, crypto, and path helpers
│   ├── config.json          # Local runtime backend config/secrets; ignored, do not read or commit contents
│   ├── go.mod               # Go module manifest
│   ├── go.sum               # Go dependency lockfile
│   └── main.go              # Backend executable entry point and route table
├── dist/                    # Vite build output; generated and ignored
├── docs/                    # Repository documentation/assets
├── node_modules/            # npm dependencies; generated and ignored
├── public/                  # Static public assets copied by Vite
│   ├── manifest.webmanifest # PWA manifest
│   ├── pwa-icon.svg         # PWA/favicon icon
│   └── sw.js                # Service worker app-shell/static cache
├── src/                     # React/TypeScript frontend source
│   ├── admin/               # Admin route UI, API client, and admin helper tests
│   ├── components/          # Home workspace components and modal components
│   │   └── ui/              # Shared shadcn/Radix-style UI primitives
│   ├── hooks/               # Shared React hooks
│   ├── lib/                 # Frontend API/storage/canvas/mask/size/clipboard/viewport utilities
│   ├── App.tsx              # Home app shell
│   ├── main.tsx             # React app entry and home/admin route selector
│   ├── store.ts             # Zustand state and frontend workflows
│   ├── types.ts             # Shared frontend domain types/defaults
│   └── index.css            # Tailwind/global theme CSS
├── components.json          # shadcn/ui metadata and aliases
├── dev-proxy.config.json    # Local Vite dev proxy config; ignored in `.gitignore`
├── index.html               # Vite HTML entry and PWA metadata
├── package.json             # npm scripts and frontend dependencies
├── package-lock.json        # npm dependency lockfile
├── postcss.config.js        # PostCSS/Tailwind pipeline config
├── tailwind.config.js       # Tailwind theme/content/plugins config
├── tsconfig.json            # TypeScript compiler config
└── vite.config.ts           # Vite/React/dev-proxy build config
```

## Directory Purposes

**`src/`:**
- Purpose: Frontend application source for the browser/PWA.
- Contains: React components, Zustand store, TypeScript domain types, browser API wrappers, IndexedDB utilities, Tailwind/global CSS, and frontend tests.
- Key files: `src/main.tsx`, `src/App.tsx`, `src/store.ts`, `src/types.ts`, `src/index.css`, `src/vite-env.d.ts`.

**`src/admin/`:**
- Purpose: Admin-only frontend route and API client.
- Contains: Admin login/page/dashboard components, admin HTTP wrapper, admin money formatting utility, and admin-specific tests.
- Key files: `src/admin/AdminPage.tsx`, `src/admin/AdminLogin.tsx`, `src/admin/AdminDashboard.tsx`, `src/admin/adminApi.ts`, `src/admin/moneyFormat.ts`.
- Add admin UI code here when it is only reachable from `/admin`.

**`src/components/`:**
- Purpose: Home workspace UI and reusable app-level modal components.
- Contains: Header, search, task grid/cards, input bar, settings, auth/register/migration modals, image detail/lightbox, mask editor, announcement/changelog/feedback/help/appearance dialogs, custom `Select`.
- Key files: `src/components/InputBar.tsx`, `src/components/TaskGrid.tsx`, `src/components/TaskCard.tsx`, `src/components/DetailModal.tsx`, `src/components/MaskEditorModal.tsx`, `src/components/LoginModal.tsx`, `src/components/RegisterModal.tsx`, `src/components/SettingsModal.tsx`, `src/components/Header.tsx`.
- Add home-page feature components here unless they are generic UI primitives.

**`src/components/ui/`:**
- Purpose: Shared low-level UI primitives built around Radix/shadcn patterns and Tailwind classes.
- Contains: Buttons, dialogs, inputs, tabs, switches, cards, badges, tables, tooltips, scroll areas, alert dialogs, empty/status components, and Sonner toaster wrapper.
- Key files: `src/components/ui/button.tsx`, `src/components/ui/dialog.tsx`, `src/components/ui/alert-dialog.tsx`, `src/components/ui/input.tsx`, `src/components/ui/tabs.tsx`, `src/components/ui/sonner.tsx`.
- Add design-system primitives here; do not place feature-specific business workflows in this directory.

**`src/hooks/`:**
- Purpose: Shared React hooks.
- Contains: Cross-component behavior that depends on React lifecycle.
- Key files: `src/hooks/useCloseOnEscape.ts`.
- Add reusable hooks here when used by multiple components or when they coordinate global browser behavior.

**`src/lib/`:**
- Purpose: Frontend non-component utilities and browser integrations.
- Contains: User API client, IndexedDB image store, dev proxy config normalization, canvas/mask preprocessing, clipboard helpers, size normalization, className utility, viewport guards/transforms, parameter display helper.
- Key files: `src/lib/backendApi.ts`, `src/lib/db.ts`, `src/lib/devProxy.ts`, `src/lib/canvasImage.ts`, `src/lib/mask.ts`, `src/lib/maskPreprocess.ts`, `src/lib/size.ts`, `src/lib/viewport.ts`, `src/lib/viewportTransform.ts`, `src/lib/clipboard.ts`, `src/lib/paramDisplay.tsx`, `src/lib/utils.ts`.
- Add typed API wrappers to `src/lib/backendApi.ts`; add pure shared helpers as focused files under `src/lib/`.

**`backend-go/`:**
- Purpose: Backend API and runtime module.
- Contains: Go module files, executable entry point, config, database models/initialization, middleware, handlers, services, utilities, tests, and runtime data/upload directories.
- Key files: `backend-go/main.go`, `backend-go/go.mod`, `backend-go/config/config.go`, `backend-go/database/database.go`, `backend-go/database/models.go`.
- Add backend production code under the appropriate package directory, not at module root except for executable bootstrap changes in `backend-go/main.go`.

**`backend-go/config/`:**
- Purpose: Load defaults from code, optionally overlay `config.json`, expose thread-safe endpoint/pricing/invite config mutation, and persist runtime admin changes.
- Contains: Config struct, endpoint pool mutexes, persistence helpers, invite/pricing accessors, tests.
- Key files: `backend-go/config/config.go`, `backend-go/config/config_test.go`.
- Add config schema fields here and expose accessor/mutator functions instead of reading `config.json` elsewhere.

**`backend-go/database/`:**
- Purpose: Own GORM connection and persisted schema definitions.
- Contains: SQLite initialization, WAL/foreign key config, AutoMigrate, bootstrap rows, model structs, tests.
- Key files: `backend-go/database/database.go`, `backend-go/database/models.go`, `backend-go/database/models_test.go`.
- Add new database tables/columns to `backend-go/database/models.go` and include new models in `AutoMigrate()` in `backend-go/database/database.go`.

**`backend-go/handler/`:**
- Purpose: HTTP-facing Gin handlers.
- Contains: Admin, auth, image, task, generate/edit, config, announcement, changelog, feedback handlers and handler tests.
- Key files: `backend-go/handler/auth.go`, `backend-go/handler/admin.go`, `backend-go/handler/generate.go`, `backend-go/handler/images.go`, `backend-go/handler/tasks.go`, `backend-go/handler/config.go`, `backend-go/handler/announcement.go`, `backend-go/handler/changelog.go`, `backend-go/handler/feedback.go`.
- Add new endpoint handlers here, keep them thin, and register routes in `backend-go/main.go`.

**`backend-go/service/`:**
- Purpose: Backend domain/business logic and external API integration.
- Contains: Auth/session/user/quota/invite logic, task persistence/conversions, image filesystem persistence, OpenAI-compatible image generation, endpoint queueing/failover, billing, analytics, announcements, changelog, feedback, money formatting, service tests.
- Key files: `backend-go/service/auth.go`, `backend-go/service/task.go`, `backend-go/service/image.go`, `backend-go/service/openai.go`, `backend-go/service/queue.go`, `backend-go/service/billing.go`, `backend-go/service/analytics.go`, `backend-go/service/models.go`, `backend-go/service/announcement.go`, `backend-go/service/changelog.go`, `backend-go/service/feedback.go`, `backend-go/service/money.go`.
- Add backend business rules here rather than in handlers.

**`backend-go/middleware/`:**
- Purpose: Request-scoped middleware for Gin.
- Contains: User JWT auth, admin JWT auth, request logging, and context helpers.
- Key files: `backend-go/middleware/middleware.go`, `backend-go/middleware/logger.go`.
- Add cross-route HTTP behavior here when it must wrap multiple routes.

**`backend-go/log/`:**
- Purpose: Initialize global structured logging.
- Contains: slog text/JSON logger setup.
- Key files: `backend-go/log/log.go`.

**`backend-go/util/`:**
- Purpose: Backend utility helpers with no HTTP/business ownership.
- Contains: Secure ID generation, SHA-256/API-key encryption helpers, runtime directory and upload path helpers.
- Key files: `backend-go/util/id.go`, `backend-go/util/crypto.go`, `backend-go/util/paths.go`.
- Add utility code here only if it is package-neutral and reusable.

**`backend-go/data/`:**
- Purpose: Runtime SQLite storage.
- Contains: Generated `app.sqlite` database files at runtime.
- Key files: `backend-go/data/app.sqlite` at runtime.
- Generated: Yes.
- Committed: No, ignored by `.gitignore`.

**`backend-go/upload/`:**
- Purpose: Runtime image file storage grouped by user ID.
- Contains: User subdirectories and uploaded/generated/mask image files.
- Key files: `backend-go/upload/<userID>/<imageID>.<ext>` at runtime.
- Generated: Yes.
- Committed: No, ignored by `.gitignore`.

**`public/`:**
- Purpose: Static assets served/copied by Vite without bundling.
- Contains: PWA manifest, icon, service worker.
- Key files: `public/sw.js`, `public/manifest.webmanifest`, `public/pwa-icon.svg`.
- Add static PWA/browser assets here; keep API routes under `/api/` so `public/sw.js` bypasses them.

**`docs/`:**
- Purpose: Repository documentation and supporting images.
- Contains: Documentation files/assets.
- Key files: `docs/images/`.

**`.planning/codebase/`:**
- Purpose: GSD-generated codebase maps consumed by planning/execution commands.
- Contains: Architecture, structure, stack, integration, convention, testing, concern documents.
- Key files: `.planning/codebase/ARCHITECTURE.md`, `.planning/codebase/STRUCTURE.md`.
- Generated: Yes, by mapping commands.
- Committed: Project-dependent; `.gitignore` currently ignores `.planning/`.

**`.claude/`:**
- Purpose: Claude/GSD local tooling configuration, commands, hooks, and worktrees.
- Contains: Agent tooling files and `.claude/worktrees/*` scratch worktrees.
- Key files: `.claude/settings.json`, `.claude/commands/`, `.claude/hooks/`, `.claude/worktrees/`.
- Generated: Local/tooling.
- Committed: No, ignored by `.gitignore`.

## Key File Locations

**Entry Points:**
- `index.html`: Vite HTML entry with PWA metadata and `#root`.
- `src/main.tsx`: Frontend runtime entry, service worker management, home/admin route split.
- `src/App.tsx`: Home application shell.
- `src/admin/AdminPage.tsx`: Admin route shell and login/dashboard gate.
- `backend-go/main.go`: Backend executable entry and complete route registration.

**Configuration:**
- `package.json`: Frontend scripts (`dev`, `build`, `preview`, `test`) and dependency declarations.
- `package-lock.json`: npm lockfile.
- `tsconfig.json`: Strict TypeScript config for `src/`.
- `vite.config.ts`: Vite React plugin, relative base, dev proxy, compile-time define.
- `tailwind.config.js`: Tailwind content globs, dark mode, theme extension, animation plugin.
- `postcss.config.js`: PostCSS/Tailwind build pipeline.
- `components.json`: shadcn UI metadata/aliases.
- `dev-proxy.config.json`: Local dev proxy input used by `vite.config.ts`; ignored by `.gitignore`.
- `backend-go/go.mod`: Go module and dependency declarations.
- `backend-go/go.sum`: Go dependency lockfile.
- `backend-go/config/config.go`: Backend runtime config schema, defaults, accessors, and persistence.
- `backend-go/config.json`: Local backend runtime config/secrets; ignored by `.gitignore`; do not read contents.
- `.gitignore`: Ignore rules for dependencies, build output, runtime data, local secrets/config, tooling, planning docs.

**Core Frontend Logic:**
- `src/store.ts`: Zustand state, task lifecycle, image cache, polling/SSE fallback, session bootstrap, uploads, deletion, reuse/edit-output workflows.
- `src/types.ts`: Frontend domain types and default settings/task params.
- `src/lib/backendApi.ts`: User/public backend API client and SSE parser.
- `src/admin/adminApi.ts`: Admin backend API client.
- `src/lib/db.ts`: IndexedDB image storage and data URL hashing.
- `src/lib/canvasImage.ts`: Canvas/image helpers and mask validation.
- `src/lib/mask.ts`: Mask target validation, input image ordering, coverage classification.
- `src/lib/maskPreprocess.ts`: Mask preprocessing helpers.
- `src/lib/size.ts`: Image size normalization/formatting helpers.
- `src/lib/viewport.ts`: Mobile viewport guards installed by `src/main.tsx`.
- `src/lib/viewportTransform.ts`: Viewport transform calculations.
- `src/lib/devProxy.ts`: Dev proxy config normalization used by `vite.config.ts`.

**Core Backend Logic:**
- `backend-go/main.go`: Route table and server bootstrap.
- `backend-go/database/database.go`: SQLite/GORM initialization, runtime DB path, AutoMigrate, bootstrap data.
- `backend-go/database/models.go`: Persisted schema models.
- `backend-go/middleware/middleware.go`: Auth/admin middleware and context helper.
- `backend-go/middleware/logger.go`: Request logging middleware.
- `backend-go/handler/generate.go`: Generate/edit request handling, async execution orchestration, task success/failure persistence, billing record creation.
- `backend-go/handler/tasks.go`: Task CRUD and SSE task status stream.
- `backend-go/handler/images.go`: Authenticated image upload/download/delete HTTP layer.
- `backend-go/handler/auth.go`: User auth, registration, migration, invite code, password and username handlers.
- `backend-go/handler/admin.go`: Admin auth, user/code/config/pricing/analytics/invite/password handlers.
- `backend-go/service/openai.go`: OpenAI-compatible image generation/edit calls, failover, concurrent generation, response conversion.
- `backend-go/service/queue.go`: Per-endpoint concurrency limiters.
- `backend-go/service/task.go`: Task model conversions, CRUD, quota transaction, pending-image counting.
- `backend-go/service/image.go`: Image deduplication, filesystem save/read/delete, data URL conversion.
- `backend-go/service/auth.go`: JWT/password auth, user quota/status, redemption codes, invite flows, admin password reset.
- `backend-go/service/billing.go`: Immutable billing record creation.
- `backend-go/service/analytics.go`: Billing summary/trend/endpoint/user aggregations.
- `backend-go/service/models.go`: Service-layer DTOs returned by handlers.

**Testing:**
- `src/*.test.ts`, `src/*.test.tsx`: Frontend Vitest tests co-located with source.
- `src/admin/*.test.ts`, `src/admin/*.test.tsx`: Admin API/UI/helper tests.
- `src/components/*.test.tsx`: Component source-pattern tests.
- `src/lib/*.test.ts`: Frontend library tests.
- `backend-go/**/*_test.go`: Go unit/handler/service tests co-located by backend package.
- `package.json`: Frontend test command `npm run test`.
- `backend-go/go.mod`: Backend tests run with `go test ./...` from `backend-go/`.

**Generated/Runtime:**
- `dist/`: Vite build output.
- `node_modules/`: npm installed dependencies.
- `backend-go/data/`: SQLite runtime database directory.
- `backend-go/upload/`: Runtime image files.
- `.claude/worktrees/`: Agent worktree scratch directories.

## Naming Conventions

**Files:**
- React components use PascalCase filenames: `src/components/InputBar.tsx`, `src/components/TaskGrid.tsx`, `src/admin/AdminDashboard.tsx`.
- Shared frontend utility files use camelCase filenames: `src/lib/backendApi.ts`, `src/lib/canvasImage.ts`, `src/lib/maskPreprocess.ts`, `src/admin/moneyFormat.ts`.
- UI primitive filenames use lowercase kebab-case when mirroring shadcn/Radix primitives: `src/components/ui/alert-dialog.tsx`, `src/components/ui/dropdown-menu.tsx`, `src/components/ui/status-badge.tsx`.
- Frontend tests are co-located and named `*.test.ts` or `*.test.tsx`: `src/store.test.ts`, `src/lib/db.test.ts`, `src/admin/adminApi.test.ts`, `src/components/LoginModal.test.tsx`.
- Go implementation files use lowercase snake_case or domain names: `backend-go/service/openai.go`, `backend-go/service/queue.go`, `backend-go/handler/generate.go`, `backend-go/database/models.go`.
- Go tests are co-located as `*_test.go`: `backend-go/service/openai_failover_test.go`, `backend-go/handler/admin_handler_test.go`.
- Config files use standard tool names at repo root: `vite.config.ts`, `tailwind.config.js`, `postcss.config.js`, `tsconfig.json`.

**Directories:**
- Frontend source is grouped by role: `src/admin/` for admin route, `src/components/` for app components, `src/components/ui/` for primitives, `src/hooks/` for hooks, `src/lib/` for non-component utilities.
- Backend source is grouped by Go package/layer: `backend-go/config/`, `backend-go/database/`, `backend-go/handler/`, `backend-go/middleware/`, `backend-go/service/`, `backend-go/util/`, `backend-go/log/`.
- Runtime directories stay under backend module root: `backend-go/data/` and `backend-go/upload/`.

## Where to Add New Code

**New Home Feature:**
- Primary UI: Add component(s) under `src/components/` and wire them from `src/App.tsx`, `src/components/Header.tsx`, `src/components/InputBar.tsx`, `src/components/TaskGrid.tsx`, or the owning modal/component.
- Shared state/actions: Add Zustand fields/actions and workflow logic to `src/store.ts` when multiple components need it or when it persists/mutates tasks/session/images.
- Types/defaults: Add domain types to `src/types.ts`.
- API calls: Add typed user/public wrappers to `src/lib/backendApi.ts` and call them from `src/store.ts` or the feature component.
- Tests: Add co-located `*.test.ts` or `*.test.tsx` beside the touched source, such as `src/store.test.ts`, `src/components/<Feature>.test.tsx`, or `src/lib/<helper>.test.ts`.

**New Admin Feature:**
- Primary UI: Add tab/state/render logic to `src/admin/AdminDashboard.tsx` if it belongs in the existing dashboard, or add a new component under `src/admin/` and import it into `src/admin/AdminDashboard.tsx`.
- API calls/types: Add typed wrappers and response interfaces to `src/admin/adminApi.ts`.
- Backend route: Add handler to `backend-go/handler/admin.go` or a focused `backend-go/handler/<domain>.go`, protect it with `middleware.AdminMiddleware()` in `backend-go/main.go`, and implement domain logic in `backend-go/service/`.
- Tests: Add `src/admin/*.test.ts(x)` for frontend/API logic and `backend-go/handler/*_test.go` or `backend-go/service/*_test.go` for backend behavior.

**New Backend API Endpoint:**
- Route registration: Add route in `backend-go/main.go` under the correct group (`/api/auth`, `/api/config`, `/api/images`, `/api/tasks`, `/api`, or `/api/admin`).
- Handler: Add/extend a file in `backend-go/handler/` for HTTP binding/validation/response.
- Business logic: Add/extend `backend-go/service/<domain>.go`.
- Persistence: Add/extend `backend-go/database/models.go`; include new models in `AutoMigrate()` in `backend-go/database/database.go`.
- Frontend client: Add typed wrapper to `src/lib/backendApi.ts` or `src/admin/adminApi.ts`.
- Tests: Add handler/service/model tests in matching `backend-go` package and frontend API tests if called from UI.

**New Database Model or Field:**
- Model definition: `backend-go/database/models.go`.
- Migration registration: `backend-go/database/database.go` inside `DB.AutoMigrate(...)` for new models.
- Service DTO/conversion: `backend-go/service/models.go` and domain-specific service file such as `backend-go/service/task.go`.
- API exposure: Relevant handler in `backend-go/handler/` and frontend type in `src/types.ts`, `src/lib/backendApi.ts`, or `src/admin/adminApi.ts`.
- Do not mutate SQLite files directly under `backend-go/data/`.

**New Image Generation Provider Behavior:**
- Endpoint/config schema: `backend-go/config/config.go` and admin UI/API wrappers in `src/admin/adminApi.ts` plus UI in `src/admin/AdminDashboard.tsx`.
- Provider call changes: `backend-go/service/openai.go`.
- Concurrency/failover behavior: `backend-go/service/queue.go` and `backend-go/service/openai.go`.
- Generation orchestration/billing attribution: `backend-go/handler/generate.go`, `backend-go/service/billing.go`, `backend-go/service/analytics.go`.
- Tests: `backend-go/service/openai_failover_test.go`, `backend-go/service/billing_test.go`, `backend-go/handler/generate_billing_test.go` or new focused tests.

**New Shared UI Primitive:**
- Implementation: `src/components/ui/<primitive>.tsx`.
- Usage: Import from `src/components/ui/<primitive>` into feature components.
- Styling helper: Use `src/lib/utils.ts` if class name merging is needed.
- Do not place business-specific API calls or store workflows in `src/components/ui/`.

**New Frontend Utility:**
- Pure/browser utility: Add a focused file under `src/lib/`, for example `src/lib/<utility>.ts`.
- React hook: Add under `src/hooks/` if it uses React lifecycle or state.
- API wrapper: Use `src/lib/backendApi.ts` for user/public APIs and `src/admin/adminApi.ts` for admin APIs.
- Tests: Add `src/lib/<utility>.test.ts` or `src/hooks/<hook>.test.ts` if test infrastructure supports it.

**New Backend Utility:**
- Implementation: Add to `backend-go/util/` only for package-neutral helpers.
- Domain-specific helpers: Prefer the relevant `backend-go/service/<domain>.go`.
- Tests: Add `backend-go/util/*_test.go` or service package tests.

**New Static/PWA Asset:**
- Static assets: Add to `public/`.
- PWA metadata: Update `public/manifest.webmanifest`.
- Service worker caching behavior: Update `public/sw.js`; keep API endpoints under `/api/` to preserve bypass logic in `public/sw.js:27`.

**New Styling or Theme Token:**
- Global CSS variables/base styles: `src/index.css`.
- Tailwind theme/content/plugin config: `tailwind.config.js`.
- Component-specific styling: Keep Tailwind classes in the owning component, using UI primitives from `src/components/ui/` when possible.

**New Build/Tooling Config:**
- Frontend dev/build: `vite.config.ts`, `tsconfig.json`, `package.json`, `postcss.config.js`, `tailwind.config.js`.
- Backend module/dependency config: `backend-go/go.mod`, `backend-go/go.sum`.
- shadcn metadata: `components.json`.
- Do not place local secrets in committed config; `backend-go/config.json` and `dev-proxy.config.json` are local/ignored.

## Special Directories

**`backend-go/data/`:**
- Purpose: SQLite runtime database storage.
- Generated: Yes.
- Committed: No; ignored in `.gitignore`.

**`backend-go/upload/`:**
- Purpose: Runtime image storage organized by backend user ID.
- Generated: Yes.
- Committed: No; ignored in `.gitignore`.

**`dist/`:**
- Purpose: Frontend production build output from Vite.
- Generated: Yes.
- Committed: No; ignored in `.gitignore`.

**`node_modules/`:**
- Purpose: Installed npm dependencies.
- Generated: Yes.
- Committed: No; ignored in `.gitignore`.

**`.claude/worktrees/`:**
- Purpose: Claude agent scratch worktrees.
- Generated: Yes.
- Committed: No; `.claude/` is ignored in `.gitignore`.

**`.planning/codebase/`:**
- Purpose: Codebase mapping documents generated for GSD planning/execution.
- Generated: Yes.
- Committed: No in current ignore rules because `.planning/` is ignored in `.gitignore`.

**`public/`:**
- Purpose: Static public/PWA files copied by Vite.
- Generated: No.
- Committed: Yes, except generated additions if any.

**`docs/`:**
- Purpose: Documentation and supporting images.
- Generated: No.
- Committed: Yes.

**`backend-go/.idea/`:**
- Purpose: IDE metadata under backend directory.
- Generated: Yes.
- Committed: No; `.idea/` is ignored in `.gitignore`.

---

*Structure analysis: 2026-05-24*
