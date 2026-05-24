# Coding Conventions

**Analysis Date:** 2026-05-24

## Naming Patterns

**Files:**
- React components: PascalCase with `.tsx` extension (e.g., `LoginModal.tsx`, `Header.tsx`, `TaskGrid.tsx`)
- Utility modules: camelCase with `.ts` extension (e.g., `backendApi.ts`, `canvasImage.ts`, `viewportTransform.ts`)
- Type definitions: `types.ts` at module level
- Test files: co-located `*.test.ts` or `*.test.tsx` next to the source (e.g., `store.test.ts`, `AdminDashboard.test.tsx`)

**Go source:**
- Package names: lowercase single word matching directory (e.g., `handler`, `service`, `config`, `middleware`)
- Source filenames: lowercase with underscores to separate words (e.g., `admin_handler_test.go`, `openai_failover_test.go`)
- Test files: `*_test.go` suffix, co-located in same package (e.g., `config/config_test.go`, `service/auth_test.go`)

**TypeScript Functions:**
- camelCase for all functions (e.g., `submitGenerateTask`, `getBackendToken`, `bootstrapBackendSession`)
- Exported API functions: descriptive verb-noun pairs (e.g., `deleteRemoteTask`, `uploadImage`, `streamTaskStatus`)
- Internal helpers: private (non-exported, e.g., `buildUrl`, `request`, `dataUrlToBlob`)

**Go Functions:**
- Exported: PascalCase (e.g., `LoginWithPassword`, `VerifyToken`, `GenerateImage`)
- Unexported: camelCase (e.g., `initAdmin`, `codexPrompt`, `mergeConcurrentResults`)
- Test helpers: camelCase with `Test` prefix for table-driven sub-tests following Go convention

**Variables:**
- React state setters: `[value, setValue]` pattern from `useState` (e.g., `const [loading, setLoading] = useState(false)`)
- Zustand selectors: single-letter abbreviation `(s) => s.field` (e.g., `useStore((s) => s.authUser)`)
- Module-level constants: UPPER_SNAKE_CASE (e.g., `DEFAULT_SETTINGS`, `DEFAULT_PARAMS`, `TOKEN_KEY`, `POLL_INTERVAL`)
- Private module-level state: camelCase (e.g., `imageCache`, `imageContentFetches`, `activeStreams`)

**Types/Interfaces:**
- React component props: interface with `Props` suffix (e.g., `RegisterModalProps`)
- Zustand store shape: `interface AppState` in `src/store.ts`
- Data structures/types: interfaces in `src/types.ts` (e.g., `TaskRecord`, `AppSettings`, `InputImage`, `MaskDraft`)
- API response types: exported interfaces (e.g., `AuthUser` in `src/lib/backendApi.ts`, `PricingConfigResponse` in `src/admin/adminApi.ts`)
- Go structs: PascalCase with JSON tags (e.g., `type Config struct`, `type User struct`, `type ApiEndpoint struct`)
- No class components are used; all components are functional

## Component Patterns

**React component structure** follows a consistent pattern (see `src/components/LoginModal.tsx`, `src/components/RegisterModal.tsx`, `src/components/Header.tsx`):

```tsx
import { useState } from 'react'
import { foo } from '../lib/backendApi'
import { useStore } from '../store'
import { UiComponent } from './ui/component'

export default function ComponentName() {
  // 1. Store selectors (useStore)
  const settings = useStore((s) => s.settings)

  // 2. Local state (useState)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  // 3. Event handlers (async with try/catch)
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      await someApi()
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err))
    } finally {
      setLoading(false)
    }
  }

  // 4. Render JSX using shadcn/ui components and Tailwind
  return (
    <div className="...">
      <UiComponent />
    </div>
  )
}
```

**Key component conventions:**
- Default export for all components (`export default function X()`)
- No prop types used; TypeScript interfaces serve this purpose
- Store access via Zustand selectors at top of component body (not inside callbacks)
- Event handlers defined inline as arrow functions (not class methods)
- All Dialogs/Modals use conditional rendering: `{showX && <XModal onClose={() => setShowX(false)} />}`
- shadcn/ui components are wrapped in `src/components/ui/` directory (e.g., `Button`, `Input`, `Dialog`, `Tabs`)
- Props typing uses `interface XProps` (e.g., `interface RegisterModalProps { onClose: () => void }`)
- Composition over inheritance: UI is built by composing smaller components rather than large monolithic components

