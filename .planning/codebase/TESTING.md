# Testing Patterns

**Analysis Date:** 2026-05-24

## Test Framework

**Runner:**
- Frontend: Vitest `4.1.5`, invoked through scripts in `package.json`. No separate `vitest.config.*` file is detected; tests run with Vite/Vitest defaults from the project root.
- Backend: Go built-in `testing` package under the `backend-go/` module, using `httptest`, Gin test mode, temporary SQLite databases, and direct GORM seeding.
- Frontend source-check tests use Vite raw imports (`?raw`) to assert component/source content without rendering. Examples: `src/components/LoginModal.test.tsx`, `src/components/RegisterModal.test.tsx`, `src/components/MigrationModal.test.tsx`, `src/admin/AdminDashboard.test.tsx`, `src/App-test.test.tsx`.

**Assertion Library:**
- Frontend: Vitest `expect`, `vi`, `describe`, `it`, `beforeEach`, `afterEach`, and `vi.waitFor` from `vitest`.
- Backend: Go standard assertions with `t.Fatalf`, `t.Errorf`, and direct comparisons. No testify-style assertion dependency is used.

**Run Commands:**
```bash
npm test              # Run all frontend Vitest tests once
npm run test:watch    # Run frontend Vitest in watch mode
npm run build         # Type-check with tsc -b and build with Vite
cd backend-go && go test ./...  # Run all backend Go tests
```

## Test File Organization

**Location:**
- Frontend tests are co-located beside implementation files in `src/`: `src/store.test.ts`, `src/lib/mask.test.ts`, `src/lib/backendApi.test.ts`, `src/admin/moneyFormat.test.ts`, `src/components/LoginModal.test.tsx`.
- Backend tests are co-located beside package implementation files in `backend-go/`: `backend-go/service/auth_test.go`, `backend-go/handler/admin_handler_test.go`, `backend-go/database/models_test.go`, `backend-go/config/config_test.go`.
- There is no separate top-level `tests/` directory. Add new tests next to the code under test.

**Naming:**
- Frontend: use `*.test.ts` for non-JSX modules and `*.test.tsx` for component/source tests: `src/lib/maskPreprocess.test.ts`, `src/components/MigrationModal.test.tsx`.
- Backend: use `*_test.go`, grouped by package and feature: `backend-go/service/openai_failover_test.go`, `backend-go/handler/generate_billing_test.go`, `backend-go/handler/admin_analytics_test.go`.
- Test names should describe behavior, not implementation. Frontend examples include `it('stores token in localStorage on success')` in `src/lib/backendApi.test.ts`; backend examples should use `TestXxx` plus subtests where helpful in `backend-go/service/*_test.go`.

**Structure:**
```text
src/
├── store.ts
├── store.test.ts
├── lib/
│   ├── backendApi.ts
│   ├── backendApi.test.ts
│   ├── mask.ts
│   ├── mask.test.ts
│   ├── maskPreprocess.ts
│   ├── maskPreprocess.test.ts
│   └── viewportTransform.test.ts
├── components/
│   ├── LoginModal.tsx
│   ├── LoginModal.test.tsx
│   ├── RegisterModal.test.tsx
│   └── MigrationModal.test.tsx
└── admin/
    ├── adminApi.ts
    ├── adminApi.test.ts
    ├── adminApi-invite.test.ts
    ├── AdminDashboard.test.tsx
    └── moneyFormat.test.ts

backend-go/
├── config/config_test.go
├── database/models_test.go
├── handler/*_test.go
└── service/*_test.go
```

## Test Structure

**Suite Organization:**
```typescript
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { submitTask, useStore } from './store'
import { submitGenerateTask, getTasks as fetchTasks, streamTaskStatus } from './lib/backendApi'

vi.mock('./lib/backendApi', () => ({
  submitGenerateTask: vi.fn().mockResolvedValue({ taskId: 'task-1', status: 'processing' }),
  getTasks: vi.fn().mockResolvedValue({ tasks: [] }),
  streamTaskStatus: vi.fn().mockImplementation((_taskId: string, onUpdate: Function) => {
    setTimeout(() => {
      const task = useStore.getState().tasks.find((t: any) => t.id === _taskId)
      if (task) onUpdate({ ...task, status: 'done', outputImages: ['img-1'] })
    }, 10)
    return new AbortController()
  }),
}))

describe('submitTask backend submission flow', () => {
  beforeEach(() => {
    vi.mocked(submitGenerateTask).mockReset()
    vi.mocked(submitGenerateTask).mockResolvedValue({ taskId: 'task-1', status: 'processing' })
    vi.mocked(fetchTasks).mockResolvedValue({ tasks: [] })
    useStore.setState({ prompt: 'test prompt', tasks: [] })
  })

  it('creates a queued task immediately', () => {
    submitTask()
    expect(useStore.getState().tasks[0].status).toBe('queued')
  })
})
```

