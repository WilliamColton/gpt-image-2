# Technology Stack

**Analysis Date:** 2026-05-22

## Languages

**Primary:**
- TypeScript 5.8.3 - React application, admin UI, state management, browser data utilities, and API clients under `src/`; configured by `tsconfig.json` and dependency versions in `package.json`.
- Go 1.23.0 - Gin HTTP backend, OpenAI-compatible image service, SQLite persistence, auth, admin APIs, and file storage under `backend-go/`; configured by `backend-go/go.mod`.

**Secondary:**
- TSX / JSX via TypeScript - React component files under `src/components/`, `src/components/ui/`, `src/admin/`, and entry rendering in `src/main.tsx`.
- JavaScript ESM - Tooling config in `tailwind.config.js` and `postcss.config.js`, plus the PWA service worker in `public/sw.js`.
- CSS / Tailwind CSS - Global styles and utility classes in `src/index.css`, with Tailwind content scanning configured in `tailwind.config.js`.
- HTML - Vite shell in `index.html` and PWA manifest in `public/manifest.webmanifest`.

## Runtime

**Environment:**
- Browser runtime - The SPA uses DOM, IndexedDB, localStorage, service workers, Cache API, FileReader, and Web Crypto APIs in `src/main.tsx`, `src/lib/db.ts`, `src/store.ts`, `public/sw.js`, and `src/lib/canvasImage.ts`.
- Node.js - Used for Vite dev server, TypeScript build, and Vitest through scripts in `package.json`; Node version is not pinned in `.nvmrc`, `.node-version`, or `package.json` engines.
- Go HTTP server - `backend-go/main.go` starts a Gin server on `config.App.Port`, with default port `3001` set in `backend-go/config/config.go`.
- SQLite runtime - `backend-go/database/database.go` opens `data/app.sqlite` using GORM and `github.com/mattn/go-sqlite3` with WAL mode and foreign keys enabled.

**Package Manager:**
- npm - JavaScript package manager inferred from `package-lock.json` and scripts in `package.json`; npm version is not pinned.
- Lockfile: present (`package-lock.json`, lockfileVersion 3).
- Go modules - Backend dependencies are pinned by `backend-go/go.mod` and `backend-go/go.sum`.

## Frameworks

**Core:**
- React 19.1.0 - SPA rendering, admin route rendering, modal/component state, and UI composition in `src/main.tsx`, `src/App.tsx`, `src/components/`, and `src/admin/`.
- Vite 6.3.2 - Frontend dev server, build pipeline, React plugin integration, base path, and local proxy configuration in `vite.config.ts`.
- Gin 1.10.1 - Backend HTTP routing, middleware, CORS, multipart upload handling, and API groups in `backend-go/main.go` and `backend-go/handler/`.
- GORM 1.30.0 with SQLite driver 1.6.0 - Database connection, AutoMigrate, model persistence, and query/update patterns in `backend-go/database/` and `backend-go/service/`.
- Tailwind CSS 3.4.17 - Utility-first styling with dark-mode class support and Zinc color aliasing in `tailwind.config.js` and `src/index.css`.
- Radix UI primitives - Accessible UI foundations for dialogs, alert dialogs, dropdown menus, popovers, selects, switches, tabs, tooltips, labels, scroll areas, and separators in `src/components/ui/`.

**Testing:**
- Vitest 4.1.5 - Frontend test runner via `npm test` and `npm run test:watch` in `package.json`; tests are colocated as `src/store.test.ts`, `src/lib/db.test.ts`, `src/lib/mask.test.ts`, `src/lib/maskPreprocess.test.ts`, and `src/lib/viewportTransform.test.ts`.
- Go standard testing - Backend unit tests use `_test.go` files such as `backend-go/config/config_test.go`, `backend-go/handler/images_test.go`, and `backend-go/service/image_test.go`.

**Build/Dev:**
- TypeScript compiler 5.8.3 - `npm run build` runs `tsc -b` before `vite build`; compiler settings live in `tsconfig.json`.
- @vitejs/plugin-react 4.4.1 - React transform and Fast Refresh integration in `vite.config.ts`.
- PostCSS 8.5.3 and Autoprefixer 10.4.21 - CSS post-processing configured by `postcss.config.js`.
- Vite dev proxy - Optional local proxy loaded from `dev-proxy.config.json`, normalized by `src/lib/devProxy.ts`, and injected into `vite.config.ts` as `__DEV_PROXY_CONFIG__`.

## Key Dependencies

**Critical:**
- `github.com/openai/openai-go/v3` 3.34.0 - OpenAI-compatible Images API SDK used by `backend-go/service/openai.go` for image generation and image edits; `backend-go/go.mod` lists it as an indirect dependency while source imports it directly.
- `github.com/gin-gonic/gin` 1.10.1 - Core backend web framework used in `backend-go/main.go` and all handlers under `backend-go/handler/`.
- `github.com/gin-contrib/cors` 1.7.6 - Global CORS middleware configured in `backend-go/main.go` with all origins allowed and Authorization headers enabled.
- `github.com/golang-jwt/jwt/v5` 5.2.2 - HS256 JWT signing and verification in `backend-go/service/auth.go` and enforcement in `backend-go/middleware/middleware.go`.
- `gorm.io/gorm` 1.30.0, `gorm.io/driver/sqlite` 1.6.0, and `github.com/mattn/go-sqlite3` 1.14.27 - Persistent backend database layer in `backend-go/database/database.go` and model definitions in `backend-go/database/models.go`.
- `react` 19.1.0 and `react-dom` 19.1.0 - Frontend rendering and component lifecycle in `src/main.tsx` and `src/`.
- `zustand` 5.0.5 - Global SPA state, persistence, task coordination, and image cache orchestration in `src/store.ts`.

