# Codebase Structure

**Analysis Date:** 2026-05-24

## Directory Layout

```
gpt_image_playground/
├── src/                              # Frontend React SPA source
│   ├── main.tsx                      # Entry point; route-splits to App or AdminPage
│   ├── App.tsx                       # Main user-facing app root
│   ├── App-test.test.tsx             # App smoke test
│   ├── store.ts                      # Zustand global state store
│   ├── store.test.ts                 # Store tests
│   ├── types.ts                      # TypeScript type definitions
│   ├── index.css                     # Global styles + CSS custom properties
│   ├── vite-env.d.ts                 # Vite environment type declarations
│   ├── admin/                        # Admin panel (lazy-loaded route)
│   │   ├── AdminPage.tsx             # Admin route root; login/dashboard toggle
│   │   ├── AdminLogin.tsx            # Admin login form
│   │   ├── AdminDashboard.tsx        # Full admin dashboard (~2300+ lines)
│   │   ├── AdminDashboard.test.tsx   # Admin dashboard tests
│   │   ├── AdminDashboard.tsx.bak    # Backup file (stale)
│   │   ├── AdminDashboard.tsx.bak2   # Backup file (stale)
│   │   ├── adminApi.ts              # Admin API client + type definitions
│   │   ├── adminApi.test.ts         # Admin API tests
│   │   ├── adminApi-invite.test.ts  # Admin API invite tests
│   │   ├── moneyFormat.ts           # Money formatting utilities (X10000 <-> display)
│   │   └── moneyFormat.test.ts      # Money format tests
│   ├── components/                  # Application components
│   │   ├── Header.tsx               # Top navigation bar
│   │   ├── SearchBar.tsx            # Search and filter controls
│   │   ├── TaskGrid.tsx             # Task list with drag-select
│   │   ├── TaskCard.tsx             # Individual task display card
│   │   ├── InputBar.tsx             # Floating prompt input + parameter bar
│   │   ├── DetailModal.tsx          # Task detail modal
│   │   ├── Lightbox.tsx             # Image lightbox viewer
│   │   ├── SettingsModal.tsx        # Settings panel (model, API, theme)
│   │   ├── ConfirmDialog.tsx        # Confirmation dialog component
│   │   ├── LoginModal.tsx           # Login modal (code + password)
│   │   ├── LoginModal.test.tsx      # Login modal tests
│   │   ├── RegisterModal.tsx        # Registration modal
│   │   ├── RegisterModal.test.tsx   # Registration modal tests
│   │   ├── MigrationModal.tsx       # Account migration modal (code-only -> password)
│   │   ├── MigrationModal.test.tsx  # Migration modal tests
│   │   ├── AnnouncementModal.tsx    # Announcement display modal
│   │   ├── ChangelogModal.tsx       # Changelog viewer modal
│   │   ├── AppearanceModal.tsx      # Appearance/theme picker
│   │   ├── FeedbackModal.tsx        # Bug/feature feedback submission
│   │   ├── HelpModal.tsx            # Help/about modal
│   │   ├── MaskEditorModal.tsx      # Image mask editor
│   │   ├── SizePickerModal.tsx      # Image size picker
│   │   ├── Select.tsx               # Custom select dropdown component
│   │   └── ui/                      # shadcn/ui reusable primitives (Radix-based)
│   │       ├── alert-dialog.tsx     # Alert dialog (Radix Alert Dialog)
│   │       ├── app-dialog.tsx       # Application dialog wrapper
│   │       ├── badge.tsx            # Badge component
│   │       ├── button.tsx           # Button component (Radix Slot)
│   │       ├── card.tsx             # Card component
│   │       ├── dialog.tsx           # Dialog (Radix Dialog)
│   │       ├── dropdown-menu.tsx    # Dropdown menu (Radix Dropdown Menu)
│   │       ├── empty-state.tsx      # Empty state placeholder
│   │       ├── input.tsx            # Input component
│   │       ├── label.tsx            # Label component (Radix Label)
│   │       ├── popover.tsx          # Popover (Radix Popover)
│   │       ├── scroll-area.tsx      # Scroll area (Radix Scroll Area)
│   │       ├── select.tsx           # Select (Radix Select)
│   │       ├── separator.tsx        # Separator (Radix Separator)
│   │       ├── sonner.tsx           # Sonner toast configuration
│   │       ├── status-badge.tsx     # Status indicator badge
│   │       ├── switch.tsx           # Toggle switch (Radix Switch)
│   │       ├── table.tsx            # Table component
│   │       ├── tabs.tsx             # Tabs (Radix Tabs)
│   │       ├── textarea.tsx         # Textarea component
│   │       └── tooltip.tsx          # Tooltip (Radix Tooltip)
│   ├── hooks/                       # Custom React hooks
│   │   └── useCloseOnEscape.ts      # ESC key handler hook
│   └── lib/                         # Utility libraries
│       ├── backendApi.ts            # User-facing API client (fetch wrappers, SSE)
│       ├── backendApi.test.ts       # API client tests
│       ├── canvasImage.ts           # Canvas-based image manipulation
│       ├── clipboard.ts             # Clipboard operations (copy text/image)
│       ├── db.ts                    # IndexedDB wrapper (image storage)
│       ├── db.test.ts               # IndexedDB tests
│       ├── devProxy.ts              # Dev proxy config normalization
│       ├── mask.ts                  # Mask image ordering logic
│       ├── mask.test.ts             # Mask tests
│       ├── maskPreprocess.ts        # Mask preprocessing utilities
│       ├── maskPreprocess.test.ts   # Mask preprocessing tests
│       ├── paramDisplay.tsx         # Parameter display component
│       ├── size.ts                  # Image size normalization
│       ├── utils.ts                 # General utilities (cn() classname merger)
│       ├── viewport.ts              # Mobile viewport guard
│       ├── viewportTransform.ts     # Viewport coordinate transforms
│       └── viewportTransform.test.ts # Viewport transform tests
├── backend-go/                      # Go backend API server
│   ├── main.go                      # Server entry; router setup, middleware, startup
│   ├── go.mod                       # Go module definition
│   ├── go.sum                       # Go module checksums
│   ├── config.json                  # Runtime config (mutable at runtime via admin API)
│   ├── config/
│   │   ├── config.go                # Config loading, endpoint pool, pricing, persistence
│   │   └── config_test.go           # Config tests
│   ├── database/
│   │   ├── database.go              # GORM init, AutoMigrate, admin/announcement seed
│   │   ├── models.go                # GORM models: User, RedemptionCode, Image, Task, Announcement, Feedback, ChangelogEntry, BillingRecord
│   │   └── models_test.go           # Model tests
│   ├── handler/
│   │   ├── generate.go              # POST /api/generate, /api/edit; async execution, billing
│   │   ├── generate_billing_test.go # Generation billing tests
│   │   ├── tasks.go                 # Task CRUD + SSE streaming endpoint
│   │   ├── auth.go                  # Auth handlers (login, register, redeem, migrate, etc.)
│   │   ├── auth_handler_test.go     # Auth handler tests
│   │   ├── admin.go                 # Admin handlers (users, codes, endpoints, analytics, etc.)
│   │   ├── admin_handler_test.go    # Admin handler tests
│   │   ├── admin_analytics_test.go  # Admin analytics tests
│   │   ├── admin_pricing_test.go    # Admin pricing tests
│   │   ├── images.go                # Image upload/download/delete handlers
│   │   ├── images_test.go           # Image handler tests
│   │   ├── config.go                # Public config endpoint
│   │   ├── announcement.go          # Announcement endpoints (public + admin)
│   │   ├── changelog.go             # Changelog endpoints (public + admin)
│   │   └── feedback.go              # Feedback endpoints (create + admin list/update)
│   ├── middleware/
│   │   ├── middleware.go            # AuthMiddleware, AdminMiddleware, GetAuthUser helper
│   │   └── logger.go               # Request logging middleware
│   ├── service/
│   │   ├── models.go                # Service-level type definitions (User, AuthUser, TaskRecord, etc.)
│   │   ├── models_test.go           # Service model tests
│   │   ├── openai.go                # OpenAI API client: generations, edits, failover, concurrent calls
│   │   ├── openai_failover_test.go  # Failover tests
│   │   ├── queue.go                 # Concurrency slot manager (per-endpoint semaphores)
│   │   ├── auth.go                  # Auth logic: JWT, redemption codes, password auth, registration, invites
│   │   ├── auth_test.go             # Auth service tests
│   │   ├── billing.go               # Billing record creation
│   │   ├── billing_test.go          # Billing tests
│   │   ├── image.go                 # Image storage: save, read, data URL conversion, SHA-256 dedup
│   │   ├── image_test.go            # Image service tests
│   │   ├── task.go                  # Task CRUD, pending image counting
│   │   ├── analytics.go             # Billing analytics aggregation queries
│   │   ├── analytics_test.go        # Analytics tests
│   │   ├── announcement.go          # Announcement CRUD
│   │   ├── changelog.go             # Changelog CRUD with validation
│   │   ├── feedback.go              # Feedback CRUD
│   │   ├── money.go                 # Money scale utilities
│   │   └── money_test.go            # Money tests
│   ├── util/
│   │   ├── crypto.go                # SHA-256 hash utility
│   │   ├── id.go                    # Random ID generation (21 chars)
│   │   └── paths.go                 # Path resolution, directory creation, path traversal guard
│   ├── log/                         # Logger initialization
│   ├── data/                        # SQLite database file(s)
│   └── upload/                      # Uploaded/generated image files (per-user subdirs)
├── public/                          # Public static assets
│   ├── manifest.webmanifest         # PWA manifest
│   ├── pwa-icon.svg                 # PWA icon
│   └── sw.js                        # Service worker
├── docs/                            # Documentation
│   └── images/                      # Documentation images
├── dist/                            # Vite build output (generated, not committed)
├── node_modules/                    # NPM dependencies (generated, not committed)
├── .claude/                         # Claude agent configuration
│   └── worktrees/                   # Agent worktree state
├── .planning/                       # Planning documents
│   └── codebase/                    # Codebase mapping documents (this directory)
├── package.json                     # NPM package definition + scripts
├── package-lock.json                # NPM lockfile
├── vite.config.ts                   # Vite build configuration
├── tsconfig.json                    # TypeScript configuration
├── tailwind.config.js               # Tailwind CSS configuration (Zinc color scheme, shadcn tokens)
├── postcss.config.js                # PostCSS configuration
├── components.json                  # shadcn/ui component registry
├── dev-proxy.config.json            # Development proxy configuration
├── index.html                       # SPA entry HTML
├── .gitignore                       # Git ignore rules
└── README.md                        # Project documentation
```

