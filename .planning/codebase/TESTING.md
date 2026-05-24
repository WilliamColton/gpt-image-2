# Testing Patterns

**Analysis Date:** 2026-05-24

## Test Framework

### Frontend (TypeScript/React)

**Runner:**
- Vitest 4.1.5
- Config: Inline within `vite.config.ts` (no separate `vitest.config.ts` file)
- Uses Vite's native transform pipeline

**Assertion Library:**
- Vitest's built-in `expect` (Jest-compatible API)
- `@testing-library` is NOT used; DOM assertions are done via `vi.fn` mocks and source-code string checks

**Run Commands:**
```bash
npm test            # Run all tests once (vitest run)
npm run test:watch  # Watch mode (vitest)
```

### Backend (Go)

**Runner:**
- Go standard `testing` package
- Run with: `go test ./...` (standard Go convention)

**Assertion Library:**
- No third-party assertion library; standard `t.Error()`, `t.Fatal()`, `t.Errorf()` with manual comparison

## Test File Organization

**Frontend:**
- Location: Co-located with source files (same directory)
- Naming: `*.test.ts` for logic, `*.test.tsx` for React component imports
- Examples:
  - `src/store.test.ts` next to `src/store.ts`
  - `src/lib/backendApi.test.ts` next to `src/lib/backendApi.ts`
  - `src/admin/adminApi.test.ts` next to `src/admin/adminApi.ts`
  - `src/components/LoginModal.test.tsx` next to `src/components/LoginModal.tsx`

**Backend:**
- Location: Co-located with source files (same package)
- Naming: `*_test.go` suffix
- Examples:
  - `backend-go/config/config_test.go`
  - `backend-go/service/auth_test.go`
  - `backend-go/handler/admin_handler_test.go`

## Existing Test Files and Coverage

### Frontend Tests (13 files)

| Test File | What It Tests | Approach |
|-----------|---------------|----------|
| `src/store.test.ts` | Zustand store: task submission flow, mask lifecycle, image caching, bootstrap, upload dedup | Mocked backendApi + db modules, fake FileReader |
| `src/App-test.test.tsx` | App.tsx conditional rendering (MigrationModal, LoginModal) | Source code string inspection via `?raw` import |
| `src/components/LoginModal.test.tsx` | LoginModal tab switching, RegisterModal creation | Source code string inspection via `?raw` import |
| `src/components/RegisterModal.test.tsx` | RegisterModal form fields, validation, API calls | Source code string inspection via `?raw` import |
| `src/components/MigrationModal.test.tsx` | MigrationModal form elements | Source code string inspection via `?raw` import |
| `src/admin/adminApi.test.ts` | Admin pricing/analytics DTOs and API functions | `vi.fn` mocks on `globalThis.fetch`, `vi.stubGlobal` for localStorage |
| `src/admin/adminApi-invite.test.ts` | Admin invite API functions | Similar pattern to adminApi.test.ts |
| `src/admin/AdminDashboard.test.tsx` | Admin dashboard rendering/behavior | React component testing |
| `src/admin/moneyFormat.test.ts` | Money formatting utilities | Pure function unit tests |
| `src/lib/backendApi.test.ts` | Backend API client functions | Mocked fetch |
| `src/lib/db.test.ts` | IndexedDB abstraction layer (get, put, hash, storeImage) | Fake IDB implementations (FakeIndexedDB, FakeIDBDatabase classes) |
| `src/lib/mask.test.ts` | Mask image ordering utilities | Pure function tests |
| `src/lib/maskPreprocess.test.ts` | Mask preprocessing | Pure function tests |
| `src/lib/viewportTransform.test.ts` | Viewport transform math | Pure function tests |

### Backend Tests (13 files)