```go
func TestSomeHandlerBehavior(t *testing.T) {
    gin.SetMode(gin.TestMode)
    r := gin.New()
    r.PUT("/api/admin/users/:id/password", handler.AdminResetPassword)

    req := httptest.NewRequest(http.MethodPut, "/api/admin/users/user-1/password", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+token)
    resp := httptest.NewRecorder()

    r.ServeHTTP(resp, req)

    if resp.Code != http.StatusOK {
        t.Fatalf("expected status 200, got %d", resp.Code)
    }
}
```

**Patterns:**
- Frontend unit tests group by module behavior with `describe`, reset mocks/state in `beforeEach`, and assert with `expect`. Follow `src/store.test.ts`, `src/lib/mask.test.ts`, and `src/admin/moneyFormat.test.ts`.
- Frontend API client tests stub `localStorage`, spy on `globalThis.fetch`, inspect URL/method/body/headers, and assert returned typed payloads. Follow `src/lib/backendApi.test.ts`, `src/admin/adminApi.test.ts`, and `src/admin/adminApi-invite.test.ts`.
- Frontend source-check tests import component files with `?raw` and assert source contains required API calls, validation text, or JSX conditions. Follow `src/components/LoginModal.test.tsx`, `src/components/MigrationModal.test.tsx`, and `src/App-test.test.tsx`.
- Backend service tests set up isolated config/data directories with `t.TempDir()`, open a temporary SQLite DB, auto-migrate models, seed records directly through `database.DB`, and clean up DB handles with `t.Cleanup`. Follow `backend-go/service/auth_test.go`, `backend-go/service/billing_test.go`, and `backend-go/service/analytics_test.go`.
- Backend handler tests build Gin routers in test mode, issue `httptest.NewRequest`, set JSON and Authorization headers, serve with `httptest.ResponseRecorder`, then inspect response status/body and database side effects. Follow `backend-go/handler/admin_handler_test.go`, `backend-go/handler/images_test.go`, and `backend-go/handler/admin_analytics_test.go`.

## Mocking

**Framework:**
- Frontend: Vitest `vi.mock`, `vi.fn`, `vi.spyOn`, `vi.stubGlobal`, `vi.mocked`, `vi.restoreAllMocks`, and `vi.unstubAllGlobals`.
- Backend: no mocking framework. Tests use temporary SQLite databases, direct package config overrides, local HTTP test servers/routers, and deterministic seeded data.

**Patterns:**
```typescript
vi.mock('./lib/backendApi', () => ({
  submitGenerateTask: vi.fn().mockResolvedValue({ taskId: 'task-1', status: 'processing' }),
  putRemoteTask: vi.fn().mockResolvedValue({ ok: true }),
  getBackendToken: vi.fn().mockReturnValue('test-token'),
  getMe: vi.fn(),
  getTasks: vi.fn().mockResolvedValue({ tasks: [] }),
  streamTaskStatus: vi.fn().mockImplementation((_taskId: string, onUpdate: Function) => {
    setTimeout(() => {
      const task = useStore.getState().tasks.find((t: any) => t.id === _taskId)
      if (task) {
        onUpdate({ ...task, status: 'done', outputImages: ['img-1'], finishedAt: Date.now(), elapsed: 1000 })
      }
    }, 10)
    return new AbortController()
  }),
}))

vi.stubGlobal('localStorage', {
  getItem: vi.fn((key: string) => store[key] ?? null),
  setItem: vi.fn((key: string, val: string) => { store[key] = val }),
  removeItem: vi.fn((key: string) => { delete store[key] }),
})

vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
  new Response(JSON.stringify({ ok: true }), {
    status: 200,
    headers: { 'Content-Type': 'application/json' },
  }),
)
```

**What to Mock:**
- Mock frontend network boundaries (`fetch`, `src/lib/backendApi.ts`, `src/admin/adminApi.ts`) when testing store/UI behavior. See `src/store.test.ts`.
- Mock browser storage and browser-only APIs (`localStorage`, `FileReader`, `fetch`, `indexedDB`) in frontend unit tests. See `src/lib/backendApi.test.ts`, `src/admin/adminApi.test.ts`, and `src/lib/db.test.ts`.
- Mock SSE/task status updates by replacing `streamTaskStatus` with a fake callback driver returning `new AbortController()`. See `src/store.test.ts`.
- Use fake IndexedDB-style classes when testing `src/lib/db.ts`; the test defines `FakeIDBRequest`, `FakeIDBObjectStore`, `FakeIDBTransaction`, `FakeIDBDatabase`, and `FakeIndexedDB` in `src/lib/db.test.ts`.
- In backend tests, prefer real temporary SQLite DBs and direct seeded rows over mocks. This validates GORM models, transactions, and query behavior in `backend-go/service/*.go` and `backend-go/handler/*.go`.

