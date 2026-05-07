---
phase: 02-frontend-adapter
verified: 2026-05-06T09:30:00Z
status: human_needed
score: 6/6 must-haves verified
overrides_applied: 0
re_verification: null
gaps: []
human_verification:
  - test: "Images API 模式：输入提示词，点击生成，观察任务卡片从 running 变为 done，确认图片显示正确，刷新页面后状态不丢失"
    expected: "任务提交后立即显示 running 卡片，后端独立完成 API 调用，完成后图片正确显示，刷新后任务和图片仍在"
    why_human: "需要运行后端 + OpenAI API Key 配置，无法通过代码静态分析验证端到端流程"
  - test: "Responses API 模式：切换到 Responses API 模式，输入提示词生成，确认任务正常完成"
    expected: "Responses API 模式下任务正常提交和完成，输出图片和元数据正确"
    why_human: "需要运行后端 + OpenAI API Key，验证实际 API 调用路由正确"
  - test: "Codex CLI 兼容模式：开启 Codex CLI，设置 n>1 多图生成，确认后端使用并发单图请求"
    expected: "quality 参数为 auto，多图生成使用并发请求，所有输出图片正确显示"
    why_human: "需要运行后端 + Codex CLI API 端点，验证并发行为"
  - test: "编辑模式（带遮罩）：上传参考图，绘制遮罩，输入提示词生成，确认 maskImageId 正确传递"
    expected: "遮罩编辑后任务正确提交到 /api/edit，输出图片为编辑后的结果"
    why_human: "需要运行后端 + OpenAI API Key，验证 multipart 上传和遮罩处理"
  - test: "API Key 可选：后端未配置 OPENAI_API_KEY 时，前端显示提示信息；后端已配置时，前端无需填写 API Key"
    expected: "LoginModal 根据后端配置显示不同提示，前端不再强制要求填写 API Key"
    why_human: "需要验证两种后端配置场景下的 UI 提示"
---

# Phase 2: 前端适配 Verification Report

**Phase Goal:** 前端改为提交任务 + 轮询结果，移除直接调用 OpenAI 的逻辑
**Verified:** 2026-05-06T09:30:00Z
**Status:** human_needed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | `callImageApi` 已移除，`executeTask` 改为提交任务到后端 + 轮询 | ✓ VERIFIED | `src/lib/api.ts` 仅 1 行 re-export `normalizeBaseUrl`；`src/store.ts` 第 485-551 行 `executeTask` 调用 `submitGenerateTask`/`submitEditTask`/`submitResponsesTask` 后进入 1s 轮询循环；grep 确认 src/ 中无 `callImageApi` 引用 |
| 2 | 页面刷新后任务卡片显示正确状态 | ✓ VERIFIED | `bootstrapBackendSession()` (store.ts:302-318) 从后端加载 tasks 并缓存图片；`updateTaskInStore` 调用 `putRemoteTask` 持久化；zustand persist 保存 settings/auth |
| 3 | Images API / Responses API 模式切换正常工作 | ✓ VERIFIED | `executeTask` (store.ts:493-499) 根据 `settings.apiMode` 路由到不同 submit 函数；三种 submit 函数均存在于 `backendApi.ts` (lines 105-126) |
| 4 | Codex CLI 兼容模式行为不变 | ✓ VERIFIED | `codexCli` 参数传递给所有 submit 函数；`quality` 在 codexCli 模式下强制为 `auto` (store.ts:421) |
| 5 | 前端设置中 API Key 变为可选 | ✓ VERIFIED | `openAIConfigured` 字段存在于 `AppSettings` (types.ts:12) 和 `AppConfig` (models.go:24)；后端 `/api/config/public` 暴露此字段 (handler/config.go:19)；`LoginModal` 根据此值显示不同提示 (LoginModal.tsx:51-54) |
| 6 | 前端不再包含直接调用 OpenAI 的逻辑 | ✓ VERIFIED | `api.ts` 仅 re-export `normalizeBaseUrl`；grep 确认 src/ 中无 `callImageApi`/`callImagesApi`/`callResponsesImageApi`/`v1/images/generations`/`v1/images/edits`/`v1/responses` 引用 |