| Test File | What It Tests |
|-----------|---------------|
| `backend-go/config/config_test.go` | Config loading and root-dir override |
| `backend-go/database/models_test.go` | Model struct definitions |
| `backend-go/handler/admin_handler_test.go` | AdminResetPassword, AdminGetInviteConfig, AdminUpdateInviteConfig, AdminListInvites |
| `backend-go/handler/admin_analytics_test.go` | Billing analytics handler endpoints |
| `backend-go/handler/admin_pricing_test.go` | Pricing config handler endpoints |
| `backend-go/handler/auth_handler_test.go` | Auth handler endpoints |
| `backend-go/handler/generate_billing_test.go` | Generate handler billing logic |
| `backend-go/handler/images_test.go` | Image upload, download, delete handler | Uses `httptest.NewRecorder`, `multipart/form-data` construction, real temp SQLite DB |
| `backend-go/service/analytics_test.go` | Analytics service methods |
| `backend-go/service/auth_test.go` | Auth service: LoginWithPassword, RegisterUser, MigrateUser, ChangePassword, SetInviteCode, GetInviteCode, ListInvites, AdminResetPassword, password hashing, username/password validation |
| `backend-go/service/billing_test.go` | Billing service methods |
| `backend-go/service/image_test.go` | Image service methods |
| `backend-go/service/models_test.go` | Model service methods |
| `backend-go/service/money_test.go` | ParseMoneyX10000, FormatMoneyX10000 (pure function tests, table-driven) |
| `backend-go/service/openai_failover_test.go` | OpenAI failover with mock endpoints |

## Mock Strategies

### Frontend Mocks

**Module mocking** using `vi.mock()` (see `src/store.test.ts`):
```typescript
vi.mock('./lib/backendApi', () => ({
  submitGenerateTask: vi.fn().mockResolvedValue({ taskId: 'task-1', status: 'processing' }),
  uploadImage: vi.fn().mockResolvedValue({ id: 'uploaded-1', dataUrl: '...', createdAt: 1, source: 'generated' }),
  getBackendToken: vi.fn().mockReturnValue('test-token'),
  getMe: vi.fn(),
  getTasks: vi.fn().mockResolvedValue({ tasks: [] }),
  streamTaskStatus: vi.fn().mockImplementation((_taskId, onUpdate) => {
    setTimeout(() => onUpdate({ ...task, status: 'done', ... }), 10)
    return new AbortController()
  }),
}))

vi.mock('./lib/db', () => ({
  putImage: vi.fn().mockResolvedValue(undefined),
  getImage: vi.fn().mockResolvedValue(null),
  hashDataUrl: vi.fn().mockResolvedValue('hash-1'),
}))
```

**Fetch mocking** (see `src/admin/adminApi.test.ts`):
```typescript
vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
  new Response(JSON.stringify(mockResponse), {
    status: 200,
    headers: { 'Content-Type': 'application/json' },
  }),
)
```

**localStorage stubbing** (see `src/admin/adminApi.test.ts`):
```typescript
vi.stubGlobal('localStorage', {
  getItem: vi.fn((key: string) => store[key] ?? null),
  setItem: vi.fn((key: string, val: string) => { store[key] = val }),
  removeItem: vi.fn((key: string) => { delete store[key] }),
})
```

**Fake Browser APIs** (see `src/lib/db.test.ts`):
```typescript
class FakeIndexedDB {
  private readonly dbs = new Map<string, FakeIDBDatabase>()
  open(name: string) {
    const req = new FakeIDBOpenDBRequest()
    queueMicrotask(() => { /* simulate async IDB */ })
    return req
  }
}
// Then: Object.defineProperty(globalThis, 'indexedDB', { value: new FakeIndexedDB() })
```

**Source code inspection** for UI tests (see `src/components/LoginModal.test.tsx`):
```typescript
import loginModalSource from './LoginModal.tsx?raw'

it('LoginModal imports Tabs, TabsList, TabsTrigger, and TabsContent', () => {
  expect(loginModalSource).toContain("import { Tabs, TabsList, TabsTrigger, TabsContent } from './ui/tabs'")
})
```