**What NOT to Mock:**
- Do not mock pure helpers such as `src/lib/mask.ts`, `src/lib/maskPreprocess.ts`, `src/lib/viewportTransform.ts`, `src/admin/moneyFormat.ts`, or `backend-go/service/money.go`; test inputs and outputs directly.
- Do not mock GORM for service tests that exercise database behavior. Use temporary SQLite and `database.DB` setup as shown in `backend-go/service/billing_test.go` and `backend-go/service/auth_test.go`.
- Do not mock Gin routing for handler behavior. Build a real `gin.Engine` in test mode and call it through `httptest`, as in `backend-go/handler/admin_handler_test.go`.
- Do not rely on real external OpenAI/API network calls in automated tests. Use controlled service inputs, helper function tests, or local fake endpoint behavior in `backend-go/service/openai_failover_test.go`.

## Fixtures and Factories

**Test Data:**
```typescript
const imageA = { id: 'image-a', dataUrl: 'data:image/png;base64,a' }

function resetStoreForTest() {
  useStore.setState({
    authUser: { id: 'user-1', label: 'test', role: 'user', imageCount: 0, quota: 0, unlimitedQuota: false, usedCount: 0 },
    prompt: '',
    inputImages: [],
    maskDraft: null,
    maskEditorImageId: null,
    params: { ...DEFAULT_PARAMS },
    tasks: [],
  })
}

function task(overrides: Partial<TaskRecord> = {}): TaskRecord {
  return {
    id: 'task-a',
    prompt: 'prompt',
    params: { ...DEFAULT_PARAMS },
    inputImageIds: [],
    maskTargetImageId: null,
    maskImageId: null,
    outputImages: [],
    status: 'done',
    error: null,
    createdAt: 1,
    finishedAt: 2,
    elapsed: 1,
    ...overrides,
  }
}
```

```go
func setupBillingTest(t *testing.T) {
    t.Helper()

    tmp := t.TempDir()
    config.App = &config.Config{
        DataDir:     filepath.Join(tmp, "data"),
        UploadDir:   filepath.Join(tmp, "upload"),
        JWTSecret:   "test-secret",
        AdminApikey: "test-admin",
    }

    db, err := gorm.Open(sqlite.Open(filepath.Join(tmp, "test.db")), &gorm.Config{})
    if err != nil {
        t.Fatalf("open db: %v", err)
    }
    database.DB = db

    if err := database.DB.AutoMigrate(&database.User{}, &database.Task{}, &database.BillingRecord{}); err != nil {
        t.Fatalf("migrate: %v", err)
    }

    t.Cleanup(func() {
        sqlDB, err := database.DB.DB()
        if err == nil {
            _ = sqlDB.Close()
        }
    })
}
```

**Location:**
- Frontend local factories live inside the test file that needs them. Examples: `resetStoreForTest` and `task` in `src/store.test.ts`; inline DTOs in `src/lib/backendApi.test.ts` and `src/admin/adminApi.test.ts`.
- Backend setup helpers live inside package test files. Examples: auth setup in `backend-go/service/auth_test.go`, billing setup in `backend-go/service/billing_test.go`, handler/router setup in `backend-go/handler/admin_handler_test.go`.
- No shared fixture directory is detected. Prefer local fixtures until duplication becomes difficult to maintain.

## Coverage

**Requirements:**
- No enforced frontend coverage threshold is detected in `package.json` or a Vitest config file.
- No enforced backend coverage threshold is detected in `backend-go/go.mod` or test configuration.
- Coverage exists through targeted unit/service/handler tests rather than a configured percentage gate.
- Known skipped coverage gap: `src/store.test.ts` contains `describe.skip('image cache behavior in store — TODO: update for current store implementation', ...)`. Treat the image cache/backend image warming behavior as under-tested until that suite is updated and re-enabled.

**View Coverage:**
```bash
npx vitest run --coverage       # View frontend coverage if coverage provider is installed/configured
cd backend-go && go test ./... -cover  # View backend package coverage
```

## Test Types