## Directory Purposes

**src/ -- Frontend**
- Purpose: All frontend source code; React TypeScript SPA with Zustand state management
- Contains: Components, store, API clients, utility libraries, type definitions, styles
- Key files: `main.tsx` (entry point), `App.tsx` (root component), `store.ts` (state/business logic)

**src/components/ -- Application Components**
- Purpose: Feature-level UI components composing the user-facing application
- Contains: Header, SearchBar, TaskGrid, TaskCard, InputBar, modal components (Settings, Login, Detail, Lightbox, MaskEditor, etc.)
- Key files: `TaskGrid.tsx` (task list with multi-select), `InputBar.tsx` (prompt input + parameter controls), `TaskCard.tsx` (individual task display)

**src/components/ui/ -- UI Primitives**
- Purpose: Reusable UI components following shadcn/ui patterns, built on Radix UI and Tailwind CSS
- Contains: Button, Card, Dialog, Input, Select, Tabs, Switch, Badge, Dropdown, Popover, ScrollArea, Separator, Table, Textarea, Tooltip, etc.
- Key files: `button.tsx` (with CVA variants), `dialog.tsx` (Radix Dialog wrapper), `tabs.tsx` (Radix Tabs wrapper)

**src/admin/ -- Admin Panel**
- Purpose: Lazy-loaded admin dashboard with user management, code generation, API endpoint configuration, analytics, announcement/changelog management, invite system
- Contains: AdminPage (route root), AdminDashboard (main dashboard), AdminLogin, adminApi (API client with type definitions)
- Key files: `AdminDashboard.tsx` (large component, ~2300+ lines with 8 tab sections), `adminApi.ts` (full admin API client)