**What is mocked:**
- Backend API module (`./lib/backendApi`) -- all HTTP calls
- IndexedDB module (`./lib/db`) -- all storage operations
- `globalThis.fetch` -- for admin API tests
- `localStorage` -- stubbed in tests that use token storage
- `FileReader` -- stubbed with synchronous fake

**What is NOT mocked:**
- Pure utility functions (e.g., mask, viewport, money format) -- tested directly
- State store during task flow tests -- real Zustand store used with mocked dependencies

### Backend Mocks

**Test setup pattern** (see `backend-go/handler/admin_handler_test.go`, `backend-go/service/auth_test.go`):
```go
func setupAdminHandlerTest(t *testing.T) *gin.Engine {
    t.Helper()
    gin.SetMode(gin.TestMode)
    tmp := t.TempDir()
    config.App = &config.Config{
        DataDir:   filepath.Join(tmp, "data"),
        JWTSecret: "test-secret",
        // ... all fields explicitly set
    }
    os.MkdirAll(config.App.DataDir, 0755)
    os.MkdirAll(config.App.UploadDir, 0755)
    // Use real SQLite in temp directory
    db, _ := gorm.Open(sqlite.Open(filepath.Join(config.App.DataDir, "test.sqlite")), &gorm.Config{})
    database.DB = db
    database.DB.AutoMigrate(&database.User{}, &database.RedemptionCode{})
    t.Cleanup(func() {
        sqlDB, _ := database.DB.DB()
        sqlDB.Close()
        config.SetRootDir(originalResolver)  // Restore global state
    })
    // Build Gin router with only the routes under test
    r := gin.New()
    adminAuth := r.Group("/api/admin", middleware.AdminMiddleware())
    adminAuth.PUT("/users/:id/password", AdminResetPassword)
    // ...
    return r
}
```

**Key patterns:**
- **Global config override**: Tests set `config.App` to a test-specific instance then restore via `t.Cleanup`
- **Real SQLite in temp dir**: Each test creates its own SQLite database in `t.TempDir()` for full isolation
- **Gin TestMode**: `gin.SetMode(gin.TestMode)` suppresses debug output
- **httptest**: `httptest.NewRecorder()` + `r.ServeHTTP(resp, req)` for HTTP handler tests
- **JWT signing**: `service.SignToken("user-1", "user", config.App.JWTSecret)` for test tokens
- **Helper functions**: `createTestUserWithPassword()`, `createTestUserWithCode()` create DB fixtures inline
- **No testify**: The project uses standard Go `testing` package exclusively, no `stretchr/testify`
- **Table-driven tests**: Used for utility functions (e.g., `TestFormatMoneyX10000` in `backend-go/service/money_test.go`)

## E2E Testing

- **No E2E testing framework detected** -- no Playwright, Cypress, or Selenium configuration
- No `e2e/` directory, no `playwright.config.*`, no `cypress.config.*`
- E2E-like coverage achieved through Go handler integration tests (full request-response cycle with real SQLite)

## CI/CD Pipeline Testing

- **No CI configuration detected** -- no `.github/workflows/`, `.gitlab-ci.yml`, `Jenkinsfile`, etc.
- Test execution is manual/local only

## Test Commands and Configuration

**Frontend:**
- `npm test` -- Runs `vitest run` (single pass)
- `npm run test:watch` -- Runs `vitest` in watch mode
- No coverage thresholds configured
- No `test` configuration in `package.json` beyond the script definitions
- Test environment: Node (not jsdom), since component tests use source inspection rather than DOM rendering

**Backend:**
- `go test ./...` -- runs all tests in all packages
- `go test ./service/ -v` -- verbose output for specific packages
- `go test -run TestAuth` -- run specific test functions by name pattern

## Test Types

