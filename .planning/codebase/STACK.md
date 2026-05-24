# Technology Stack

**Analysis Date:** 2026-05-24

## Languages

**Primary:**
- TypeScript 5.8.3 - React frontend and admin UI in `src/**/*.ts` and `src/**/*.tsx`; compiler settings live in `tsconfig.json`.
- Go 1.25.0 - API backend in `backend-go/**`; module declared in `backend-go/go.mod` as `gpt-image-playground/backend`.

**Secondary:**
- CSS / Tailwind CSS - application styling in `src/index.css`, Tailwind theme configuration in `tailwind.config.js`, and PostCSS wiring in `postcss.config.js`.
- JavaScript config files - build and UI tooling configuration in `tailwind.config.js`, `postcss.config.js`, and `components.json`.
- JSON configuration - frontend/package metadata in `package.json` and backend runtime configuration schema in `backend-go/config/config.go`.

## Runtime

**Environment:**
- Browser SPA/PWA - React is mounted from `src/main.tsx`, service worker registration is controlled in `src/main.tsx`, app-shell caching lives in `public/sw.js`, and PWA metadata lives in `public/manifest.webmanifest`.
- Node.js - required for Vite development/build/test commands in `package.json`; no `engines` field is declared in `package.json`. Local toolchain observed: Node.js v22.21.1.
- Go native HTTP server - Gin server entry point in `backend-go/main.go`; `backend-go/go.mod` declares Go 1.25.0. Local toolchain observed: go1.25.10 linux/amd64.

**Package Manager:**
- npm 11.15.0 - root frontend package manager.
- Lockfile: present at `package-lock.json`.
- Go modules - backend dependency management via `backend-go/go.mod` and `backend-go/go.sum`.

## Frameworks

**Core:**
- React 19.1.0 - SPA rendering and component model in `src/main.tsx`, `src/App.tsx`, `src/components/**`, and `src/admin/**`.
- React DOM 19.1.0 - browser root mounting in `src/main.tsx`.
- Vite 6.3.2 - frontend dev server and static build, configured in `vite.config.ts`.
- Gin 1.10.1 - Go HTTP routing, middleware, and JSON APIs in `backend-go/main.go` and `backend-go/handler/**`.
- GORM 1.30.0 with SQLite driver 1.6.0 - persistence layer in `backend-go/database/database.go` and models in `backend-go/database/models.go`.
- Zustand 5.0.5 - frontend state management and persistence in `src/store.ts`.
- Tailwind CSS 3.4.17 - utility-first styling configured in `tailwind.config.js` and consumed from `src/index.css`.
- Radix UI primitives - dialog, alert dialog, dropdown, label, popover, scroll area, select, separator, switch, tabs, and tooltip components under `src/components/ui/**`.
- shadcn-style component setup - UI aliases and Tailwind integration declared in `components.json`; generated/maintained UI components live in `src/components/ui/**`.

**Testing:**
- Vitest 4.1.5 - frontend unit tests under `src/**/*.test.ts` and `src/**/*.test.tsx`; commands defined in `package.json`.
- Go standard `testing` package - backend tests under `backend-go/**/*_test.go`.
- In-memory / temp SQLite testing - backend tests use `gorm.io/driver/sqlite` in files such as `backend-go/database/models_test.go` and `backend-go/handler/auth_handler_test.go`.

**Build/Dev:**
- TypeScript build mode - `npm run build` executes `tsc -b && vite build` from `package.json`.
- @vitejs/plugin-react 4.4.1 - React transform plugin configured in `vite.config.ts`.
- PostCSS 8.5.3 and Autoprefixer 10.4.21 - CSS processing configured in `postcss.config.js`.
- tailwindcss-animate 1.0.7 - Tailwind animation plugin loaded from `tailwind.config.js`.
- Vite dev proxy - optional local proxy loaded by `vite.config.ts` through `src/lib/devProxy.ts` and `dev-proxy.config.json`.

## Key Dependencies

**Critical:**
- `github.com/openai/openai-go/v3` 3.34.0 - OpenAI-compatible image generation/edit client in `backend-go/service/openai.go`.
- `github.com/gin-gonic/gin` 1.10.1 - backend routing and request handling in `backend-go/main.go` and `backend-go/handler/**`.
- `gorm.io/gorm` 1.30.0 and `gorm.io/driver/sqlite` 1.6.0 - SQLite ORM layer in `backend-go/database/database.go`.
- `github.com/golang-jwt/jwt/v5` 5.2.2 - HS256 JWT signing and verification in `backend-go/service/auth.go` and `backend-go/middleware/middleware.go`.
- `golang.org/x/crypto` 0.52.0 - bcrypt password hashing in `backend-go/service/auth.go`.
- `zustand` 5.0.5 - persisted frontend app state in `src/store.ts`.
- `react` 19.1.0 and `react-dom` 19.1.0 - core UI runtime in `src/main.tsx` and `src/App.tsx`.