**shadcn/ui pattern** (see `src/components/ui/button.tsx`):
```tsx
import * as React from 'react'
import { Slot } from '@radix-ui/react-slot'
import { cva, type VariantProps } from 'class-variance-authority'
import { cn } from '../../lib/utils'

const buttonVariants = cva('base-classes', {
  variants: { variant: {...}, size: {...} },
  defaultVariants: { variant: 'default', size: 'default' },
})

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement>, VariantProps<typeof buttonVariants> {
  asChild?: boolean
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, asChild = false, ...props }, ref) => {
    const Comp = asChild ? Slot : 'button'
    return <Comp className={cn(buttonVariants({ variant, size, className }))} ref={ref} {...props} />
  },
)
Button.displayName = 'Button'

export { Button, buttonVariants }
```

## State Management

**Primary tool: Zustand v5** (single store in `src/store.ts`)

Pattern:
```typescript
import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface AppState {
  field: Type
  setField: (value: Type) => void
}

export const useStore = create<AppState>()(
  persist(
    (set, get) => ({
      field: defaultValue,
      setField: (field) => set({ field }),
    }),
    {
      name: 'gpt-image-playground',
      partialize: (state) => ({
        // Only persist specific fields to localStorage
        settings: state.settings,
        authUser: state.authUser,
        params: state.params,
        // ... other persisted slices
      }),
    },
  ),
)
```

**Key patterns:**
- Single monolithic store (not sliced) -- all state in `interface AppState` in `src/store.ts`
- `persist` middleware persists auth, settings, params, and dismissed states to localStorage
- Selectors use identity-function pattern: `useStore((s) => s.field)` for individual field access
- State updates use functional updaters via `set((st) => ({ ... }))` for derived/conditional updates
- Store actions (non-trivial async logic) are exported as standalone functions outside the store (e.g., `submitTask()`, `bootstrapBackendSession()`, `initStore()`)
- These action functions use `useStore.getState()` to read/write state imperatively rather than being bound to React's render cycle
- Global module-level state: `imageCache` (Map), `imageContentFetches` (Map), `activeStreams` (Map), `pollTimer` (setInterval handle) -- all in `src/store.ts`
- SSE streams managed via AbortController map stored outside React tree
- Confirm dialog pattern: transient UI state (`confirmDialog` field in store) with action/cancelAction callbacks

**Local component state:**
- `useState` used for form fields, loading/error states, and modal visibility toggles within components
- No React Context usage detected (Zustand serves as the single global state container)

## Styling

**Approach: Tailwind CSS 3 with shadcn/ui design system**

- Tailwind utilities directly in className (no CSS Modules, no styled-components)
- CSS custom properties (HSL variables) for theming via `:root` and `.dark` in `src/index.css`
- `tailwind.config.js`: maps `gray` to `zinc`, defines semantic tokens (border, input, ring, background, foreground, primary, secondary, etc.)
- Dark mode: class-based (`darkMode: 'class'`), toggled via `document.documentElement.classList.toggle('dark', ...)` in `src/App.tsx`
- Custom animations defined as `@keyframes` in `src/index.css` (modal-in, fade-in, zoom-in, confirm-in, dropdown-down/up) with utility classes (`animate-modal-in`, etc.)
- Custom CSS classes for: safe area insets (`.safe-area-x`, `.safe-area-top`), scrollbar styling, collapse animations, drag selection prevention
- External fonts: HarmonyOS Sans SC (UI), Maple Mono (mono)
- `cn()` utility in `src/lib/utils.ts` merges Tailwind classes using `clsx` + `tailwind-merge`

**Class naming:**
- Utility-first: Tailwind classes used directly, no custom component class naming convention beyond the design system tokens
- Custom utility classes: kebab-case (e.g., `.safe-area-x`, `.hide-scrollbar`, `.mask-edge-r`)
- Animation classes: prefixed with `animate-` (e.g., `.animate-modal-in`, `.animate-dropdown-up`)
- Data attributes: used sparingly for behavior hooks (e.g., `data-no-drag-select`, `data-home-main`, `data-scroll-locked`)

## API Patterns

**Frontend: Native fetch with thin wrapper** (see `src/lib/backendApi.ts`)

```typescript
const API_BASE_URL = import.meta.env.VITE_BACKEND_URL?.trim()?.replace(/\/+$/, '') || 'http://localhost:3001'

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers)
  const token = getBackendToken()
  if (token) headers.set('Authorization', `Bearer ${token}`)
  if (options.body && !(options.body instanceof FormData) && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }
  const response = await fetch(buildUrl(path), { ...options, headers, cache: 'no-store' })
  if (!response.ok) {
    let message = `HTTP ${response.status}`
    try { const payload = await response.json(); message = payload.error || payload.message || message } catch { message = await response.text() }
    throw new Error(message)
  }
  return response.json() as Promise<T>
}
```