**src/lib/ -- Utilities**
- Purpose: Shared utility modules for API communication, image storage, image processing, clipboard operations
- Contains: `backendApi.ts` (fetch wrappers + SSE streaming), `db.ts` (IndexedDB), `canvasImage.ts`, `mask.ts`, `size.ts`, `viewport.ts`, `clipboard.ts`
- Key files: `backendApi.ts` (all user-facing API calls including SSE), `db.ts` (IndexedDB wrapper with SHA-256 hashing)

**src/hooks/ -- Custom Hooks**
- Purpose: Reusable React hooks
- Contains: `useCloseOnEscape.ts` (ESC key listener)

**backend-go/ -- Go Backend**
- Purpose: Monolithic REST API server; handles all business logic, database operations, external API calls
- Contains: main.go, config/, database/, handler/, service/, middleware/, util/, log/
- Key files: `main.go` (server bootstrap, route registration), `config.json` (runtime config)

**backend-go/handler/ -- HTTP Handlers**
- Purpose: Gin HTTP handler functions; request parsing, validation, response formatting
- Contains: One file per resource domain (auth, generate, tasks, images, admin, config, announcement, changelog, feedback)
- Key files: `generate.go` (image generation + async execution + billing), `tasks.go` (CRUD + SSE streaming), `admin.go` (all admin endpoints)