**Infrastructure:**
- `github.com/gin-contrib/cors` 1.7.6 - permissive CORS middleware configured in `backend-go/main.go`.
- `sonner` 2.0.1 - toast notifications wired through `src/store.ts` and rendered by `src/components/ui/sonner.tsx`.
- `lucide-react` 0.468.0 - icon set used by React components in `src/components/**` and `src/admin/**`.
- `class-variance-authority`, `clsx`, and `tailwind-merge` - class composition utilities for UI components in `src/components/ui/**` and `src/lib/utils.ts`.
- Browser IndexedDB API - image persistence helper in `src/lib/db.ts`.
- Browser Cache API and Service Worker API - PWA app-shell caching in `public/sw.js`.
- Browser `localStorage` - auth token storage in `src/lib/backendApi.ts` and `src/admin/adminApi.ts`; Zustand persistence key configured in `src/store.ts`.

## Configuration

**Environment:**
- Frontend backend URL is configured with optional `VITE_BACKEND_URL`; it is read in `src/lib/backendApi.ts`, `src/admin/adminApi.ts`, and `src/store.ts`, with a development default pointing at the local Go backend.
- Vite built-ins `import.meta.env.PROD` and `import.meta.env.BASE_URL` control production-only service worker registration in `src/main.tsx`.
- Backend configuration is file-based, not environment-variable-based. `backend-go/config/config.go` loads `backend-go/config.json` from the backend working directory.
- `backend-go/config.json` is present and ignored by git according to `.gitignore`; treat it as the secrets/config store for `jwtSecret`, `adminApikey`, `apiEndpoints[].apiKey`, endpoint base URLs, model, API mode, timeout, pricing, and invite settings.
- `.env`, `.env.*`, and `*.env` files are not detected in the repository scan.
- `dev-proxy.config.json` is present and ignored by git according to `.gitignore`; `vite.config.ts` reads it only for `vite serve` and uses `src/lib/devProxy.ts` to normalize proxy settings.
- Project-specific skill directories are not detected at `.claude/skills/` or `.agents/skills/`.

**Build:**
- Frontend scripts in `package.json`: `npm run dev` starts Vite, `npm run build` runs TypeScript build plus Vite build, `npm run preview` serves the production build, and `npm test` runs Vitest once.
- TypeScript compiler options in `tsconfig.json`: `target` ES2020, DOM libs enabled, `moduleResolution` `bundler`, `jsx` `react-jsx`, and `strict` mode enabled.
- Vite configuration in `vite.config.ts`: React plugin, relative `base: './'`, optional dev proxy, and `__DEV_PROXY_CONFIG__` define injection.
- Tailwind configuration in `tailwind.config.js`: class-based dark mode, content paths `index.html` and `src/**/*.{js,ts,jsx,tsx}`, CSS-variable color tokens, and `tailwindcss-animate` plugin.
- shadcn-style aliases in `components.json`: `components` -> `src/components`, `utils` -> `src/lib/utils`, `ui` -> `src/components/ui`, `lib` -> `src/lib`, and `hooks` -> `src/hooks`.
- Backend dependency and build configuration lives in `backend-go/go.mod` and `backend-go/go.sum`; no Makefile, Dockerfile, or CI workflow files are detected in the repository scan.

## Platform Requirements

**Development:**
- Install frontend dependencies with npm from the repository root containing `package.json` and `package-lock.json`.
- Run frontend dev server with `npm run dev`; Vite serves `index.html` and routes `/admin` to the lazy-loaded admin bundle in `src/main.tsx`.
- Use Go 1.25.x for backend development from `backend-go/`; the backend entry point is `backend-go/main.go`.
- Provide a local `backend-go/config.json` before using authenticated/admin/OpenAI-backed flows; the schema is defined by `backend-go/config/config.go`.
- Runtime directories are created by `backend-go/util/paths.go`: `backend-go/data/` for SQLite and `backend-go/upload/` for image files.

**Production:**
- Frontend builds to static assets in `dist/` via `npm run build`; `README.md` documents static hosting options, but no deployment manifest is detected in the repository scan.
- Backend runs as a Gin HTTP server on the configured port from `backend-go/config/config.go` and `backend-go/config.json`.
- Persist backend runtime storage: SQLite files under `backend-go/data/` and uploaded/generated images under `backend-go/upload/`.
- Configure at least one OpenAI-compatible endpoint in `apiEndpoints` for image generation; endpoint scheduling and failover are implemented in `backend-go/service/openai.go` and `backend-go/service/queue.go`.
- No external cache, object store, queue service, CI workflow, Dockerfile, or hosted deployment configuration is detected in the repository scan.

---

*Stack analysis: 2026-05-24*
