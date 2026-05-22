# Testing Patterns

**Analysis Date:** 2026-05-22

## Test Framework

**Runner:**
- Frontend: Vitest `^4.1.5` from `package.json`.
- Frontend config: not detected; Vitest uses defaults from Vite/TypeScript because no `vitest.config.*` or `jest.config.*` exists at the repo root.
- Backend: Go standard `testing` package under `backend-go/`, with Gin handler tests using `net/http/httptest`.
- Backend config: `backend-go/go.mod` uses Go `1.23.0` and test dependencies already used by app packages, including Gin, GORM, and SQLite.

**Assertion Library:**
- Frontend: Vitest `expect`, including `toEqual`, `toBe`, `toThrow`, `toBeCloseTo`, `expect.any`, and `expect.objectContaining`, as shown in `src/lib/mask.test.ts`, `src/lib/viewportTransform.test.ts`, `src/lib/db.test.ts`, and `src/store.test.ts`.
- Backend: Go standard assertions through `if` checks plus `t.Fatalf` and `t.Errorf`, as shown in `backend-go/config/config_test.go`, `backend-go/service/image_test.go`, and `backend-go/handler/images_test.go`.

**Run Commands:**
```bash
npm run test                 # Run all frontend Vitest tests once from package.json
npm run test:watch           # Run frontend Vitest in watch mode from package.json
npm run build                # Run TypeScript build plus Vite production build from package.json
(cd backend-go && go test ./...) # Run all backend Go tests
```
- Coverage command: not configured in `package.json`; use `vitest run --coverage` only after adding/configuring a coverage provider.
- No root all-in-one test script runs both frontend and backend; execute frontend and backend commands separately.

## Test File Organization

**Location:**
- Frontend tests are co-located with the module under test in `src/`, such as `src/lib/mask.test.ts` next to `src/lib/mask.ts`, `src/lib/maskPreprocess.test.ts` next to `src/lib/maskPreprocess.ts`, `src/lib/viewportTransform.test.ts` next to `src/lib/viewportTransform.ts`, `src/lib/db.test.ts` next to `src/lib/db.ts`, and `src/store.test.ts` next to `src/store.ts`.
- Backend tests are co-located with the package under test in `backend-go/`, such as `backend-go/config/config_test.go`, `backend-go/service/image_test.go`, and `backend-go/handler/images_test.go`.
- No separate `tests/`, `__tests__/`, Cypress, Playwright, or E2E directory is present.

**Naming:**
- Frontend: use `*.test.ts` for TypeScript unit tests; no `*.test.tsx`, `*.spec.ts`, or `*.spec.tsx` app tests are currently present.
- Backend: use Go `_test.go` files in the same package, with `TestXxx` function names such as `TestGetEndpointPool_MultiEndpoint` in `backend-go/config/config_test.go`, `TestSaveImageBufferReusesSameUserDuplicate` in `backend-go/service/image_test.go`, and `TestImagesGetReturnsImageForOwner` in `backend-go/handler/images_test.go`.
- Test descriptions use behavior-oriented English for frontend `it(...)` cases and descriptive Go function suffixes for backend cases.

**Structure:**
```
src/lib/<module>.ts
src/lib/<module>.test.ts
src/store.ts
src/store.test.ts
backend-go/<package>/<module>.go
backend-go/<package>/<module>_test.go
```

## Test Structure

**Suite Organization:**
```typescript
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { functionUnderTest } from './module'

describe('domain or function name', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
  })

  it('states the expected behavior', async () => {
    const result = await functionUnderTest()
    expect(result).toEqual(expectedValue)
  })
})
```
- The simplest pure helper tests omit `beforeEach` and `vi`, as in `src/lib/mask.test.ts`, `src/lib/maskPreprocess.test.ts`, and `src/lib/viewportTransform.test.ts`.
- Stateful tests group related store behavior under multiple `describe` blocks in `src/store.test.ts`, including `announcement state in store`, `mask draft lifecycle in store actions`, `submitTask backend submission flow`, and `image cache behavior in store`.
- IndexedDB tests define local fake classes above the suite in `src/lib/db.test.ts`, then reset `globalThis.indexedDB` in `beforeEach`.

**Backend suite pattern:**
```go
func TestBehaviorName(t *testing.T) {
	setupPackageTest(t)

	result, err := FunctionUnderTest(input)
	if err != nil {
		t.Fatalf("operation failed: %v", err)
	}
	if result != expected {
		t.Fatalf("expected %v, got %v", expected, result)
	}
}
```
- Use package-local setup helpers with `t.Helper()`, such as `setupImageServiceTest` in `backend-go/service/image_test.go` and `setupImagesHandlerTest` in `backend-go/handler/images_test.go`.
- Use `t.TempDir()` for per-test filesystem isolation in backend tests that touch SQLite or uploads, as shown in `backend-go/service/image_test.go` and `backend-go/handler/images_test.go`.