**backend-go/service/ -- Business Logic**
- Purpose: Business logic layer; database operations, external API calls, authentication logic
- Contains: One file per domain (auth, openai, queue, billing, image, task, analytics, announcement, changelog, feedback)
- Key files: `openai.go` (OpenAI API client with failover), `queue.go` (concurrency slot manager), `auth.go` (JWT, password auth, registration, invites)

**backend-go/config/ -- Configuration**
- Purpose: Runtime configuration management; JSON file loading, endpoint pool management, config persistence
- Contains: `config.go` (App struct, endpoint pool with RWMutex, persistence methods for endpoints/pricing/invite config)
- Key files: `config.go` (singleton config with hot-swappable endpoint pool)

**backend-go/database/ -- Database Layer**
- Purpose: GORM ORM initialization, model definitions, seed data
- Contains: `database.go` (SQLite init with WAL mode, AutoMigrate), `models.go` (8 GORM models)
- Key files: `models.go` (all database table definitions)

**backend-go/middleware/ -- HTTP Middleware**
- Purpose: Authentication, authorization, request logging
- Contains: `middleware.go` (JWT auth + admin role check), `logger.go` (request logging)
- Key files: `middleware.go` (AuthMiddleware with Bearer token + query param support)

**backend-go/util/ -- Utilities**
- Purpose: Shared utility functions (ID generation, crypto, path resolution)
- Contains: `crypto.go`, `id.go`, `paths.go`
- Key files: `id.go` (21-char random ID), `paths.go` (path traversal guard for file serving)

**public/ -- Static Assets**
- Purpose: Served as static files by the web server (or Vite dev server)
- Contains: PWA manifest, service worker, icon

## Key File Locations

**Entry Points:**
- `src/main.tsx`: Frontend entry; creates React root, conditionally renders App or AdminPage based on URL path
- `backend-go/main.go`: Backend entry; config load, DB init, route registration, middleware setup, server start on port 3001
- `index.html`: SPA HTML shell; minimal `<div id="root">` structure

**Configuration:**
- `vite.config.ts`: Vite build config; React plugin, dev proxy, `base: './'`
- `tsconfig.json`: TypeScript config; strict compilation
- `tailwind.config.js`: Tailwind config; dark mode via `class`, Zinc color palette, shadcn/ui CSS variable tokens, tailwindcss-animate plugin
- `postcss.config.js`: PostCSS config with Tailwind + Autoprefixer
- `components.json`: shadcn/ui component registry; style: "default", tailwind config path, CSS variables
- `backend-go/config.json`: Runtime config (port, JWT secret, model, endpoints, pricing, invite settings)
- `dev-proxy.config.json`: Dev proxy config for Vite dev server API forwarding
- `package.json`: NPM scripts: `dev`, `build` (tsc + vite build), `preview`, `test` (vitest), `test:watch`

**Core Logic:**
- `src/store.ts`: Zustand store; auth, settings, task lifecycle, image cache, polling, SSE, all user actions (submitTask, reuseConfig, editOutputs, removeTask, etc.)
- `src/lib/backendApi.ts`: User API client; login, register, getMe, image upload, task submission, SSE streaming, image URL construction
- `src/types.ts`: All TypeScript types (AppSettings, TaskParams, InputImage, TaskRecord, StoredImage, Announcement, ChangelogEntry, BugFeedback)
- `backend-go/service/openai.go`: OpenAI API integration; generations/edits calls, failover orchestration, concurrent multi-image, data URL conversion
- `backend-go/service/auth.go`: Full auth logic (JWT, bcrypt, redemption codes, registration, password auth, migration, invite system)
- `backend-go/config/config.go`: Runtime config management; endpoint pool with priority sorting, pricing, invite settings, persistence

