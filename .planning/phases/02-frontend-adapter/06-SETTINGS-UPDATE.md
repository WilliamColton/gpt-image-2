---
phase: 2
plan: 6
type: execute
wave: 2
depends_on:
  - 5
files_modified:
  - src/types.ts
  - src/lib/backendApi.ts
  - src/store.ts
  - src/components/SettingsModal.tsx
  - src/components/LoginModal.tsx
autonomous: true
requirements:
  - FE-02
  - CFG-03
---

# Plan 06: 更新前端设置和配置逻辑

## Goal

移除前端直接调用 OpenAI 所需的设置（API Key、Base URL），设置 UI 简化。

## 架构说明

后端不配置全局 API Key。每个用户使用自己的 apikey 登录，apikey 即为该用户的 OpenAI API Key（存储为加密的 `ApikeyCipher`）。

## Tasks

### Task 1: 确认 AppSettings 类型

**Read first:** `src/types.ts`（第 1-24 行）

**Action:**
- `AppSettings` 接口中 `apiKey` 字段保持 `string | undefined`（已经是可选的，无需改动）
- 无需新增 `openAIConfigured` 字段（后端不使用全局 API Key）

**Acceptance criteria:**
- `AppSettings` 保持现有结构不变

---

### Task 2: 确认 getPublicConfig 返回类型

**Read first:** `src/lib/backendApi.ts`（第 64-66 行）, `backend-go/handler/config.go`, `backend-go/service/models.go`

**Action:**
- 确认 `getPublicConfig` 返回类型与后端 `AppConfig` 一致
- 无需添加 `openAIConfigured` 字段

**Acceptance criteria:**
- `getPublicConfig` 返回值与后端 `/api/config/public` 响应一致

---

### Task 3: 确认 bootstrapBackendSession 合并配置

**Read first:** `src/store.ts`（第 300-316 行）

**Action:** 确认 `bootstrapBackendSession` 函数：
- 从 `publicConfig` 中读取配置并合并到 settings
- 保留用户的 `apiKey`（如果有）

**Acceptance criteria:**
- `bootstrapBackendSession` 正确合并 publicConfig 到 settings
- 用户本地的 `apiKey` 被保留

---

### Task 4: 确认 LoginModal 逻辑

**Read first:** `src/components/LoginModal.tsx`

**Action:**
- 登录表单保持不变（apikey 是后端登录凭证，同时也是用户的 OpenAI API Key）
- 无需添加 `openAIConfigured` 相关提示

**Acceptance criteria:**
- LoginModal 登录逻辑不变
- 不显示任何 API Key 配置状态提示

---

### Task 5: 确认 SettingsModal 简洁

**Read first:** `src/components/SettingsModal.tsx`

**Action:**
- SettingsModal 当前只显示图片数量和退出按钮（已经很简洁）
- 确认 SettingsModal 不引用 `settings.apiKey`, `settings.baseUrl` 等字段

**Acceptance criteria:**
- SettingsModal 不显示 API Key、Base URL 输入框
- SettingsModal 保持现有的简洁样式（图片数量 + 退出按钮）

## must_haves

- 前端不再强制要求填写 API Key
- 现有登录流程不变（apikey 是后端认证凭证，同时用作 OpenAI API Key）