**Patterns:**
- Setup pattern: reset mocks and global browser APIs in `beforeEach`, as in `src/store.test.ts` lines that reset mocked backend calls, stub `FileReader`, and stub `fetch`.
- Setup pattern: reset Zustand state directly with `useStore.setState(...)` for store tests, using helpers like `resetStoreForTest` and `task` in `src/store.test.ts`.
- Setup pattern: replace browser APIs with fakes when jsdom support is not assumed, such as `FakeIndexedDB` in `src/lib/db.test.ts` and `TestFileReader` in `src/store.test.ts`.
- Teardown pattern: frontend tests rely on `vi.restoreAllMocks()` or mock resets in `beforeEach`; backend tests use `t.Cleanup` to close SQLite handles in `backend-go/service/image_test.go` and `backend-go/handler/images_test.go`.
- Assertion pattern: pure utilities assert exact return objects and arrays with `toEqual`, as in `src/lib/viewportTransform.test.ts` and `src/lib/maskPreprocess.test.ts`.
- Assertion pattern: async flows assert eventual state with `await vi.waitFor(...)`, as in `src/store.test.ts` for submit and cache warming flows.
- Assertion pattern: backend HTTP tests assert status codes, headers, response body snippets, database rows, and filesystem side effects with `httptest.ResponseRecorder`, GORM queries, and `os.Stat`, as in `backend-go/handler/images_test.go`.

## Mocking

**Framework:**
- Frontend: Vitest `vi.mock`, `vi.mocked`, `vi.fn`, `vi.stubGlobal`, and `vi.waitFor`.
- Backend: no mocking framework; tests use real temp SQLite databases, temp directories, Gin test routers, and helper-generated JWTs.

**Patterns:**
```typescript
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

beforeEach(() => {
  vi.mocked(getBackendToken).mockReturnValue('test-token')
  vi.stubGlobal('FileReader', TestFileReader)
  vi.stubGlobal('fetch', vi.fn())
})
```
- Module mocks are declared at the top of `src/store.test.ts` for `src/lib/backendApi.ts` and `src/lib/db.ts`.
- Use `vi.mocked(importedFunction)` to set per-test responses, as shown throughout `src/store.test.ts`.
- Use `vi.stubGlobal` for browser APIs required by code under test, such as `FileReader` and `fetch` in `src/store.test.ts`.
- Use local fake classes instead of partial mocks when the API has event-driven behavior, such as `FakeIDBRequest`, `FakeIDBDatabase`, and `FakeIndexedDB` in `src/lib/db.test.ts`.

**Backend integration setup pattern:**
```go
func setupImagesHandlerTest(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	tmp := t.TempDir()
	config.App = &config.Config{
		DataDir: filepath.Join(tmp, "data"),
		UploadDir: filepath.Join(tmp, "upload"),
		JWTSecret: "test-secret",
	}
	// create directories, open sqlite, migrate, seed rows, register routes
	return r
}
```
- Handler tests should build a fresh Gin engine with only the routes needed for the test, following `setupImagesHandlerTest` in `backend-go/handler/images_test.go`.
- Service tests should configure `config.App`, create a temp SQLite DB, assign `database.DB`, run `AutoMigrate`, and close the DB through `t.Cleanup`, following `setupImageServiceTest` in `backend-go/service/image_test.go`.

**What to Mock:**
- Mock frontend network boundary modules such as `src/lib/backendApi.ts` when testing store behavior in `src/store.test.ts`.
- Mock frontend persistence boundary `src/lib/db.ts` when testing store cache behavior in `src/store.test.ts`.
- Stub browser APIs that are unavailable or need deterministic behavior, including `fetch`, `FileReader`, and `indexedDB`, as shown in `src/store.test.ts` and `src/lib/db.test.ts`.
- In Go handler tests, avoid external services by testing local handlers with `httptest` and temp SQLite; generate valid JWTs through `service.SignToken` as in `backend-go/handler/images_test.go`.

**What NOT to Mock:**
- Do not mock pure helper modules when testing their behavior; test `src/lib/mask.ts`, `src/lib/maskPreprocess.ts`, and `src/lib/viewportTransform.ts` directly with concrete inputs.
- Do not mock GORM or filesystem writes in backend image service tests; `backend-go/service/image_test.go` verifies real SQLite rows and uploaded files in `t.TempDir()`.
- Do not mock Gin routing or auth middleware in handler tests where authorization behavior matters; `backend-go/handler/images_test.go` registers `middleware.AuthMiddleware()` and uses real signed tokens.
- Do not mock Zustand itself; manipulate the real `useStore` state through `useStore.setState(...)` and `useStore.getState()` in `src/store.test.ts`.

## Fixtures and Factories