**Testing:**
- Frontend: Co-located test files (`src/App-test.test.tsx`, `src/store.test.ts`, `src/admin/AdminDashboard.test.tsx`, `src/lib/backendApi.test.ts`, `src/lib/db.test.ts`, etc.) -- uses Vitest
- Backend: Co-located `_test.go` files in each package (handler/, service/, database/, config/) -- uses Go's built-in testing

## Naming Conventions

**Files:**
- Frontend: PascalCase for components (`TaskCard.tsx`), camelCase for utilities (`backendApi.ts`)
- Frontend tests: Co-located `*.test.ts` or `*.test.tsx`
- Backend: lowercase for all Go files, `*_test.go` for tests
- CSS: `index.css` for global styles

**Directories:**
- Frontend: lowercase with hyphenation (`components/ui/`)
- Backend: lowercase single-word (`handler/`, `service/`, `database/`)
- No nested directories in backend; all handlers/services/utilities are flat within their package

## Where to Add New Code

**New Feature (Frontend):**
- Primary code: `src/components/NewFeature.tsx` (or a new subdirectory `src/features/new-feature/` for complex features)
- State: Add to existing `src/store.ts` (single Zustand store) -- extend `AppState` interface and add setters/actions
- Types: Add to `src/types.ts`
- API calls: Add to `src/lib/backendApi.ts`

**New Feature (Backend):**
- Handler: `backend-go/handler/newfeature.go`
- Service: `backend-go/service/newfeature.go`
- Route: Register in `backend-go/main.go` in the appropriate route group
- Database: Add model to `backend-go/database/models.go` if new table needed; add AutoMigrate in `database.go`

**New Admin Feature:**
- Dashboard tab: Add to `src/admin/AdminDashboard.tsx` (add new `Tab` type and new section)
- API client: Add to `src/admin/adminApi.ts`
- Backend handler: Add to `backend-go/handler/admin.go` under the admin route group
- Backend service: Add to `backend-go/service/`

**New UI Primitive:**
- Implementation: `src/components/ui/new-component.tsx`
- Follow existing shadcn/ui pattern: import Radix primitives, use `cn()` from `lib/utils.ts`, export with ref forwarding

**New Utility:**
- Frontend: `src/lib/newUtil.ts`
- Backend: `backend-go/util/newutil.go`

**Tests:**
- Frontend: Co-locate with source (`src/components/NewFeature.test.tsx`)
- Backend: Co-locate with source (`backend-go/handler/newfeature_test.go`)

## Route Definitions

### Frontend Routes (client-side)
| Path | Component | Auth | Notes |
|------|-----------|------|-------|
| `/` | `App.tsx` | Optional | Main user-facing app |
| `/admin` | `admin/AdminPage.tsx` | Admin token in localStorage | Lazy-loaded admin dashboard |
| `/admin/*` | `admin/AdminPage.tsx` | Admin token in localStorage | All subpaths serve admin |

### Backend API Routes (register in `backend-go/main.go`)
| Method | Path | Auth | Handler | Purpose |
|--------|------|------|---------|---------|
| GET | `/api/health` | None | `handler.Health` | Health check |
| GET | `/api/announcement` | None | `handler.AnnouncementPublic` | Public announcement |
| GET | `/api/changelog/latest` | None | `handler.ChangelogLatestPublic` | Latest published changelog |
| GET | `/api/changelog` | None | `handler.ChangelogListPublic` | Published changelog list |
| POST | `/api/auth/login` | None | `handler.AuthLogin` | Code-based login |
| GET | `/api/auth/me` | User | `handler.AuthMe` | Current user info |
| POST | `/api/auth/redeem` | User | `handler.AuthRedeem` | Redeem code for quota |
| POST | `/api/auth/login-password` | None | `handler.AuthLoginPassword` | Password login |
| POST | `/api/auth/register` | None | `handler.AuthRegister` | User registration |
| POST | `/api/auth/migrate` | User | `handler.AuthMigrate` | Migrate code-only account to password |
| POST | `/api/auth/change-password` | User | `handler.AuthChangePassword` | Password change |
| PUT | `/api/auth/username` | User | `handler.AuthChangeUsername` | Username change |
| PUT | `/api/auth/invite-code` | User | `handler.AuthSetInviteCode` | Set/get invite code |
| GET | `/api/auth/invite-code` | User | `handler.AuthGetInviteCode` | Get own invite code |
| GET | `/api/auth/invited-users` | User | `handler.AuthGetInvitedUsers` | Users invited by current user |
| GET | `/api/config/public` | None | `handler.ConfigPublic` | Public app config |
| POST | `/api/images` | User | `handler.ImagesUpload` | Upload image |
| GET | `/api/images/:id` | User (query token) | `handler.ImagesGet` | Get image file |
| DELETE | `/api/images/:id` | User | `handler.ImagesDelete` | Delete image |
| GET | `/api/tasks` | User | `handler.TasksList` | List user tasks |
| GET | `/api/tasks/:id/stream` | User | `handler.TaskStream` | SSE task status stream |
| PUT | `/api/tasks/:id` | User | `handler.TasksUpdate` | Update task (favorite, etc.) |
| DELETE | `/api/tasks/:id` | User | `handler.TasksDelete` | Delete single task |
| DELETE | `/api/tasks` | User | `handler.TasksClear` | Clear all tasks |
| POST | `/api/generate` | User | `handler.GenerateImage` | Submit image generation |
| POST | `/api/edit` | User | `handler.GenerateImage` | Submit image editing |
| POST | `/api/feedback` | User | `handler.FeedbackCreate` | Submit bug/feature feedback |