**Key patterns:**
- Token stored in localStorage under key `'gpt-image-playground-admin-token'` (admin) or `'gpt-image-playground-token'` (user)
- Token sent as `Authorization: Bearer <token>` header
- Error handling: errors thrown as `Error` instances, caught with `err instanceof Error ? err.message : String(err)` in components
- Image upload uses `FormData` with `multipart/form-data` (Content-Type auto-set by browser)
- SSE streaming: raw `fetch` + `response.body.getReader()` with manual `TextDecoder` parsing for Server-Sent Events
- Public endpoints: `fetch` used directly without auth header (e.g., `getPublicConfig`, `getPublicAnnouncement`)
- No axios, no fetch library abstraction beyond the internal `request()` helper
- `cache: 'no-store'` on all fetch requests to avoid stale data

**Backend API (Gin):**
- Route groups: `/api/auth`, `/api/images`, `/api/tasks`, `/api`, `/api/admin`
- Auth middleware: extracts Bearer token from header or query param, verifies JWT, sets user context
- Admin middleware: verifies admin role from JWT
- All responses use `gin.H` maps (e.g., `gin.H{"ok": true}`, `gin.H{"error": "message"}`)
- Request bodies validated inline with `c.ShouldBindJSON()` and manual field checks
- Error responses in Chinese (user-facing i18n)
- Health check: `GET /api/health` returns `{"ok": true}`

## Go Backend Patterns

**Architecture: Handler -> Service -> Database (GORM)**

**Handler layer** (`backend-go/handler/`):
- One file per feature area: `auth.go`, `generate.go`, `admin.go`, `images.go`, `tasks.go`, `feedback.go`, `announcement.go`, `changelog.go`, `config.go`
- Each handler function receives `*gin.Context`, binds JSON, calls service, returns JSON
```go
func HandlerName(c *gin.Context) {
    var body struct { Field string `json:"field"` }
    if err := c.ShouldBindJSON(&body); err != nil { /* 400 */ return }
    result, err := service.Function(body.Field)
    if err != nil { /* error response */ return }
    c.JSON(http.StatusOK, gin.H{"key": result})
}
```
- User extraction: `user := middleware.GetAuthUser(c)`
- Task execution: submitted synchronously (task created in DB), then `go executeImageGeneration(...)` runs async with concurrency slots
- No explicit DI; service layer accessed directly as package functions

**Service layer** (`backend-go/service/`):
- Stateless package-level functions operating on GORM `database.DB`
- Database operations use GORM query builder (e.g., `database.DB.Where("id = ?", id).First(&user).Error`)
- Error pattern: return `(result, error)`, callers check `if err != nil`
- Transaction pattern: not widely used; single-row operations by ID
- Config access: `config.App` global singleton, `config.GetEndpointPool()` for thread-safe reads

**Middleware** (`backend-go/middleware/`):
- `AuthMiddleware()`: validates JWT, fetches user from DB, sets `c.Set("user", &service.AuthUser{...})`
- `AdminMiddleware()`: validates admin role from JWT
- `RequestLogger()`: logging middleware (defined in `middleware/logger.go`)
- AuthUser retrieval: `middleware.GetAuthUser(c)` extracts from Gin context

**Error handling:**
- Structured logging with `slog.Warn`/`slog.Error` (key-value pairs)
- User-facing errors in Chinese returned in `gin.H{"error": "..."}`
- Internal errors logged to slog, user gets generic message
- No panic-based error handling; all errors handled via return values

**Database patterns:**
- SQLite via GORM (`gorm.io/driver/sqlite`)
- Models defined as structs with GORM tags in `backend-go/database/models.go`
- Single DB connection (`database.DB *gorm.DB`), `SetMaxOpenConns(1)` for SQLite
- AutoMigrate on startup for all models
- ID generation via `util.GenerateID()` (likely ULID/UUID)

## TypeScript Conventions

**Configuration** (see `tsconfig.json`):
- `strict: true` -- full strict mode enabled
- `noUnusedLocals: false` -- unused local variables not enforced
- `noUnusedParameters: false` -- unused parameters not enforced
- `noFallthroughCasesInSwitch: true`
- `verbatimModuleSyntax: true` -- explicit `import type` required for type-only imports
- Target: `ES2020`, module: `ESNext`, jsx: `react-jsx`

