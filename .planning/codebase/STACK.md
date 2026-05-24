# Technology Stack

**Analysis Date:** 2026-05-24

## Languages

**Primary:**
- TypeScript 5.8.3 - Frontend application (`src/`)
- Go 1.25.0 - Backend API server (`backend-go/`)

**Secondary:**
- JavaScript (ES2020+) - Build config, service worker, postcss config
- HTML - Single entry point `index.html`
- JSON - Runtime config (`backend-go/config.json`, `dev-proxy.config.json`)

## Runtime

**Frontend:**
- Vite 6.3.2 - Dev server and production bundler (`vite.config.ts`)
- Node.js (implied by Vite 6.x requirement: >=18.0 - exact version not pinned)
- ESM modules (`"type": "module"` in `package.json`)

**Backend:**
- Go 1.25.0 - Compiled binary, no external runtime required
- SQLite - Embedded database (file-based, no server process)
- gin-gonic/gin 1.10.1 - HTTP server

**Package Manager:**
- npm (version not pinned) - Frontend dependencies
- Lockfile: `package-lock.json` (committed)
- Go modules - Backend dependencies (`backend-go/go.mod`)
- Lockfile: `go.sum` implied by Go modules

## Frameworks

**Core:**
- React 19.1.0 - UI library (`src/main.tsx`, `src/App.tsx`)
- React DOM 19.1.0 - React renderer
- Gin 1.10.1 - Go HTTP web framework (`backend-go/main.go`)

**Testing:**
- Vitest 4.1.5 - Frontend test runner (ESM-compatible, Vite-native)
- Go standard `testing` package - Backend tests (files named `*_test.go`)

**Build/Dev:**
- Vite 6.3.2 - Frontend dev server, HMR, production bundling
- `@vitejs/plugin-react` 4.4.1 - React Fast Refresh and JSX transform
- TypeScript compiler (`tsc -b`) - Type checking as part of `npm run build`
- PostCSS 8.5.3 - CSS processing (Tailwind)
- Autoprefixer 10.4.21 - CSS vendor prefix automation
- Go standard `go build` / `go run` - Backend compilation

## Key Dependencies

**Critical (Frontend):**
- Zustand 5.0.5 - Global state management with localStorage persist middleware (`src/store.ts`)
- Sonner 2.0.1 - Toast notification system (`src/components/ui/sonner.tsx`)
- Lucide React 0.468.0 - Icon library (used across all UI components)
- class-variance-authority 0.7.1 - Component variant definitions (`src/components/ui/button.tsx`, etc.)
- clsx 2.1.1 - Conditional className merging
- tailwind-merge 2.6.0 - Utility class deduplication for Tailwind

**UI Primitives (Radix UI via shadcn/ui):**
- `@radix-ui/react-dialog` 1.1.6 - Modal dialogs
- `@radix-ui/react-alert-dialog` 1.1.6 - Confirmation dialogs
- `@radix-ui/react-dropdown-menu` 2.1.6 - Dropdown menus
- `@radix-ui/react-popover` 1.1.6 - Popover panels
- `@radix-ui/react-select` 2.1.6 - Select dropdowns
- `@radix-ui/react-tabs` 1.1.3 - Tab panels
- `@radix-ui/react-tooltip` 1.1.8 - Tooltips
- `@radix-ui/react-switch` 1.1.3 - Toggle switches
- `@radix-ui/react-scroll-area` 1.2.3 - Custom scroll areas
- `@radix-ui/react-separator` 1.1.2 - Visual dividers
- `@radix-ui/react-label` 2.1.2 - Form labels
- `@radix-ui/react-slot` 1.1.2 - Slot-based composition (for `asChild` pattern)

**Infrastructure (Backend):**
- GORM 1.30.0 - Go ORM with AutoMigrate, query builder (`backend-go/database/database.go`)
- gorm.io/driver/sqlite 1.6.0 - SQLite GORM driver
- golang-jwt/jwt/v5 5.2.2 - JWT token signing/verification (`backend-go/service/auth.go`)
- golang.org/x/crypto 0.52.0 - bcrypt password hashing (`backend-go/service/auth.go`)
- gin-contrib/cors 1.7.6 - CORS middleware

## Configuration

**Environment:**
- Frontend: `VITE_BACKEND_URL` environment variable (used in `src/lib/backendApi.ts` and `src/store.ts`) — sets the backend API base URL, defaults to `http://localhost:3001`
- Backend: `backend-go/config.json` — runtime configuration file (port, JWT secret, API endpoints, pricing, invite settings); default port is 3001
- `.env` files: Not detected in project root

**Build:**
- `vite.config.ts` — Vite configuration with React plugin, dev proxy support, and build-time config injection (`__DEV_PROXY_CONFIG__`)
- `tsconfig.json` — TypeScript config targeting ES2020, strict mode, bundler module resolution
- `tailwind.config.js` — Tailwind CSS 3.x with shadcn/ui CSS variable color system, class-based dark mode, Zinc color palette, custom font families
- `postcss.config.js` — PostCSS with Tailwind and Autoprefixer plugins
- `components.json` — shadcn/ui configuration (base color: Zinc, CSS variables enabled, TSX output)

**Backend Runtime Config (`backend-go/config.json`):**
- `port` (default: 3001) — HTTP server port
- `jwtSecret` — Secret key for JWT token signing
- `adminApikey` — Admin authentication key
- `model` (default: `gpt-image-2`) — OpenAI model name
- `apiMode` (default: `images`) — API mode selection
- `timeout` (default: 6000) — Request timeout in seconds
- `codexCli` (default: true) — Codex CLI compatibility mode
- `apiEndpoints[]` — Pool of API endpoints with baseUrl, apiKey, maxConcurrency, priority, costPerImageX10000
- `salePriceX10000` — Per-image sale price (internal accounting)
- `inviteEnabled`, `inviteInviterReward`, `inviteInviteeReward`, `inviteDefaultQuota` — Invite system config

## Platform Requirements

**Development:**
- Node.js (any version supporting Vite 6.x)
- Go 1.25.0+
- No external database server (SQLite is file-based)
- Internet access for OpenAI API calls

**Production:**
- Frontend: Static file server (Vite outputs to `dist/` with relative base path `./`)
- Backend: Go binary with file write access for SQLite database and image uploads
- Deployment target: Self-hosted (backend IP `http://43.133.38.194:3004` configured in dev proxy)
- No cloud platform-specific dependencies

---

*Stack analysis: 2026-05-24*