### Admin API Routes
| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| POST | `/api/admin/login` | None | Admin login (apikey) |
| GET | `/api/admin/users` | Admin | List all users |
| PUT | `/api/admin/users/:id/quota` | Admin | Update user quota |
| PUT | `/api/admin/users/:id/status` | Admin | Toggle user status |
| DELETE | `/api/admin/users/:id` | Admin | Delete user |
| DELETE | `/api/admin/users` | Admin | Batch delete users |
| POST | `/api/admin/codes` | Admin | Create redemption code(s) |
| GET | `/api/admin/codes` | Admin | List redemption codes |
| DELETE | `/api/admin/codes` | Admin | Batch delete codes |
| GET | `/api/admin/config/endpoints` | Admin | Get API endpoints |
| PUT | `/api/admin/config/endpoints` | Admin | Update API endpoints |
| GET | `/api/admin/config/pricing` | Admin | Get pricing config |
| PUT | `/api/admin/config/pricing` | Admin | Update pricing config |
| GET | `/api/admin/announcement` | Admin | Get announcement |
| PUT | `/api/admin/announcement` | Admin | Update announcement |
| GET | `/api/admin/feedback` | Admin | List feedbacks |
| PUT | `/api/admin/feedback/:id/status` | Admin | Update feedback status |
| GET | `/api/admin/changelog` | Admin | List changelog entries |
| POST | `/api/admin/changelog` | Admin | Create changelog entry |
| PUT | `/api/admin/changelog/:id` | Admin | Update changelog entry |
| DELETE | `/api/admin/changelog/:id` | Admin | Delete changelog entry |
| GET | `/api/admin/analytics/summary` | Admin | Billing summary |
| GET | `/api/admin/analytics/trend` | Admin | Billing trend data |
| GET | `/api/admin/analytics/endpoints` | Admin | Billing endpoint breakdown |
| GET | `/api/admin/analytics/users` | Admin | Billing user breakdown |
| PUT | `/api/admin/users/:id/password` | Admin | Reset user password |
| GET | `/api/admin/invite-config` | Admin | Get invite config |
| PUT | `/api/admin/invite-config` | Admin | Update invite config |
| GET | `/api/admin/invites` | Admin | List invite codes + usage |

## Special Directories

**dist/:**
- Purpose: Vite production build output
- Generated: Yes (`npm run build`)
- Committed: No (in `.gitignore`)

**node_modules/:**
- Purpose: NPM dependencies
- Generated: Yes (`npm install`)
- Committed: No (in `.gitignore`)

**backend-go/data/:**
- Purpose: SQLite database file location
- Generated: Yes (on first startup)
- Committed: No (contains runtime data)

**backend-go/upload/:**
- Purpose: Uploaded and generated image files
- Generated: Yes (at runtime)
- Committed: No (user content)

**backend-go/log/:**
- Purpose: Log files
- Generated: Yes (at runtime)
- Committed: No

**backend-go/.idea/:**
- Purpose: JetBrains IDE settings
- Generated: Yes (by IDE)
- Committed: No

**.claude/worktrees/:**
- Purpose: Claude agent worktree state (sandbox for concurrent agent operations)
- Generated: Yes (by Claude Code tools)
- Committed: No (local agent state)

**docs/images/:**
- Purpose: Documentation screenshots and images
- Generated: No (manually added)
- Committed: Yes

---

*Structure analysis: 2026-05-24*