**Test Data:**
```typescript
function img(id: string): InputImage {
  return { id, dataUrl: `data:image/png;base64,${id}` }
}

function task(overrides: Partial<TaskRecord> = {}): TaskRecord {
  return {
    id: 'task-a',
    prompt: 'prompt',
    params: { ...DEFAULT_PARAMS },
    inputImageIds: [],
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
- Use small local factories near the tests that need them, such as `img` in `src/lib/mask.test.ts`, `image` in `src/lib/maskPreprocess.test.ts`, and `task` in `src/store.test.ts`.
- Use `DEFAULT_PARAMS` and `DEFAULT_SETTINGS` from `src/types.ts` for store and task fixtures in `src/store.test.ts`; do not duplicate default values unless the test explicitly validates them.
- Use raw data URLs like `data:image/png;base64,aaa` for frontend image fixtures in `src/lib/db.test.ts` and `src/store.test.ts`.
- Use byte slices and explicit MIME strings for backend image fixtures, such as `[]byte("png bytes")` and `"image/png"` in `backend-go/handler/images_test.go` and `backend-go/service/image_test.go`.

**Location:**
- Fixtures and factories are currently local to each test file; no shared fixture directory exists.
- Keep a fixture local when it is only used by one module's tests, matching `src/lib/mask.test.ts`, `src/lib/maskPreprocess.test.ts`, and `backend-go/handler/images_test.go`.
- Create shared test helpers only when multiple test files need the same setup; current duplication is limited and explicit.

## Coverage

**Requirements:**
- No coverage threshold is enforced in `package.json`, `vite.config.ts`, or a Vitest config file.
- No Go coverage threshold is enforced under `backend-go/`.
- Current frontend coverage focuses on pure helpers, IndexedDB utilities, and store flows: `src/lib/mask.ts`, `src/lib/maskPreprocess.ts`, `src/lib/viewportTransform.ts`, `src/lib/db.ts`, and `src/store.ts`.
- Current backend coverage focuses on endpoint pool sorting/copy behavior, image storage deduplication, and image HTTP ownership/delete behavior: `backend-go/config/config.go`, `backend-go/service/image.go`, and `backend-go/handler/images.go`.
- React component rendering tests are not present; files such as `src/components/TaskCard.tsx`, `src/components/InputBar.tsx`, `src/components/ConfirmDialog.tsx`, and `src/admin/AdminDashboard.tsx` are tested indirectly, if at all.

**View Coverage:**
```bash
npx vitest run --coverage          # Frontend coverage, requires coverage provider setup if missing
(cd backend-go && go test -cover ./...) # Backend package coverage summary
```

## Test Types

**Unit Tests:**
- Pure TypeScript utility unit tests are the strongest pattern; add new tests next to the utility using `describe('<function or domain>')` and exact object assertions, following `src/lib/mask.test.ts`, `src/lib/maskPreprocess.test.ts`, and `src/lib/viewportTransform.test.ts`.
- TypeScript persistence utility tests use fake browser APIs and async `resolves` assertions, following `src/lib/db.test.ts`.
- Go service unit/integration tests use temp SQLite and temp upload directories to verify real persistence behavior, following `backend-go/service/image_test.go`.

**Integration Tests:**
- Frontend store integration tests mock network and IndexedDB boundaries while exercising real Zustand state and async store actions in `src/store.test.ts`.
- Backend handler integration tests use Gin, real middleware, temp SQLite, multipart requests, and `httptest` in `backend-go/handler/images_test.go`.
- Backend config tests exercise package-level endpoint state directly in `backend-go/config/config_test.go`; reset shared package state with `setEndpoints` inside each test.

**E2E Tests:**
- Not used. No Playwright, Cypress, browser automation config, or E2E test directory is present.
- Manual UI behavior is not captured by automated tests for admin screens, modals, drag/drop, swipe selection, or service worker registration.

## Common Patterns

**Async Testing:**
```typescript
submitTask()

await vi.waitFor(() => {
  const tasks = useStore.getState().tasks
  expect(tasks[0].status).toBe('done')
}, { timeout: 5000 })
```
- Use `vi.waitFor` for store actions that depend on timers, mocked SSE callbacks, or background cache warming, as shown in `src/store.test.ts`.
- Use promise assertions for direct async utilities, such as `await expect(putImage(image)).resolves.toBe('img-1')` in `src/lib/db.test.ts`.
- Use `Promise.all` to validate deduplication/concurrency behavior, as in the concurrent `ensureImageCached('same-img')` test in `src/store.test.ts`.

**Error Testing:**
```typescript
expect(() => orderInputImagesForMask([img('a')], 'missing')).toThrow('遮罩主图已不存在')
expect(() => assertUsableMaskCoverage('empty')).toThrow('请先涂抹需要编辑的区域')
```
- For synchronous validation helpers, assert localized error messages with `toThrow`, following `src/lib/mask.test.ts`.
- For frontend async failure paths, mock rejected/failed boundary responses and assert state or fallback values, such as `ensureImageCached` falling back to a remote URL in `src/store.test.ts`.
- For backend HTTP errors, assert status codes and response bodies from `httptest.ResponseRecorder`, as in `TestImagesGetRejectsOtherUser` and delete/read-after-delete tests in `backend-go/handler/images_test.go`.
- For backend service errors that should fail setup or operations, use `t.Fatalf` immediately to stop the test with context, as in `backend-go/service/image_test.go` and `backend-go/handler/images_test.go`.

---

*Testing analysis: 2026-05-22*