**Score:** 6/6 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `src/lib/backendApi.ts` | 包含 submitGenerateTask, submitEditTask, submitResponsesTask | ✓ VERIFIED | Lines 105-126, 三个函数均存在且调用正确端点 |
| `src/lib/api.ts` | 移除所有 OpenAI 调用逻辑，仅保留工具函数 | ✓ VERIFIED | 仅 1 行 re-export `normalizeBaseUrl` |
| `src/store.ts` | executeTask 使用后端提交 + 轮询 | ✓ VERIFIED | Lines 485-551, 完整的 submit + poll + timeout 逻辑 |
| `src/types.ts` | AppSettings 包含 openAIConfigured | ✓ VERIFIED | Line 12, DEFAULT_SETTINGS line 25 |
| `backend-go/config/config.go` | OpenAIConfigured 从环境变量读取 | ✓ VERIFIED | Line 30 (字段), Line 85 (读取 OPENAI_API_KEY) |
| `backend-go/service/models.go` | AppConfig 包含 OpenAIConfigured | ✓ VERIFIED | Line 24 |
| `backend-go/handler/config.go` | ConfigPublic 暴露 OpenAIConfigured | ✓ VERIFIED | Line 19 |
| `src/components/LoginModal.tsx` | 显示 API Key 可用状态提示 | ✓ VERIFIED | Lines 51-54, 根据 openAIConfigured 显示不同提示 |
| `src/components/SettingsModal.tsx` | 不包含 API Key / Base URL 输入 | ✓ VERIFIED | 仅显示图片数量 + 退出按钮 |
| `src/store.test.ts` | 测试后端提交流程 | ✓ VERIFIED | 6 个测试覆盖 submit/poll/timeout/error/empty/unauthenticated |
| `src/lib/api.test.ts` | 测试 normalizeBaseUrl | ✓ VERIFIED | 2 个测试，无已删除函数的测试 |

### Key Link Verification