**Infrastructure:**
- `github.com/google/uuid` 1.6.0 - Backend ID generation support in `backend-go/util/id.go`.
- `lucide-react` 1.16.0 - Icon set used throughout `src/components/` and `src/admin/`.
- `sonner` 2.0.7 - Toast notifications through `src/components/Toast.tsx` and `src/components/ui/sonner.tsx`.
- `class-variance-authority` 0.7.1 - Variant class composition in shadcn-style UI primitives such as `src/components/ui/button.tsx` and `src/components/ui/badge.tsx`.
- `clsx` 2.1.1 and `tailwind-merge` 3.6.0 - Class composition helper in `src/lib/utils.ts`, consumed by UI primitives in `src/components/ui/`.
- `@radix-ui/react-*` packages - Accessible headless primitives used by wrappers in `src/components/ui/`.

## Configuration

**Environment:**
- Frontend backend base URL comes from `import.meta.env.VITE_BACKEND_URL` with fallback `http://localhost:3001` in `src/lib/backendApi.ts`, `src/admin/adminApi.ts`, and `src/store.ts`.
- Backend runtime configuration is loaded from `config.json` in the Go process working directory by `backend-go/config/config.go`; when run from `backend-go/`, this corresponds to `backend-go/config.json`.
- `backend-go/config.json` is present and `.gitignore` marks it as containing secrets; keep values out of codebase maps and commits. Configure `port`, `jwtSecret`, `adminApikey`, `model`, `apiMode`, `timeout`, `codexCli`, and `apiEndpoints` there.
- API endpoint pool configuration is also mutable through admin routes in `backend-go/handler/admin.go`; `backend-go/config/config.go` persists `apiEndpoints` back to `config.json`.
- Optional local dev proxy configuration lives in `dev-proxy.config.json`, is loaded only for `vite serve` in `vite.config.ts`, and is gitignored by `.gitignore`.
- `.env*`, `.nvmrc`, `.node-version`, `.python-version`, `requirements.txt`, `pyproject.toml`, and `Cargo.toml` are not detected at the repository root.
- README-documented `VITE_DEFAULT_API_URL` and Docker `API_URL` appear in `README.md`; source references for the current frontend backend connection use `VITE_BACKEND_URL` in `src/lib/backendApi.ts`, `src/admin/adminApi.ts`, and `src/store.ts`.
- Project skills are not detected under `.claude/skills/` or `.agents/skills/`.

**Build:**
- `package.json` scripts: `npm run dev` starts Vite, `npm run build` runs `tsc -b && vite build`, `npm run preview` starts Vite preview, `npm test` runs Vitest once, and `npm run test:watch` runs Vitest watch mode.
- `tsconfig.json` targets ES2020, uses `moduleResolution: "bundler"`, enables `strict`, uses `jsx: "react-jsx"`, and includes only `src`.
- `vite.config.ts` sets `base: './'`, installs `@vitejs/plugin-react`, injects `__DEV_PROXY_CONFIG__`, enables `server.host`, and configures optional proxy rewriting from `dev-proxy.config.json`.
- `tailwind.config.js` enables `darkMode: 'class'`, scans `index.html` and `src/**/*.{js,ts,jsx,tsx}`, aliases `gray` to Tailwind Zinc colors, and configures CSS-variable-backed sans/mono font families.
- `postcss.config.js` installs `tailwindcss` and `autoprefixer`.
- `backend-go/go.mod` declares module `gpt-image-playground/backend` with Go 1.23.0 and backend package versions.

## Platform Requirements

**Development:**
- Install frontend dependencies with npm using `package.json` and `package-lock.json`; run the frontend with `npm run dev` from the repository root.
- Run tests with `npm test` for frontend Vitest suites and `go test ./...` from `backend-go/` for backend tests.
- Build frontend static assets with `npm run build`; output is `dist/`.
- Use Go 1.23-compatible tooling for `backend-go/`; `github.com/mattn/go-sqlite3` requires a CGO-capable build environment for the SQLite driver.
- Provide a backend `config.json` for non-default secrets and endpoint pool values before running `backend-go/main.go`; keep `backend-go/config.json` uncommitted.
- Use a modern browser with IndexedDB, localStorage, Cache API, service workers, FileReader, and Web Crypto support for full frontend behavior in `src/lib/db.ts`, `src/store.ts`, and `public/sw.js`.

**Production:**
- Frontend deployable artifact is `dist/` from Vite; static hosting references appear in `README.md` for Vercel, GitHub Pages, Nginx, Netlify, and GHCR Docker usage.
- Backend production service is the Go Gin server in `backend-go/main.go`; it requires a writable data directory for SQLite at `data/app.sqlite` and writable upload storage under `upload/` relative to the Go process working directory.
- Configure frontend deployments with `VITE_BACKEND_URL` so `src/lib/backendApi.ts`, `src/admin/adminApi.ts`, and `src/store.ts` point at the deployed backend instead of the localhost fallback.
- No repository-root `Dockerfile`, `docker-compose*.yml`, `.github/workflows/`, or `vercel.json` is detected in the current checkout.

---

*Stack analysis: 2026-05-22*
