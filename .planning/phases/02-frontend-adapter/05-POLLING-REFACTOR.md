---
phase: 2
plan: 5
type: execute
wave: 1
depends_on: []
files_modified:
  - src/lib/api.ts
  - src/lib/backendApi.ts
  - src/store.ts
autonomous: true
requirements:
  - FE-01
  - FE-02
  - FE-05
---

# Plan 05: 改造 api.ts — 提交任务到后端 + 轮询

## Goal

将 `executeTask` 从直接调用 OpenAI API 改为提交任务到后端 + 轮询等待结果。移除前端直接调用 OpenAI 的核心逻辑。

## Tasks

### Task 1: 新增后端提交和轮询 API

**Read first:** `src/lib/backendApi.ts`

**Action:** 在 `backendApi.ts` 新增 3 个函数：

```typescript
/** 提交图片生成任务到后端 */
export async function submitGenerateTask(taskId: string, prompt: string, params: TaskParams, inputImageIds: string[], codexCli: boolean): Promise<{ taskId: string; status: string }> {
  return request('/api/generate', {
    method: 'POST',
    body: JSON.stringify({ taskId, prompt, params, inputImageIds, codexCli }),
  })
}

/** 提交图片编辑任务到后端 */
export async function submitEditTask(taskId: string, prompt: string, params: TaskParams, inputImageIds: string[], maskImageId: string | null, codexCli: boolean): Promise<{ taskId: string; status: string }> {
  return request('/api/edit', {
    method: 'POST',
    body: JSON.stringify({ taskId, prompt, params, inputImageIds, maskImageId, codexCli }),
  })
}

/** 提交 Responses API 生成任务到后端 */
export async function submitResponsesTask(taskId: string, prompt: string, params: TaskParams, inputImageIds: string[], codexCli: boolean): Promise<{ taskId: string; status: string }> {
  return request('/api/responses-generate', {
    method: 'POST',
    body: JSON.stringify({ taskId, prompt, params, inputImageIds, codexCli }),
  })
}
```

**Acceptance criteria:**
- `backendApi.ts` 包含 `submitGenerateTask`, `submitEditTask`, `submitResponsesTask` 三个导出函数
- 三个函数都调用对应的后端端点 (`/api/generate`, `/api/edit`, `/api/responses-generate`)
- 函数参数包含 `taskId`, `prompt`, `params`, `inputImageIds`, `codexCli`（edit 版本额外包含 `maskImageId`）

---

### Task 2: 重写 executeTask — 后端提交 + 轮询

**Read first:** `src/store.ts`（当前 executeTask 函数，第 491-550 行）

**Action:** 将 `executeTask` 函数从直接调用 `callImageApi` 改为：

1. 根据 `settings.apiMode` 和是否有输入图片/遮罩选择提交函数：
   - `settings.apiMode === 'images'` + 有输入图片且有 mask → `submitEditTask`
   - `settings.apiMode === 'images'` + 无 mask → `submitGenerateTask`
   - `settings.apiMode === 'responses'` → `submitResponsesTask`

2. 提交后进入轮询循环：
   ```typescript
   // 轮询间隔：1s，最多 timeout 秒
   const deadline = Date.now() + settings.timeout * 1000
   while (Date.now() < deadline) {
     await new Promise(r => setTimeout(r, 1000))
     const { tasks } = await getTasks()
     const updated = tasks.find(t => t.id === taskId)
     if (!updated) continue
     if (updated.status === 'done') {
       // 处理完成结果
       break
     }
     if (updated.status === 'error') {
       // 处理错误
       break
     }
   }
   // 超时处理
   ```

3. 完成时处理：
   - 将后端返回的 `outputImages`（ID 列表）映射到后端图片 URL 存入 imageCache
   - 使用 `getRemoteImageDataUrl(id)` 构建 URL 并缓存
   - 更新任务的 `actualParams`, `actualParamsByImage`, `revisedPromptByImage`（从后端任务数据获取）
   - 更新任务 status 为 `done`
   - 显示成功 toast

4. 错误/超时处理：
   - 更新任务 status 为 `error`
   - 显示错误信息

**Acceptance criteria:**
- `executeTask` 不再调用 `callImageApi`
- `executeTask` 调用 `submitGenerateTask` / `submitEditTask` / `submitResponsesTask` 之一
- 提交后有轮询循环，每秒检查一次任务状态
- 任务完成时 outputImages 正确存入 imageCache（使用 `getRemoteImageDataUrl`）
- 超时（超过 settings.timeout 秒）更新任务为 error 状态

---

### Task 3: 清理 api.ts 中不再需要的直接调用逻辑

**Read first:** `src/lib/api.ts`

**Action:**
- 删除 `callImageApi`, `callImagesApi`, `callImagesApiSingle`, `callImagesApiConcurrent`, `callResponsesImageApi`, `callResponsesImageApiSingle`, `callResponsesImageApiConcurrent` 函数
- 保留 `normalizeBaseUrl` 导出（其他模块可能使用）
- 保留辅助函数如 `dataUrlToBlob`, `imageDataUrlToPngBlob`, `maskDataUrlToPngBlob` 等（这些用于图片上传）
- 删除不再需要的导入（如 `buildApiUrl`, `readClientDevProxyConfig` 如果不再使用）
- 删除不再需要的常量（如 `MAX_MASK_EDIT_FILE_BYTES`, `MAX_IMAGE_INPUT_PAYLOAD_BYTES` 如果不再需要）
- 最终 `api.ts` 应只包含图片处理工具函数，不含 OpenAI API 调用逻辑

**Acceptance criteria:**
- `api.ts` 不包含 `callImageApi` 函数
- `api.ts` 不包含 `/v1/images/generations`, `/v1/images/edits`, `/v1/responses` 的 fetch 调用
- `api.ts` 仍然导出 `normalizeBaseUrl`
- `api.ts` 中的图片工具函数（blob 转换等）保留

---

### Task 4: 更新 store.ts 中 submitTask 的导入和引用

**Read first:** `src/store.ts`（第 1-31 行导入部分）

**Action:**
- 从 `import { callImageApi } from './lib/api'` 改为导入新的后端提交函数
- 添加 `import { submitGenerateTask, submitEditTask, submitResponsesTask, getTasks as fetchTasks } from './lib/backendApi'`
- 移除对 `callImageApi` 的引用
- 在 `executeTask` 中，轮询时使用 `fetchTasks()` 获取最新任务列表（避免与 store 中的 `getTasks` 命名冲突，导入时重命名为 `fetchTasks`）

**Acceptance criteria:**
- `store.ts` 不再导入 `callImageApi`
- `store.ts` 导入 `submitGenerateTask`, `submitEditTask`, `submitResponsesTask`
- `store.ts` 导入后端的 `getTasks` 函数（重命名为 `fetchTasks` 以避免与 store action 冲突）

## must_haves

- 后端提交替代直接 OpenAI 调用
- 轮询机制（1s 间隔，带超时）
- 输出图片从后端获取并缓存
- 保留 api.ts 中的图片工具函数