**Import conventions:**
- Type-only imports use `import type { ... }` syntax (enforced by verbatimModuleSyntax)
- Import order (observed from `src/store.ts`, `src/App.tsx`, `src/components/LoginModal.tsx`):
  1. React / framework imports first
  2. Third-party library imports (zustand, sonner, lucide-react)
  3. Local type imports (`import type { ... } from './types'`)
  4. Local module imports (`import { ... } from './lib/backendApi'`)
  5. Component imports (`import Component from './components/Component'`)
  6. UI component imports (`import { Button } from './components/ui/button'`)
  7. Style imports (CSS) -- only in entry point (`src/main.tsx`)
- Path aliases: shadcn/ui configured in `components.json` (e.g., `@/components/ui` maps to `src/components/ui`), but imports use relative paths
- Absolute path aliases: no `~` or `@` TS path aliases configured in tsconfig

**Interfaces vs types:**
- `interface` used for object shapes that may be extended (component props, data structures like `TaskRecord`, `AppSettings`)
- `type` used for union types and aliases (`TaskStatus`, `ApiMode`, `ThemeMode`, `BugFeedbackCategory`)
- Both styles co-exist; preference is:
  - Exported data models: `interface`
  - Union/enum-like values: `type`
  - Component props: `interface`

**Generics:**
- `request<T>(...)`: generic API call wrapper
- `dbTransaction<T>(...)`: generic IndexedDB transaction wrapper
- Zustand `setSelectedTaskIds` accepts union type `string[] | ((prev: string[]) => string[])`
- Minimal generics usage overall

## Git Conventions

Based on recent commit history:
- Commit messages: English, conventional-commit-like prefix with `:` separator
  - `feat:` for features (e.g., `feat: adopt shadcn/ui components, sonner toast, and enhance invite system`)
  - `fix:` for bug fixes (e.g., `fix: replace white-opacity dark backgrounds with gray-900/gray-800 in admin`)
- Messages start with lowercase after the prefix
- Detail-oriented descriptions mentioning specific components/files changed
- No conventional commit scope notation (no `feat(scope):`)

## Formatting and Linting

- **No ESLint configuration detected** -- no `.eslintrc*`, `eslint.config.*`, or `biome.json` in project
- **No Prettier configuration detected** -- no `.prettierrc*` or `prettier.config.*` in project  
- Code style is maintained through convention alone
- Indentation: 2 spaces (TypeScript), tabs (Go -- standard `gofmt`)
- Go code likely follows `gofmt` standard formatting (standard for Go projects)
- No editor config file (`.editorconfig`) detected

## File Size and Composition Guidelines

- Store file (`src/store.ts`): ~937 lines, largest single file -- contains all Zustand state, image cache logic, task lifecycle, and helper functions
- Component files: generally 50-150 lines each (`LoginModal.tsx`: 145 lines, `RegisterModal.tsx`: 100 lines, `Header.tsx`: 97 lines)
- UI components (`src/components/ui/`): 30-60 lines each
- API module (`src/lib/backendApi.ts`): ~298 lines -- contains all API functions
- Type definitions (`src/types.ts`): ~145 lines
- Go handler files: 200-425 lines (`admin.go`: 425 lines largest handler, `generate.go`: 286 lines)
- Go service files: 150-370 lines (`openai.go`: 370 lines, `auth.go` tests: ~988 lines)
- No explicit file size limit enforced, but files generally stay under 500 lines (except `store.ts` and `auth_test.go`)
- Component composition: Modals/features extracted into separate files rather than inlining into parent components

## Module Design

**Exports:** Named exports for UI components and utilities, default exports for React components
```typescript
// shadcn/ui components: named exports
export { Button, buttonVariants }

// Feature components: default exports
export default function LoginModal() { ... }

// Utilities: named exports
export function cn(...inputs: ClassValue[]) { ... }
export function hashDataUrl(dataUrl: string): Promise<string> { ... }
```

**Barrel files:** Not used. Each file imports directly from its dependency.

## Custom Hooks

- `src/hooks/useCloseOnEscape.ts` -- single custom hook in the project
- Primary state management through Zustand store, not React hooks
- Local component state via `useState`, side effects via `useEffect`
- No `useReducer`, `useMemo`, or `useCallback` observed in components (simple state, no perf optimization needed)

---

*Convention analysis: 2026-05-24*