### Unit Tests (Frontend)
- Pure function tests: `mask.test.ts`, `maskPreprocess.test.ts`, `viewportTransform.test.ts`, `moneyFormat.test.ts`
- API client tests with mocked fetch: `adminApi.test.ts`, `adminApi-invite.test.ts`, `backendApi.test.ts`
- IndexedDB abstraction tests with fake IDB: `db.test.ts`
- Store logic tests with mocked dependencies: `store.test.ts`
- Component structure tests via source string matching: `LoginModal.test.tsx`, `RegisterModal.test.tsx`, `MigrationModal.test.tsx`, `App-test.test.tsx`
- Full component rendering tests: `AdminDashboard.test.tsx`

### Integration Tests (Backend)
- Handler tests exercise the full stack: HTTP -> handler -> service -> real SQLite database
- Service tests exercise service -> real SQLite database
- Auth middleware tested end-to-end (valid token -> handler -> 200, no token -> 401/403)

### Snapshot Tests
- **Not used** -- no snapshot test files detected

### E2E Tests
- **Not used** -- no E2E framework detected

## Test Structure Patterns

### Frontend (Vitest)
```typescript
import { beforeEach, describe, expect, it, vi } from 'vitest'

describe('Feature name', () => {
  beforeEach(() => {
    // Reset mocks, set up test state
    useStore.setState({ ...initialState })
  })

  it('does specific thing', async () => {
    // Arrange: configure mocks
    vi.mocked(someFunction).mockResolvedValue(...)

    // Act: call the function
    await functionUnderTest()

    // Assert: check expectations
    expect(someFunction).toHaveBeenCalledWith(...)
    expect(useStore.getState().field).toBe(expected)
  })

  it('handles error case', async () => {
    vi.mocked(someFunction).mockRejectedValue(new Error('fail'))
    // ... verify error handling
  })
})
```

### Backend (Go)
```go
func TestFeatureName(t *testing.T) {
    r := setupHandlerTest(t)
    createTestUser(t, "user-1")
    token := tokenForTestUser(t, "user-1")

    body := `{"field":"value"}`
    req := httptest.NewRequest(http.MethodPost, "/api/endpoint", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+token)
    resp := httptest.NewRecorder()
    r.ServeHTTP(resp, req)

    if resp.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d body=%s", resp.Code, resp.Body.String())
    }
    // Verify DB state
    var row database.Model
    database.DB.Where("id = ?", "user-1").First(&row)
    // ... assertions
}
```

## Common Patterns

### Async Testing (Frontend)
```typescript
// Using vi.waitFor for polling-based state changes
submitTask()
await vi.waitFor(() => {
  const tasks = useStore.getState().tasks
  expect(tasks[0].status).toBe('done')
}, { timeout: 5000 })

// SSE mock with setTimeout
vi.mocked(streamTaskStatus).mockImplementation((_taskId, onUpdate) => {
  setTimeout(() => {
    onUpdate({ ...task, status: 'done', outputImages: ['img-1'] })
  }, 10)
  return new AbortController()
})
```

### Error Testing
```typescript
// Frontend: mock rejection, verify state
vi.mocked(streamTaskStatus).mockImplementation((_taskId, onUpdate) => {
  setTimeout(() => onUpdate({ ...task, status: 'error', error: 'Generation failed' }), 10)
  return new AbortController()
})
await vi.waitFor(() => {
  expect(useStore.getState().tasks[0].status).toBe('error')
  expect(useStore.getState().tasks[0].error).toBe('Generation failed')
})
```

```go
// Backend: verify error status codes and messages
if resp.Code != http.StatusBadRequest {
    t.Fatalf("expected 400, got %d", resp.Code)
}
```

### Store Reset Pattern
```typescript
function resetStoreForTest() {
  useStore.setState({
    authUser: { id: 'user-1', label: 'test', role: 'user', imageCount: 0, quota: 0, usedCount: 0 },
    prompt: '',
    inputImages: [],
    maskDraft: null,
    params: { ...DEFAULT_PARAMS },
    tasks: [],
    // ... all fields
  })
}
```