**Unit Tests:**
- Pure frontend helpers should be tested with direct input/output assertions. Use `src/lib/mask.test.ts`, `src/lib/maskPreprocess.test.ts`, `src/lib/viewportTransform.test.ts`, and `src/admin/moneyFormat.test.ts` as patterns.
- Frontend API clients should be unit-tested by mocking `fetch` and browser storage, then asserting request URL, method, headers, body, token persistence, and parsed result. Use `src/lib/backendApi.test.ts`, `src/admin/adminApi.test.ts`, and `src/admin/adminApi-invite.test.ts`.
- Backend pure/domain helpers should be tested with table-driven tests where applicable. Use `backend-go/service/money_test.go` and helper tests in `backend-go/handler/generate_billing_test.go`.

**Integration Tests:**
- Frontend store tests integrate Zustand state, mocked API clients, mocked IndexedDB/image cache boundaries, and async task status callbacks. Use `src/store.test.ts`.
- Backend service tests integrate service logic with a real temporary SQLite database and GORM models. Use `backend-go/service/auth_test.go`, `backend-go/service/billing_test.go`, `backend-go/service/analytics_test.go`, and `backend-go/service/task.go` coverage tests.
- Backend handler tests integrate Gin routing/middleware-style auth headers, request JSON/multipart bodies, service/database behavior, and HTTP responses. Use `backend-go/handler/admin_handler_test.go`, `backend-go/handler/images_test.go`, `backend-go/handler/admin_analytics_test.go`, and `backend-go/handler/admin_pricing_test.go`.

**E2E Tests:**
- No browser E2E framework such as Playwright or Cypress is detected.
- No full-stack E2E test directory is detected.
- Use backend handler tests and frontend API/store tests as the current end-to-end substitute for route/client contract coverage.

## Common Patterns

**Async Testing:**
```typescript
submitTask()

await vi.waitFor(() => {
  const tasks = useStore.getState().tasks
  expect(tasks[0].status).toBe('done')
}, { timeout: 5000 })
```

```typescript
vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
  new Response(JSON.stringify({ token: 'new-token', user, needsMigration: false }), {
    status: 200,
    headers: { 'Content-Type': 'application/json' },
  }),
)

await loginWithPassword('testuser', 'pass12345678')

expect(localStorage.setItem).toHaveBeenCalledWith('gpt-image-playground-token', 'new-token')
```

**Error Testing:**
```typescript
vi.mocked(streamTaskStatus).mockImplementation((_taskId: string, onUpdate: Function) => {
  setTimeout(() => {
    const task = useStore.getState().tasks.find((t: any) => t.id === _taskId)
    if (task) {
      onUpdate({ ...task, status: 'error', error: 'Generation failed', outputImages: [], finishedAt: Date.now(), elapsed: 1000 })
    }
  }, 10)
  return new AbortController()
})

submitTask()

await vi.waitFor(() => {
  const tasks = useStore.getState().tasks
  expect(tasks[0].status).toBe('error')
  expect(tasks[0].error).toBe('Generation failed')
}, { timeout: 5000 })
```

```go
if err := service.ChangePassword(user.ID, "wrong-old", "newpass123"); err == nil {
    t.Fatalf("expected error for wrong old password")
}
```

**Source-Check Testing:**
```typescript
import loginModalSource from './LoginModal.tsx?raw'
import registerModalSource from './RegisterModal.tsx?raw'

describe('auth modal source requirements', () => {
  it('calls password login and bootstraps backend session', () => {
    expect(loginModalSource).toContain('loginWithPassword')
    expect(loginModalSource).toContain('bootstrapBackendSession')
  })
})
```

Use source-check tests only for structural acceptance criteria that are difficult to render in the current test setup. Prefer behavior tests for pure helpers, API clients, store actions, and backend handlers/services.

**Backend HTTP Testing:**
```go
req := httptest.NewRequest(http.MethodPut, "/api/admin/users/user-1/password", strings.NewReader(body))
req.Header.Set("Content-Type", "application/json")
req.Header.Set("Authorization", "Bearer "+token)
resp := httptest.NewRecorder()
r.ServeHTTP(resp, req)

if resp.Code != http.StatusOK {
    t.Fatalf("expected status %d, got %d", http.StatusOK, resp.Code)
}
```

**Backend Database Testing:**
- Use `t.TempDir()` for data/upload directories and SQLite files.
- Override `config.App` and `database.DB` inside setup helpers.
- Run `database.DB.AutoMigrate(...)` for the models required by the test.
- Seed directly with `database.DB.Create(...)` or `database.DB.Save(...)`.
- Close the underlying SQL DB in `t.Cleanup` to avoid locked SQLite files.

---

*Testing analysis: 2026-05-24*