| From | To | Via | Status | Details |
|------|-----|-----|--------|---------|
| `store.ts executeTask` | `backendApi submitGenerateTask` | import + call | ✓ WIRED | store.ts:20 imports, store.ts:496 calls |
| `store.ts executeTask` | `backendApi submitEditTask` | import + call | ✓ WIRED | store.ts:21 imports, store.ts:494 calls |
| `store.ts executeTask` | `backendApi submitResponsesTask` | import + call | ✓ WIRED | store.ts:22 imports, store.ts:498 calls |
| `store.ts executeTask` | `backendApi getTasks` (as fetchTasks) | import + polling call | ✓ WIRED | store.ts:18 imports renamed, store.ts:505 calls in loop |
| `store.ts bootstrapBackendSession` | `backendApi getPublicConfig` | import + call | ✓ WIRED | store.ts:17 imports, store.ts:307 calls |
| `LoginModal.tsx` | `store settings.openAIConfigured` | useStore selector | ✓ WIRED | LoginModal.tsx:10 reads from store |
| `backend handler/config.go` | `config.App.OpenAIConfigured` | direct field access | ✓ WIRED | handler/config.go:19 reads from config |
| `backend config.go` | `OPENAI_API_KEY` env var | os.Getenv | ✓ WIRED | config.go:85 reads env var |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|----------|---------------|--------|-------------------|--------|
| `LoginModal.tsx` | `openAIConfigured` | `useStore((s) => s.settings.openAIConfigured)` | Backend config API → publicConfig merge | ✓ FLOWING |
| `store.ts executeTask` | `updated.status` | `fetchTasks()` polling | Backend `/api/tasks` endpoint | ✓ FLOWING |
| `store.ts bootstrapBackendSession` | `publicConfig` | `getPublicConfig()` | Backend `/api/config/public` endpoint | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Tests pass | `npx vitest run` | 32 tests pass across 5 files | ✓ PASS |
| No OpenAI direct calls in src/ | `grep callImageApi src/` | No matches | ✓ PASS |
| api.ts only re-exports normalizeBaseUrl | `cat src/lib/api.ts` | 1 line: `export { normalizeBaseUrl } from './devProxy'` | ✓ PASS |
| backendApi.ts has 3 submit functions | `grep submit.*Task src/lib/backendApi.ts` | 3 matches: submitGenerateTask, submitEditTask, submitResponsesTask | ✓ PASS |
| store.ts imports submit functions | `grep submit.*Task src/store.ts` | 3 imports from backendApi | ✓ PASS |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|------------|-------------|--------|----------|
| FE-01 | Plan 05 | `callImageApi` 改为提交任务到后端，轮询等待结果 | ✓ SATISFIED | `executeTask` 使用 submit + 1s polling (store.ts:485-551) |
| FE-02 | Plan 05, 06 | 移除前端直接向 OpenAI 发送请求和 API Key 的逻辑 | ✓ SATISFIED | api.ts 仅 re-export normalizeBaseUrl；SettingsModal 无 API Key 输入 |
| FE-03 | Plan 07 | 保持 Images API 和 Responses API 两种模式的切换能力 | ✓ SATISFIED | executeTask 根据 apiMode 路由到不同 submit 函数 (store.ts:493-499) |
| FE-04 | Plan 07 | 保持 Codex CLI 兼容模式的行为 | ✓ SATISFIED | codexCli 传递给 submit 函数；quality=auto 强制 (store.ts:421) |
| FE-05 | Plan 05, 07 | 任务卡片显示实时状态，页面刷新后状态不丢失 | ✓ SATISFIED | 轮询更新状态；putRemoteTask 持久化；bootstrapBackendSession 恢复 |
| CFG-03 | Plan 06 | 前端设置中 API Key 改为可选 | ✓ SATISFIED | openAIConfigured 从后端传递；LoginModal 显示提示；SettingsModal 无 API Key |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| (none) | - | - | - | - |

No anti-patterns found. No TODOs, stubs, or empty implementations in modified files.

### Human Verification Required

**5 items need manual testing with running backend + OpenAI API Key:**

1. **Images API 端到端**
   - Test: 启动后端（配置 OPENAI_API_KEY），启动前端，使用 Images API 模式输入提示词生成
   - Expected: 任务从 running 变为 done，图片正确显示，刷新后状态保留
   - Why: 需要实际 OpenAI API 调用验证

2. **Responses API 端到端**
   - Test: 切换到 Responses API 模式，输入提示词生成
   - Expected: 任务正常完成，输出图片和元数据正确
   - Why: 需要验证实际 API 路由

3. **Codex CLI 兼容模式**
   - Test: 开启 Codex CLI，设置 n>1 多图生成
   - Expected: quality=auto，并发请求，所有图片正确显示
   - Why: 需要 Codex CLI API 端点验证并发行为

4. **编辑模式（带遮罩）**
   - Test: 上传参考图，绘制遮罩，输入提示词生成
   - Expected: maskImageId 正确传递，输出为编辑结果
   - Why: 需要验证 multipart 上传和遮罩处理

5. **API Key 可选验证**
   - Test: 后端未配置 OPENAI_API_KEY 时检查 LoginModal 提示；已配置时确认无需填写
   - Expected: 提示信息正确，前端不强制要求 API Key
   - Why: 需要验证两种后端配置场景

### Gaps Summary

No gaps found. All 6 must-haves verified at code level. All artifacts exist, are substantive, and are properly wired. Tests pass (32/32). No anti-patterns detected.

The only remaining verification is runtime behavior (5 human verification items), which requires a running backend with OpenAI API Key configured.

---

_Verified: 2026-05-06T09:30:00Z_
_Verifier: Claude (gsd-verifier)_