### Test Fixture Factory
```typescript
function task(overrides: Partial<TaskRecord> = {}): TaskRecord {
  return {
    id: 'task-a',
    prompt: 'prompt',
    params: { ...DEFAULT_PARAMS },
    inputImageIds: [],
    status: 'done',
    error: null,
    createdAt: 1,
    finishedAt: 2,
    elapsed: 1,
    ...overrides,
  }
}
```

## Test Coverage Gaps

### Frontend Gaps

**Components without tests:**
- `src/components/Header.tsx` -- no test file
- `src/components/InputBar.tsx` -- no test file
- `src/components/TaskGrid.tsx` -- no test file
- `src/components/TaskCard.tsx` -- no test file
- `src/components/SearchBar.tsx` -- no test file
- `src/components/DetailModal.tsx` -- no test file
- `src/components/Lightbox.tsx` -- no test file
- `src/components/SettingsModal.tsx` -- no test file
- `src/components/ConfirmDialog.tsx` -- no test file
- `src/components/MaskEditorModal.tsx` -- no test file
- `src/components/SizePickerModal.tsx` -- no test file
- `src/components/AnnouncementModal.tsx` -- no test file
- `src/components/ChangelogModal.tsx` -- no test file
- `src/components/FeedbackModal.tsx` -- no test file
- `src/components/AppearanceModal.tsx` -- no test file
- `src/components/HelpModal.tsx` -- no test file
- `src/components/Select.tsx` -- no test file

**Utility modules without tests:**
- `src/lib/canvasImage.ts` -- no test file
- `src/lib/clipboard.ts` -- no test file
- `src/lib/size.ts` -- no test file
- `src/lib/paramDisplay.tsx` -- no test file
- `src/lib/devProxy.ts` -- no test file
- `src/lib/viewport.ts` -- no test file (viewportTransform has tests, but viewport guards don't)
- `src/hooks/useCloseOnEscape.ts` -- no test file

**Test quality concerns:**
- UI component tests rely on source-code string matching (e.g., `LoginModal.test.tsx`) rather than DOM rendering/interaction -- these tests are fragile (break on any refactor) and don't validate runtime behavior
- No user interaction simulation (no `@testing-library/user-event` or `fireEvent`)
- No accessibility testing
- The `store.test.ts` has a `describe.skip` block for image cache tests labeled "TODO: update for current store implementation" -- indicates stale/abandoned test cases

### Backend Gaps

**Handlers without dedicated tests:**
- `backend-go/handler/announcement.go` -- no `*_test.go` file
- `backend-go/handler/changelog.go` -- no `*_test.go` file
- `backend-go/handler/config.go` -- no `*_test.go` file
- `backend-go/handler/feedback.go` -- no `*_test.go` file
- `backend-go/handler/tasks.go` -- no `*_test.go` file (task list, stream, update, delete, clear endpoints)
- `backend-go/handler/generate.go` -- partial coverage via `generate_billing_test.go` only; no test for the full image generation + edit flow

**Services without dedicated tests:**
- `backend-go/service/announcement.go` -- no test file
- `backend-go/service/changelog.go` -- no test file
- `backend-go/service/feedback.go` -- no test file
- `backend-go/service/task.go` -- no test file
- `backend-go/service/queue.go` -- no test file (concurrency slot acquisition untested)

**Integration gaps:**
- No end-to-end test for the full generate->stream->done workflow
- No concurrent task submission stress tests
- No SSE stream failure/reconnection tests
- No endpoint failover tests at the HTTP handler level (failover is tested at service level only)
- No test for rate limiting or quota enforcement at the handler level
- Middleware tests: `AuthMiddleware`, `AdminMiddleware`, `RequestLogger` -- no dedicated middleware test files

**Other gaps:**
- No coverage reports generated or enforced
- No performance/benchmark tests
- No security-focused tests (SQL injection, XSS, token expiry, etc.)
- No test for graceful shutdown or error recovery paths

---

*Testing analysis: 2026-05-24*
