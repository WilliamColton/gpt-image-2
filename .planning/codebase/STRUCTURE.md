# Codebase Structure

**Analysis Date:** 2026-05-22

## Directory Layout

```
gpt_image_playground/
├── .claude/                 # Claude local workspace/memory; ignored
├── .planning/               # GSD planning artifacts and codebase maps
│   ├── codebase/            # Generated architecture/stack/quality/concern docs
│   └── phases/              # Existing phase plans
├── backend-go/              # Go/Gin backend API, services, SQLite, uploads
│   ├── config/              # Runtime config loader and endpoint pool state
│   ├── database/            # GORM initialization and persistent models
│   ├── handler/             # Gin HTTP handlers for `/api/*`
│   ├── log/                 # slog initialization
│   ├── middleware/          # Auth/admin/request logging middleware
│   ├── service/             # Backend domain logic and external API calls
│   ├── util/                # ID, hashing, and upload path helpers
│   ├── data/                # Runtime SQLite DB; ignored
│   ├── upload/              # Runtime image blobs; ignored
│   ├── config.json          # Runtime backend secrets/config; ignored, do not read
│   ├── go.mod               # Go module manifest
│   └── main.go              # Backend entry point and route registration
├── docs/                    # Static documentation assets/screenshots
│   └── images/              # README/demo image assets
├── public/                  # Static PWA assets copied by Vite
│   ├── manifest.webmanifest # PWA manifest
│   ├── pwa-icon.svg         # App icon
│   └── sw.js                # Service worker source
├── src/                     # React/Vite frontend source
│   ├── admin/               # Admin route, dashboard, and admin API client
│   ├── components/          # User app feature components and modals
│   │   └── ui/              # Shared Radix/Tailwind primitives
│   ├── hooks/               # Shared React hooks
│   ├── lib/                 # Browser utility modules and API/storage clients
│   ├── App.tsx              # User app shell
│   ├── main.tsx             # Frontend entry and route switch
│   ├── store.ts             # Zustand store and task/image orchestration
│   ├── types.ts             # Shared frontend domain types/defaults
│   └── index.css            # Tailwind globals, fonts, animations, safe areas
├── dist/                    # Vite build output; ignored/generated
├── node_modules/            # npm dependencies; ignored/generated
├── index.html               # Vite HTML host
├── package.json             # Frontend scripts and dependencies
├── package-lock.json        # npm lockfile
├── postcss.config.js        # PostCSS/Tailwind pipeline config
├── tailwind.config.js       # Tailwind theme/content config
├── tsconfig.json            # TypeScript compiler config
├── vite.config.ts           # Vite/react/dev proxy config
└── dev-proxy.config.json    # Local dev proxy config; ignored, may contain endpoint info
```

## Directory Purposes

**Repository root:**
- Purpose: Houses frontend app, backend app, static assets, build/config files, and planning docs.
- Contains: `package.json`, `vite.config.ts`, `tsconfig.json`, `tailwind.config.js`, `postcss.config.js`, `index.html`, `backend-go/`, `src/`, `public/`, `docs/`, `.planning/`.
- Key files: `package.json`, `vite.config.ts`, `index.html`, `backend-go/main.go`, `src/main.tsx`.

**`src/`:**
- Purpose: React/Vite frontend source for both user-facing app and admin app.
- Contains: App shells, Zustand store, domain types, components, hooks, API clients, browser storage helpers, tests, and global CSS.
- Key files: `src/main.tsx`, `src/App.tsx`, `src/store.ts`, `src/types.ts`, `src/index.css`.

**`src/components/`:**
- Purpose: User-facing feature components and modals for the image generation workspace.
- Contains: Header/search/task/input components, image detail/lightbox/mask editor, settings/login/feedback/announcement/changelog modals, confirm dialog, toast bridge, and legacy/custom select.
- Key files: `src/components/Header.tsx`, `src/components/SearchBar.tsx`, `src/components/TaskGrid.tsx`, `src/components/TaskCard.tsx`, `src/components/InputBar.tsx`, `src/components/DetailModal.tsx`, `src/components/MaskEditorModal.tsx`, `src/components/SettingsModal.tsx`, `src/components/LoginModal.tsx`, `src/components/FeedbackModal.tsx`, `src/components/ChangelogModal.tsx`, `src/components/AnnouncementModal.tsx`, `src/components/ConfirmDialog.tsx`, `src/components/Toast.tsx`, `src/components/Select.tsx`.

**`src/components/ui/`:**
- Purpose: Shared UI primitives and small reusable visual components.
- Contains: Radix wrappers, Tailwind variant components, cards, table, tabs, switch, dialog, alert dialog, dropdown, popover, tooltip, scroll area, sonner bridge, empty state, and status badge.
- Key files: `src/components/ui/button.tsx`, `src/components/ui/dialog.tsx`, `src/components/ui/alert-dialog.tsx`, `src/components/ui/card.tsx`, `src/components/ui/input.tsx`, `src/components/ui/textarea.tsx`, `src/components/ui/table.tsx`, `src/components/ui/status-badge.tsx`, `src/components/ui/empty-state.tsx`, `src/components/ui/sonner.tsx`.

**`src/admin/`:**
- Purpose: Admin-only frontend route and transport layer.
- Contains: Admin shell, admin login form, admin dashboard, and admin API client/types.
- Key files: `src/admin/AdminPage.tsx`, `src/admin/AdminLogin.tsx`, `src/admin/AdminDashboard.tsx`, `src/admin/adminApi.ts`.

**`src/lib/`:**
- Purpose: Browser-side utilities, transport wrappers, image/canvas/mask helpers, IndexedDB access, and dev proxy normalization.
- Contains: `backendApi`, `db`, image canvas helpers, mask helpers, viewport/gesture helpers, size formatting/normalization, clipboard helpers, class merge helper, and tests.
- Key files: `src/lib/backendApi.ts`, `src/lib/db.ts`, `src/lib/canvasImage.ts`, `src/lib/mask.ts`, `src/lib/maskPreprocess.ts`, `src/lib/viewportTransform.ts`, `src/lib/viewport.ts`, `src/lib/size.ts`, `src/lib/clipboard.ts`, `src/lib/devProxy.ts`, `src/lib/utils.ts`.

**`src/hooks/`:**
- Purpose: Shared React hooks.
- Contains: Escape-key modal stack management.
- Key files: `src/hooks/useCloseOnEscape.ts`.

**`backend-go/`:**
- Purpose: Backend server module for auth, image storage, task persistence, generation orchestration, admin management, announcements, feedback, and changelog APIs.
- Contains: Go module files, entry point, packages for config/database/handlers/middleware/services/utils, plus ignored runtime config/data/upload directories.
- Key files: `backend-go/main.go`, `backend-go/go.mod`, `backend-go/config/config.go`, `backend-go/database/database.go`, `backend-go/database/models.go`.

**`backend-go/config/`:**
- Purpose: Load backend runtime config and manage mutable endpoint pool state.
- Contains: Config structs, default values, endpoint sorting/cloning, mutex-protected endpoint getter/setter, and config persistence for endpoint edits.
- Key files: `backend-go/config/config.go`, `backend-go/config/config_test.go`.

**`backend-go/database/`:**
- Purpose: Initialize SQLite/GORM and define persistent table schemas.
- Contains: DB singleton, migrations, admin/default announcement seeders, and GORM model structs.
- Key files: `backend-go/database/database.go`, `backend-go/database/models.go`.

**`backend-go/handler/`:**
- Purpose: HTTP boundary for all Gin routes.
- Contains: Handlers for auth, config, images, tasks/SSE, generation/edit, admin, announcement, feedback, and changelog endpoints, plus handler tests.
- Key files: `backend-go/handler/auth.go`, `backend-go/handler/config.go`, `backend-go/handler/images.go`, `backend-go/handler/tasks.go`, `backend-go/handler/generate.go`, `backend-go/handler/admin.go`, `backend-go/handler/announcement.go`, `backend-go/handler/feedback.go`, `backend-go/handler/changelog.go`, `backend-go/handler/images_test.go`.

**`backend-go/service/`:**
- Purpose: Backend domain logic independent of HTTP.
- Contains: Service DTOs, auth/quota/redemption code logic, image file persistence, task DB conversion/CRUD, OpenAI Images API calls, endpoint queue/limiters, announcement, feedback, and changelog services.
- Key files: `backend-go/service/models.go`, `backend-go/service/auth.go`, `backend-go/service/image.go`, `backend-go/service/task.go`, `backend-go/service/openai.go`, `backend-go/service/queue.go`, `backend-go/service/announcement.go`, `backend-go/service/feedback.go`, `backend-go/service/changelog.go`, `backend-go/service/image_test.go`.

**`backend-go/middleware/`:**
- Purpose: Cross-cutting Gin middleware.
- Contains: User JWT middleware, admin JWT middleware, authenticated user extraction, and request logging.
- Key files: `backend-go/middleware/middleware.go`, `backend-go/middleware/logger.go`.

**`backend-go/util/`:**
- Purpose: Stateless backend helpers.
- Contains: Random ID generation, SHA-256 hashing, runtime directory creation, user upload directory construction, relative upload path conversion, and upload path resolution guard.
- Key files: `backend-go/util/id.go`, `backend-go/util/crypto.go`, `backend-go/util/paths.go`.

**`public/`:**
- Purpose: Static assets served/copied by Vite.
- Contains: Web manifest, PWA icon, and service worker.
- Key files: `public/manifest.webmanifest`, `public/pwa-icon.svg`, `public/sw.js`.

**`docs/`:**
- Purpose: Documentation/media assets.
- Contains: Example images under `docs/images/`.
- Key files: `docs/images/example_pc_1.png`, `docs/images/example_mb_1.jpg`.

**`.planning/`:**
- Purpose: GSD planning and codebase mapping output.
- Contains: Codebase docs in `.planning/codebase/` and phase planning directories in `.planning/phases/`.
- Key files: `.planning/codebase/ARCHITECTURE.md`, `.planning/codebase/STRUCTURE.md`.

## Key File Locations

**Entry Points:**
- `index.html`: Browser HTML host and Vite script tag.
- `src/main.tsx`: Frontend entry; installs viewport guards, handles service worker, and branches between user/admin apps.
- `src/App.tsx`: User app composition and bootstrap effects.
- `src/admin/AdminPage.tsx`: Admin route shell.
- `backend-go/main.go`: Backend entry point, middleware setup, route registration, and server listen.
- `public/sw.js`: PWA service worker source.

**Configuration:**
- `package.json`: Frontend scripts and npm dependencies.
- `tsconfig.json`: TypeScript compiler settings for `src/`.
- `vite.config.ts`: React plugin, base path, dev proxy injection, and Vite dev server proxy.
- `tailwind.config.js`: Tailwind content globs, dark mode, Zinc-as-gray palette, font families.
- `postcss.config.js`: PostCSS plugin wiring.
- `dev-proxy.config.json`: Local dev proxy config; ignored by Git and should be treated as environment-specific.
- `backend-go/go.mod`: Go module dependencies.
- `backend-go/config/config.go`: Backend default config, runtime config loading, endpoint pool state/persistence.
- `backend-go/config.json`: Backend runtime config/secrets; ignored by Git, note existence only and do not read or quote contents.
- `.gitignore`: Generated/runtime/secret file exclusions.

**Core Logic:**
- `src/store.ts`: Client global state, task submission, SSE/polling, image cache, and task/image actions.
- `src/types.ts`: Frontend domain types and defaults for settings/tasks/images/content models.
- `src/lib/backendApi.ts`: User-facing backend API client and user token storage.
- `src/admin/adminApi.ts`: Admin backend API client and admin token storage.
- `src/lib/db.ts`: IndexedDB image store and data URL hashing.
- `src/lib/canvasImage.ts`: Image/canvas conversion and mask validation/preview.
- `src/lib/mask.ts`: Mask coverage and target ordering helpers.
- `src/lib/maskPreprocess.ts`: Mask target resize/PNG preparation helpers.
- `src/lib/viewportTransform.ts`: Mask editor view transform math.
- `backend-go/handler/generate.go`: Async generation request lifecycle.
- `backend-go/service/openai.go`: OpenAI Images API client, failover, concurrent multi-image calls, data URL decoding.
- `backend-go/service/queue.go`: Endpoint concurrency slot acquisition and limiter reset.
- `backend-go/service/task.go`: Task model/DTO conversion and task persistence.
- `backend-go/service/image.go`: Image deduplication, file writes, data URL/file reads, and image deletes.
- `backend-go/service/auth.go`: JWT, login, redemption codes, users, quota, and admin user mutations.
- `backend-go/database/models.go`: Database schema for users/codes/images/tasks/announcements/feedback/changelog.

**User Interface:**
- `src/components/Header.tsx`: Top navigation and modal triggers.
- `src/components/SearchBar.tsx`: Task search/status/favorite filters.
- `src/components/TaskGrid.tsx`: Filtered task grid and drag/selection behavior.
- `src/components/TaskCard.tsx`: Per-task summary card, thumbnails, timing, swipe selection.
- `src/components/InputBar.tsx`: Prompt/params/input image/mask controls and submit trigger.
- `src/components/DetailModal.tsx`: Task detail, output navigation, prompt/metadata display, copy/reuse/edit/delete actions.
- `src/components/Lightbox.tsx`: Fullscreen image viewing, zoom/pan, mask preview, image navigation.
- `src/components/MaskEditorModal.tsx`: Canvas-based mask drawing editor.
- `src/components/SettingsModal.tsx`: Theme, quota, redeem, logout controls.
- `src/components/LoginModal.tsx`: Redemption-code login/register modal.
- `src/components/FeedbackModal.tsx`: User feedback form.
- `src/components/AnnouncementModal.tsx`: Public announcement display.
- `src/components/ChangelogModal.tsx`: Published changelog browser.
- `src/components/ConfirmDialog.tsx`: Store-driven confirmation modal.
- `src/components/Toast.tsx`: Store-to-Sonner toast bridge.

**Admin Interface:**
- `src/admin/AdminPage.tsx`: Admin route shell.
- `src/admin/AdminLogin.tsx`: Admin key login form.
- `src/admin/AdminDashboard.tsx`: Admin tabs and CRUD interactions.
- `src/admin/adminApi.ts`: Admin transport wrappers and admin token management.

**Backend API Surface:**
- `backend-go/main.go`: Canonical route list.
- `backend-go/handler/auth.go`: `/api/auth/login`, `/api/auth/me`, `/api/auth/redeem`.
- `backend-go/handler/config.go`: `/api/config/public`.
- `backend-go/handler/images.go`: `/api/images` upload/get/delete.
- `backend-go/handler/tasks.go`: `/api/tasks` list/update/delete/clear and `/api/tasks/:id/stream` SSE.
- `backend-go/handler/generate.go`: `/api/generate` and `/api/edit`.
- `backend-go/handler/admin.go`: `/api/admin/login`, users, codes, and endpoint config.
- `backend-go/handler/announcement.go`: `/api/announcement` and admin announcement routes.
- `backend-go/handler/feedback.go`: `/api/feedback` and admin feedback routes.
- `backend-go/handler/changelog.go`: `/api/changelog`, `/api/changelog/latest`, and admin changelog routes.

**Testing:**
- `src/store.test.ts`: Store/task/image/session behavior tests.
- `src/lib/mask.test.ts`: Mask helper tests.
- `src/lib/maskPreprocess.test.ts`: Mask preprocessing tests.
- `src/lib/viewportTransform.test.ts`: View transform math tests.
- `src/lib/db.test.ts`: IndexedDB/hash behavior tests.
- `backend-go/config/config_test.go`: Endpoint pool sorting/copy/persistence behavior tests.
- `backend-go/handler/images_test.go`: Image handler behavior tests.
- `backend-go/service/image_test.go`: Image service behavior tests.

## Naming Conventions

**Files:**
- React component files use PascalCase: `src/components/TaskGrid.tsx`, `src/components/MaskEditorModal.tsx`, `src/admin/AdminDashboard.tsx`.
- Radix/Tailwind primitive files use lowercase kebab-case: `src/components/ui/alert-dialog.tsx`, `src/components/ui/dropdown-menu.tsx`, `src/components/ui/status-badge.tsx`.
- Frontend non-component utilities use camelCase or concise lowercase names: `src/lib/backendApi.ts`, `src/lib/canvasImage.ts`, `src/lib/maskPreprocess.ts`, `src/lib/viewportTransform.ts`, `src/lib/db.ts`, `src/lib/utils.ts`.
- Frontend tests are colocated with source using `*.test.ts`: `src/store.test.ts`, `src/lib/db.test.ts`.
- Go source files use lowercase snake-free feature names: `backend-go/handler/generate.go`, `backend-go/service/openai.go`, `backend-go/database/models.go`.
- Go tests are colocated with packages using `*_test.go`: `backend-go/config/config_test.go`, `backend-go/handler/images_test.go`.

**Directories:**
- Top-level source directories are domain/layer names: `src/components/`, `src/admin/`, `src/lib/`, `src/hooks/`, `backend-go/handler/`, `backend-go/service/`, `backend-go/database/`.
- User feature UI lives in `src/components/`; admin-only UI lives in `src/admin/`; shared primitives live in `src/components/ui/`.
- Backend packages are singular layer names (`handler`, `service`, `database`, `middleware`, `config`, `util`) matching Go package names.
- Runtime generated folders stay under `backend-go/data/`, `backend-go/upload/`, `dist/`, and `node_modules/` and are ignored.

## Where to Add New Code

**New user-facing feature:**
- Primary UI: add feature component under `src/components/` using PascalCase, e.g. `src/components/NewFeatureModal.tsx`.
- Shared state/action: add fields and actions to `src/store.ts` when multiple components need the state or backend task/image/session data changes.
- Shared types/defaults: add to `src/types.ts` when the shape crosses components or API boundaries.
- API calls: add user-facing transport wrappers to `src/lib/backendApi.ts`.
- Backend route: register route in `backend-go/main.go`, add HTTP binding in `backend-go/handler/<feature>.go`, add domain behavior in `backend-go/service/<feature>.go`, and add schema to `backend-go/database/models.go` if persistence is required.
- Tests: add frontend tests next to the relevant `src/*.test.ts` or `src/lib/*.test.ts`; add Go tests in the matching `backend-go/<package>/*_test.go`.

**New admin feature:**
- Primary UI: extend tabs/state/rendering in `src/admin/AdminDashboard.tsx` or add a focused admin component under `src/admin/` if the tab becomes large.
- API calls: add wrappers and TypeScript DTOs to `src/admin/adminApi.ts`.
- Backend route: register under the admin group in `backend-go/main.go`, protect with `middleware.AdminMiddleware()` through the `adminAuth` route group, implement in `backend-go/handler/<feature>.go`, and put domain logic in `backend-go/service/<feature>.go`.
- Tests: add service/handler tests under `backend-go/service/` or `backend-go/handler/`; add frontend tests when logic is extractable from `src/admin/AdminDashboard.tsx`.

**New shared UI primitive:**
- Implementation: add to `src/components/ui/` using lowercase kebab-case when wrapping Radix or representing a reusable primitive, e.g. `src/components/ui/new-widget.tsx`.
- Styling: use `cn()` from `src/lib/utils.ts`; reuse Tailwind/Zinc/dark-mode conventions from `src/components/ui/button.tsx` and `src/components/ui/dialog.tsx`.
- Consumers: import directly from the primitive file, e.g. `import { Button } from './ui/button'` from `src/components/*` or `import { Button } from '../components/ui/button'` from `src/admin/*`.

**New user app component/module:**
- Implementation: `src/components/<PascalName>.tsx` for UI; `src/lib/<camelName>.ts` for pure/browser utility logic.
- State: keep local form-only state inside the component; move shared or persisted state to `src/store.ts`.
- Modals: use primitives from `src/components/ui/dialog.tsx` and add `data-no-drag-select` when the modal should not trigger task-grid drag selection.

**New backend endpoint:**
- Route registration: `backend-go/main.go` in the appropriate group (`auth`, `cfg`, `images`, `tasks`, `generate`, or `adminAuth`).
- Handler: `backend-go/handler/<feature>.go` with `c.ShouldBindJSON`, `c.Param`, `c.Query`, and `c.JSON`/`c.File`/SSE response handling.
- Service: `backend-go/service/<feature>.go` with reusable business logic and database operations.
- Models: `backend-go/database/models.go` plus `backend-go/database/database.go` `AutoMigrate` list for new tables.
- Auth: use `middleware.AuthMiddleware()` for user routes and `middleware.AdminMiddleware()` for admin routes from `backend-go/middleware/middleware.go`.

**New external API or generation behavior:**
- Endpoint/failover logic: extend `backend-go/service/openai.go`.
- Concurrency behavior: extend `backend-go/service/queue.go`.
- Admin config surface: extend `backend-go/config/config.go`, `backend-go/handler/admin.go`, `src/admin/adminApi.ts`, and `src/admin/AdminDashboard.tsx`.
- Public config: expose safe values only through `backend-go/handler/config.go` and consume through `src/lib/backendApi.ts`/`src/store.ts`.

**New database-backed entity:**
- Database model: add struct and `TableName()` in `backend-go/database/models.go`.
- Migration: add the struct to `DB.AutoMigrate(...)` in `backend-go/database/database.go`.
- Service DTO/conversion: add to `backend-go/service/<feature>.go` or `backend-go/service/models.go` if shared across services.
- Frontend type: add to `src/types.ts` when the entity crosses the browser API boundary.

**Utilities:**
- Frontend shared helpers: `src/lib/`.
- Frontend React hooks: `src/hooks/`.
- Backend stateless helpers: `backend-go/util/`.
- Backend cross-cutting request concerns: `backend-go/middleware/`.
- Class name composition: use `src/lib/utils.ts` and avoid duplicating `clsx`/`tailwind-merge` logic.

**Static assets and PWA files:**
- Public static assets: `public/`.
- Documentation images: `docs/images/`.
- Service worker behavior: `public/sw.js` and registration in `src/main.tsx`.

**Configuration changes:**
- Frontend dependency/scripts: `package.json` and `package-lock.json`.
- Build/dev behavior: `vite.config.ts` and `tsconfig.json`.
- Styling theme/content: `tailwind.config.js` and `src/index.css`.
- Backend defaults/config shape: `backend-go/config/config.go`.
- Runtime secrets/endpoint values: `backend-go/config.json` only, never committed or documented with values.

## Special Directories

**`backend-go/data/`:**
- Purpose: Runtime SQLite database storage, including `backend-go/data/app.sqlite`.
- Generated: Yes
- Committed: No

**`backend-go/upload/`:**
- Purpose: Runtime uploaded/generated image files organized by user ID.
- Generated: Yes
- Committed: No

**`dist/`:**
- Purpose: Vite production build output.
- Generated: Yes
- Committed: No

**`node_modules/`:**
- Purpose: npm dependency install output.
- Generated: Yes
- Committed: No

**`.planning/codebase/`:**
- Purpose: Generated codebase maps consumed by GSD planning/execution commands.
- Generated: Yes
- Committed: Project-dependent; treat as planning artifact source for future agents.

**`.planning/phases/`:**
- Purpose: Existing GSD phase plans.
- Generated: Yes
- Committed: Project-dependent.

**`docs/images/`:**
- Purpose: Static example/demo images for documentation.
- Generated: No
- Committed: Yes

**`public/`:**
- Purpose: Static PWA assets copied to production build root.
- Generated: No
- Committed: Yes

**`.claude/`:**
- Purpose: Claude local workspace data.
- Generated: Yes
- Committed: No

**`backend-go/.idea/`:**
- Purpose: IDE project metadata.
- Generated: Yes
- Committed: No

---

*Structure analysis: 2026-05-22*
